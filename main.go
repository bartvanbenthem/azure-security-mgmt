package main

import (
	"os"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/bartvanbenthem/azure-update-mgmt/printer"
)

func main() {
	// create an authorizer from the following environment variables
	// AZURE_CLIENT_ID
	// AZURE_CLIENT_SECRET
	// AZURE_TENANT_ID
	auth, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		panic(err)
	}
	// GET AZURE_SUBSCRIPTION_ID
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")

	// test Virtual Machine information printer
	var print printer.PrintClient
	print.VM(auth, subscriptionID)
}
