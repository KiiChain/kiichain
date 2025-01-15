package evmrpc_test

import (
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestAssocation(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	// Sign empty payload prepended with Ethereum Signed Message
	customMessageHash := crypto.Keccak256Hash([]byte("\x19Ethereum Signed Message:\n0"))
	signature, err := crypto.Sign(customMessageHash[:], privateKey)
	if err != nil {
		log.Fatalf("Failed to sign payload: %v", err)
	}

	txArgs := map[string]interface{}{
		"r":              fmt.Sprintf("0x%v", new(big.Int).SetBytes(signature[:32]).Text(16)),
		"s":              fmt.Sprintf("0x%v", new(big.Int).SetBytes(signature[32:64]).Text(16)),
		"v":              fmt.Sprintf("0x%v", new(big.Int).SetBytes([]byte{signature[64]}).Text(16)),
		"custom_message": "\x19Ethereum Signed Message:\n0",
	}

	body := sendRequestGoodWithNamespace(t, "kii", "associate", txArgs)
	require.Equal(t, nil, body["result"])
}

func TestGetKiiAddress(t *testing.T) {
	body := sendRequestGoodWithNamespace(t, "kii", "getKiiAddress", "0x1df809C639027b465B931BD63Ce71c8E5834D9d6")
	require.Equal(t, "kii1mf0llhmqane5w2y8uynmghmk2w4mh0xltzm959", body["result"])
}

func TestGetEvmAddress(t *testing.T) {
	body := sendRequestGoodWithNamespace(t, "kii", "getEVMAddress", "kii1mf0llhmqane5w2y8uynmghmk2w4mh0xltzm959")
	require.Equal(t, "0x1df809C639027b465B931BD63Ce71c8E5834D9d6", body["result"])
}

func TestGetCosmosTx(t *testing.T) {
	body := sendRequestGoodWithNamespace(t, "kii", "getCosmosTx", "0xc1f0d26c419dea496540ab96a3331a9a79f084d7bc9662178dcd7c0bc407dc33")
	fmt.Println(body)
	require.Equal(t, "B5378F3256A6F7E8FBB5BD2E972517634C9B7142F0368970C356E2F1150D4B05", body["result"])
}

func TestGetEvmTx(t *testing.T) {
	body := sendRequestGoodWithNamespace(t, "kii", "getEvmTx", "690D39ADF56D4C811B766DFCD729A415C36C4BFFE80D63E305373B9518EBFB14")
	fmt.Println(body)
	require.Equal(t, "0xc1f0d26c419dea496540ab96a3331a9a79f084d7bc9662178dcd7c0bc407dc33", body["result"])
}
