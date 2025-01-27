package rest

import (
	"log"

	"github.com/cosmos/cosmos-sdk/client"
	clientRest "github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/gorilla/mux"
)

func RegisterRoutes(clientCtx client.Context, router *mux.Router) {
	r := clientRest.WithHTTPDeprecationHeaders(router)

	log.Println(r) // TODO: Delete

	// Register Query routes

	// Register Tx routes
}
