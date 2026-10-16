package v2

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/keeper"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v11/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"
	channeltypesv2 "github.com/cosmos/ibc-go/v11/modules/core/04-channel/v2/types"
	"github.com/cosmos/ibc-go/v11/modules/core/api"
)

var (
	_ api.IBCModule                   = (*IBCMiddleware)(nil)
	_ api.PacketDataUnmarshaler       = (*IBCMiddleware)(nil)
	_ api.WriteAcknowledgementWrapper = (*IBCMiddleware)(nil)
)

type IBCMiddleware struct {
	app             api.IBCModule
	keeper          keeper.Keeper
	writeAckWrapper api.WriteAcknowledgementWrapper
	chanKeeperV2    ratelimittypes.ChannelKeeperV2
}

// NewIBCMiddleware creates a new IBCMiddleware instance.
//
// Deprecated: use NewIBCMiddlewareWithAsyncAcknowledgements when the middleware
// may be used in the IBC v2 async acknowledgement path.
func NewIBCMiddleware(k keeper.Keeper, app api.IBCModule) IBCMiddleware {
	return IBCMiddleware{
		app:    app,
		keeper: k,
	}
}

// NewIBCMiddlewareWithAsyncAcknowledgements creates a new IBCMiddleware instance
// with the dependencies required to process IBC v2 async acknowledgements.
// It panics if writeAckWrapper or chanKeeperV2 is nil.
func NewIBCMiddlewareWithAsyncAcknowledgements(
	k keeper.Keeper,
	app api.IBCModule,
	writeAckWrapper api.WriteAcknowledgementWrapper,
	chanKeeperV2 ratelimittypes.ChannelKeeperV2,
) IBCMiddleware {
	im := NewIBCMiddleware(k, app)
	im.SetWriteAcknowledgementWrapper(writeAckWrapper)
	im.SetChannelKeeperV2(chanKeeperV2)
	return im
}

// SetWriteAcknowledgementWrapper sets the underlying IBC v2 write acknowledgement wrapper used for async acknowledgements.
// It panics if writeAckWrapper is nil.
func (im *IBCMiddleware) SetWriteAcknowledgementWrapper(writeAckWrapper api.WriteAcknowledgementWrapper) {
	if writeAckWrapper == nil {
		panic(ratelimittypes.ErrWriteAcknowledgementWrapperNil)
	}

	im.writeAckWrapper = writeAckWrapper
}

// SetChannelKeeperV2 sets the IBC v2 channel keeper used to retrieve async packets before writing acknowledgements.
// It panics if chanKeeperV2 is nil.
func (im *IBCMiddleware) SetChannelKeeperV2(chanKeeperV2 ratelimittypes.ChannelKeeperV2) {
	if chanKeeperV2 == nil {
		panic(ratelimittypes.ErrChannelKeeperV2Nil)
	}

	im.chanKeeperV2 = chanKeeperV2
}

// UnmarshalPacketData delegates packet data unmarshaling to the underlying application.
func (im IBCMiddleware) UnmarshalPacketData(payload channeltypesv2.Payload) (any, error) {
	packetDataUnmarshaler, ok := im.app.(api.PacketDataUnmarshaler)
	if !ok {
		return nil, errors.New("underlying application does not implement packet data unmarshaler")
	}

	return packetDataUnmarshaler.UnmarshalPacketData(payload)
}

func (im IBCMiddleware) OnSendPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload channeltypesv2.Payload,
	signer sdk.AccAddress,
) error {
	packet, err := v2ToV1Packet(payload, sourceClient, destinationClient, sequence)
	if err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 rate limiting OnSendPacket failed to convert v2 packet to v1 packet: %s", err.Error()))
		return err
	}
	if err := im.keeper.SendRateLimitedPacket(ctx, packet); err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 packet send was denied: %s", err.Error()))
		return err
	}
	return im.app.OnSendPacket(ctx, sourceClient, destinationClient, sequence, payload, signer)
}

func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload channeltypesv2.Payload,
	relayer sdk.AccAddress,
) channeltypesv2.RecvPacketResult {
	packet, err := v2ToV1Packet(payload, sourceClient, destinationClient, sequence)
	if err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 rate limiting OnRecvPacket failed to convert v2 packet to v1 packet: %s", err.Error()))
		return channeltypesv2.RecvPacketResult{
			Status:          channeltypesv2.PacketStatus_Failure,
			Acknowledgement: channeltypes.NewErrorAcknowledgement(err).Acknowledgement(),
		}
	}
	// Check if the packet would cause the rate limit to be exceeded,
	// and if so, return an ack error
	if err := im.keeper.ReceiveRateLimitedPacket(ctx, packet); err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 packet receive was denied: %s", err.Error()))
		return channeltypesv2.RecvPacketResult{
			Status:          channeltypesv2.PacketStatus_Failure,
			Acknowledgement: channeltypes.NewErrorAcknowledgement(err).Acknowledgement(),
		}
	}

	result := im.app.OnRecvPacket(ctx, sourceClient, destinationClient, sequence, payload, relayer)
	if result.Status != channeltypesv2.PacketStatus_Async {
		packetInfo, err := keeper.ParsePacketInfo(packet, ratelimittypes.PACKET_RECV)
		if err != nil {
			im.keeper.Logger(ctx).Error("ICS20 rate limiting OnRecvPacket failed to parse packet data for pending receive cleanup", "error", err)
			return result
		}
		if err := im.keeper.RemovePendingReceivePacket(ctx, packetInfo.ChannelID, sequence, packetInfo.Denom); err != nil {
			im.keeper.Logger(ctx).Error("ICS20 rate limiting OnRecvPacket failed to remove pending receive packet", "error", err)
		}
	}

	return result
}

