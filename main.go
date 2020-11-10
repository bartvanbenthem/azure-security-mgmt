package main

import (
	"fmt"
	"os"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/bartvanbenthem/azure-update-mgmt/law"
	"github.com/bartvanbenthem/azure-update-mgmt/vm"
)

func main() {
	// GET authorizations
	rmAuth := resourceManagerAuthorizer()
	lawAuth := loganalyticsAuthorizer()

	// GET AZURE_SUBSCRIPTION_ID
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")

	// Get virtual machines and workspaces
	var workspaces []string
	var vmclient vm.RmVMClient
	for _, vm := range vmclient.List(rmAuth, subscriptionID) {
		workspace := vmclient.GetWorkspaceID(rmAuth, vm.Name, vm.ResourceGroup, vm.SubscriptionID)
		managedby := vmclient.GetManagedByTag(rmAuth, vm.Name, vm.ResourceGroup, vm.SubscriptionID)
		if managedby == os.Getenv("AZURE_MANAGED_BY_TAGGING_VALUE") {
			workspaces = append(workspaces, workspace)
		}
	}
	// Get unique values from the string slice of workspace ID`s
	uworkspaces := UniqueString(workspaces)

	// computerlist query result LAW
	var computerlist law.ComputerUpdatesList
	for _, w := range uworkspaces {
		result := computerlist.ReturnObject(lawAuth, w)
		resultHR := computerlist.ConvertToReadableObject(result)
		for _, r := range resultHR.Rows {
			fmt.Printf("%v,%v,%v,%v,%v,%v\n",
				r.DisplayName, r.OSType, r.MissingSecurityUpdatesCount,
				r.MissingCriticalUpdatesCount, r.Compliance, r.LastAssessedTime)
		}
	}
}

func resourceManagerAuthorizer() autorest.Authorizer {
	var rmAuth autorest.Authorizer
	var err error
	if len(os.Getenv("AZURE_CLIENT_ID")) == 0 || len(os.Getenv("AZURE_CLIENT_SECRET")) == 0 {
		// create an resource manager authorizer from the az cli configuration
		rmAuth, err = auth.NewAuthorizerFromCLI()
		if err != nil {
			panic(err)
		}
	} else {
		// create an resource manager authorizer from the following environment variables
		// AZURE_CLIENT_ID  | AZURE_CLIENT_SECRET | AZURE_TENANT_ID
		rmAuth, err = auth.NewAuthorizerFromEnvironment()
		if err != nil {
			panic(err)
		}
	}
	return rmAuth
}

func loganalyticsAuthorizer() autorest.Authorizer {
	var lawAuth autorest.Authorizer
	var err error
	if len(os.Getenv("AZURE_CLIENT_ID")) == 0 || len(os.Getenv("AZURE_CLIENT_SECRET")) == 0 {
		// create an LAW authorizer from the az cli configuration
		lawAuth, err = auth.NewAuthorizerFromCLIWithResource("https://api.loganalytics.io")
		if err != nil {
			panic(err)
		}
	} else {
		// create an LAW authorizer from the following environment variables
		// AZURE_CLIENT_ID  | AZURE_CLIENT_SECRET | AZURE_TENANT_ID
		lawAuth, err = auth.NewAuthorizerFromEnvironmentWithResource("https://api.loganalytics.io")
		if err != nil {
			panic(err)
		}
	}
	return lawAuth
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

func vmPrint(auth autorest.Authorizer, subscriptionID string) {
	var vmclient vm.RmVMClient
	fmt.Printf("%-20v %-40v %-10v %-40v %v\n", "Name", "workspaceID", "ostype", "UUID", "managedby")
	fmt.Printf("%-20v %-40v %-10v %-40v %v\n", "----", "-----------", "------", "----", "---------")
	for _, vm := range vmclient.List(auth, subscriptionID) {
		workspace := vmclient.GetWorkspaceID(auth, vm.Name, vm.ResourceGroup, vm.SubscriptionID)
		managedby := vmclient.GetManagedByTag(auth, vm.Name, vm.ResourceGroup, vm.SubscriptionID)
		ostype := vmclient.GetOSType(auth, vm.Name, vm.ResourceGroup, vm.SubscriptionID)
		vmid := vmclient.GetVMID(auth, vm.Name, vm.ResourceGroup, vm.SubscriptionID)
		fmt.Printf("%-20v %-40v %-10v %-40v %v\n", vm.Name, workspace, ostype, vmid, managedby)
	}
}
