package main

import (
	"fmt"
	"os"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/bartvanbenthem/azure-mgmt/loganalytics/raw"
	"github.com/bartvanbenthem/azure-mgmt/loganalytics/updatemgmt"
)

func main() {
	// create an authorizer from the following environment variables
	// AZURE_CLIENT_ID
	// AZURE_CLIENT_SECRET
	// AZURE_TENANT_ID
	lawAuth, err := auth.NewAuthorizerFromEnvironmentWithResource("https://api.loganalytics.io")
	if err != nil {
		panic(err)
	}

	//AZURE_EXAMPLE_LAW_WORKSPACE
	workspace := os.Getenv("AZURE_EXAMPLE_LAW_WORKSPACE")

	var lawgen raw.RawClient
	var updmgmt updatemgmt.UpdateMgmtClient
	// run updatemgmt computer list query on law
	qresult, err := lawgen.Query(lawAuth, workspace, updmgmt.ComputerList())
	// return comma sepperated results for csv file
	result := lawgen.ReturnQueryResultCommaSep(qresult)

	// print comma sepperated results
	fmt.Println(result.Header)
	for _, row := range result.Rows {
		fmt.Printf("%v\n", row)
	}

}
