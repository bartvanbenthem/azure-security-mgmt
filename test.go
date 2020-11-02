package main

import (
	"fmt"
	"os"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/bartvanbenthem/azure-update-mgmt/vm"
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

	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")

	var vmclient vm.RmVMClient
	for _, vm := range vmclient.List(auth, subscriptionID) {
		workspace := vmclient.GetWorkspaceID(auth, vm.Name, vm.ResourceGroup, vm.SubscriptionID)
		managedby := vmclient.GetManagedByTag(auth, vm.Name, vm.ResourceGroup, vm.SubscriptionID)
		fmt.Printf("%-20v %-40v %v\n", vm.Name, workspace, managedby)
	}
}
