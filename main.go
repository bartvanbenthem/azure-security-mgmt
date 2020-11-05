package main

import (
	"os"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/bartvanbenthem/azure-security-mgmt/printer"
	"github.com/bartvanbenthem/azure-security-mgmt/vm"
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

	// LAW AUTH
	lawAuth, err := auth.NewAuthorizerFromEnvironmentWithResource("https://api.loganalytics.io")
	if err != nil {
		panic(err)
	}

	// GET AZURE_SUBSCRIPTION_ID
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")

	// test Virtual Machine information printer
	//var print printer.PrintClient
	//print.VMFormatted(rmAuth, subscriptionID)

	var workspaces []string
	var vmclient vm.RmVMClient
	for _, vm := range vmclient.List(rmAuth, subscriptionID) {
		workspace := vmclient.GetWorkspaceID(rmAuth, vm.Name, vm.ResourceGroup, vm.SubscriptionID)
		workspaces = append(workspaces, workspace)
	}

	// Get unique values from the string slice of workspace ID`s
	uworkspaces := UniqueString(workspaces)

	var print printer.PrintClient
	for _, w := range uworkspaces {
		print.LAWTableRowsCommaSep(lawAuth, w)
	}

}

func UniqueString(s []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
