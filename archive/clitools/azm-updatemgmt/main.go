package main

import (
	"log"
	"os"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/bartvanbenthem/azure-loganalytics/base"
	"github.com/bartvanbenthem/azure-update-mgmt/loganalytics/updatemgmt"
	"github.com/bartvanbenthem/azure-update-mgmt/vm"
)

func main() {
	// create an authorizer from the following environment variables
	// AZURE_CLIENT_ID
	// AZURE_CLIENT_SECRET
	// AZURE_TENANT_ID
	a, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		panic(err)
	}

	lawAuth, err := auth.NewAuthorizerFromEnvironmentWithResource("https://api.loganalytics.io")
	if err != nil {
		panic(err)
	}

	// Create scope from environment variable AZURE_SUBSCRIPTION_ID
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")

	// Get all virtual machines and unique workspaces
	var vmclient vm.RmVMClient
	var workspaces []string

	// list vms
	for _, vm := range vmclient.List(a, subscriptionID) {
		ws := vmclient.GetWorkspaceID(a, vm.Name, vm.ResourceGroup, vm.SubscriptionID)
		workspaces = AppendIfMissing(workspaces, ws)
	}

	var lawgen base.LawClient
	var updmgmt updatemgmt.UpdateMgmtClient
	// run updatemgmt computer list query on law
	for _, ws := range workspaces {
		qresult, err := lawgen.Query(lawAuth, ws, updmgmt.ComputerList())
		if err != nil {
			log.Println(err)
		}
		// print comma sepperated results for csv file
		lawgen.PrintQueryResultCommaSep(qresult)
	}
}

func AppendIfMissing(slice []string, element string) []string {
	for _, existing := range slice {
		if existing == element {
			return slice
		}
	}
	return append(slice, element)
}
