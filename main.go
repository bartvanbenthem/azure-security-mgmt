package main

import (
	"fmt"
	"os"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/bartvanbenthem/azure-update-mgmt/law"
	"github.com/bartvanbenthem/azure-update-mgmt/printer"
)

func main() {

	// create an authorizer from the following environment variables
	// AZURE_CLIENT_ID
	// AZURE_CLIENT_SECRET
	// AZURE_TENANT_ID
	rmAuth, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		panic(err)
	}
	// GET AZURE_SUBSCRIPTION_ID
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")

	// test Virtual Machine information printer
	var print printer.PrintClient
	print.VM(rmAuth, subscriptionID)

	// test LAW query
	lawAuth, err := auth.NewAuthorizerFromEnvironmentWithResource("https://api.loganalytics.io")
	if err != nil {
		panic(err)
	}

	workspace := os.Getenv("AZURE_EXAMPLE_LAW_WORKSPACE")
	var lawclient law.LAWClient
	var q law.KustoQuery
	qresult, err := lawclient.Query(lawAuth, workspace, q.ComputerUpdatesList())
	if err != nil {
		fmt.Println(err)
	}

	result := lawclient.ReturnQueryResultCommaSep(qresult)
	// print comma sepperated results
	fmt.Println(result.Header)
	for _, row := range result.Rows {
		fmt.Printf("%v\n", row)
	}
}
