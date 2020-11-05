package main

import (
	"fmt"
	"os"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/bartvanbenthem/azure-security-mgmt/law"
)

func main() {
	/*
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
	*/

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

	bresult, err := lawclient.ResultParserByte(qresult)
	if err != nil {
		fmt.Println(err)
	}

	res := lawclient.ResultParserLAWQueryResult(bresult)

	for _, table := range res.Tables {
		for _, col := range table.Columns {
			fmt.Printf("%v,", col.Name)
		}
		fmt.Printf("\n")
	}

	for _, table := range res.Tables {
		for _, row := range table.Rows {
			for _, item := range row {
				fmt.Printf("%v,", item)
			}
			fmt.Printf("\n")
		}
	}
}