func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload channeltypesv2.Payload,
	relayer sdk.AccAddress,
) error {
	packet, err := v2ToV1Packet(payload, sourceClient, destinationClient, sequence)
	if err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 rate limiting OnTimeoutPacket failed to convert v2 packet to v1 packet: %s", err.Error()))
		return err
	}
	if err := im.keeper.TimeoutRateLimitedPacket(ctx, packet); err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 RateLimited OnTimeoutPacket failed: %s", err.Error()))
		return err
	}
	return im.app.OnTimeoutPacket(ctx, sourceClient, destinationClient, sequence, payload, relayer)
}

func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	acknowledgement []byte,
	payload channeltypesv2.Payload,
	relayer sdk.AccAddress,
) error {
	packet, err := v2ToV1Packet(payload, sourceClient, destinationClient, sequence)
	if err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 rate limiting OnAckPacketfailed to convert v2 packet to v1 packet: %s", err.Error()))
		return err
	}
	if err := im.keeper.AcknowledgeRateLimitedPacket(ctx, packet, acknowledgement); err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 RateLimited OnAckPacket failed: %s", err.Error()))
		return err
	}
	return im.app.OnAcknowledgementPacket(ctx, sourceClient, destinationClient, sequence, acknowledgement, payload, relayer)
}

func (im IBCMiddleware) WriteAcknowledgement(ctx sdk.Context, clientID string, sequence uint64, ack channeltypesv2.Acknowledgement) error {
	if im.chanKeeperV2 == nil {
		return ratelimittypes.ErrChannelKeeperV2Nil
	}
	if im.writeAckWrapper == nil {
		return ratelimittypes.ErrWriteAcknowledgementWrapperNil
	}

	packet, found := im.chanKeeperV2.GetAsyncPacket(ctx, clientID, sequence)
	if !found {
		im.keeper.Logger(ctx).Error("ICS20 rate limiting WriteAcknowledgement failed: async packet not found", "clientID", clientID, "sequence", sequence)
		return ratelimittypes.ErrAsyncPacketNotFound.Wrapf("clientID: %s, sequence: %d", clientID, sequence)
	}

	// Async acknowledgements can only be for single payload packets.
	if len(ack.AppAcknowledgements) != 1 || len(packet.Payloads) != 1 {
		im.keeper.Logger(ctx).Error("ICS20 rate limiting WriteAcknowledgement failed: async acknowledgements can only be for single payload packets", "clientID", clientID, "sequence", sequence)
		return im.writeAckWrapper.WriteAcknowledgement(ctx, clientID, sequence, ack)
	}

	v1Packet, err := v2ToV1Packet(packet.Payloads[0], packet.SourceClient, packet.DestinationClient, packet.Sequence)
	if err != nil {
		im.keeper.Logger(ctx).Error("ICS20 rate limiting WriteAcknowledgement failed to convert v2 packet to v1 packet", "error", err)
		return err
	}

	if ack.Success() {
		packetInfo, err := keeper.ParsePacketInfo(v1Packet, ratelimittypes.PACKET_RECV)
		if err != nil {
			im.keeper.Logger(ctx).Error("ICS20 rate limiting WriteAcknowledgement failed to parse packet data for pending receive cleanup", "error", err)
			return im.writeAckWrapper.WriteAcknowledgement(ctx, clientID, sequence, ack)
		}
		if err := im.keeper.RemovePendingReceivePacket(ctx, packetInfo.ChannelID, sequence, packetInfo.Denom); err != nil {
			im.keeper.Logger(ctx).Error("ICS20 rate limiting WriteAcknowledgement failed to remove pending receive packet", "error", err)
			return err
		}
	} else {
		if err := im.keeper.UndoReceivePacket(ctx, v1Packet); err != nil {
			im.keeper.Logger(ctx).Error("ICS20 rate limiting WriteAcknowledgement failed to undo receive packet", "error", err)
			return err
		}
	}

	return im.writeAckWrapper.WriteAcknowledgement(ctx, clientID, sequence, ack)
}

func v2ToV1Packet(payload channeltypesv2.Payload, sourceClient, destinationClient string, sequence uint64) (channeltypes.Packet, error) {
	transferRepresentation, err := transfertypes.UnmarshalPacketData(payload.Value, payload.Version, payload.Encoding)
	if err != nil {
		return channeltypes.Packet{}, err
	}

	packetData := transfertypes.FungibleTokenPacketData{
		Denom:    transferRepresentation.Token.Denom.Path(),
		Amount:   transferRepresentation.Token.Amount,
		Sender:   transferRepresentation.Sender,
		Receiver: transferRepresentation.Receiver,
		Memo:     transferRepresentation.Memo,
	}

	packetDataBz, err := json.Marshal(packetData)
	if err != nil {
		return channeltypes.Packet{}, err
	}

	return channeltypes.Packet{
		Sequence:           sequence,
		SourcePort:         payload.SourcePort,
		SourceChannel:      sourceClient,
		DestinationPort:    payload.DestinationPort,
		DestinationChannel: destinationClient,
		Data:               packetDataBz,
		TimeoutHeight:      clienttypes.Height{},
		TimeoutTimestamp:   0,
	}, nil
}
