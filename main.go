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
	var computerlist law.ComputerListQueryResult
	fmt.Printf("%-50v %-10v %-10v %-10v\n", "name", "security", "critical", "compliance")
	fmt.Printf("%-50v %-10v %-10v %-10v\n", "----", "--------", "--------", "----------")
	for _, w := range uworkspaces {
		result := computerlist.ReturnObject(lawAuth, w)
		for _, r := range result {
			fmt.Printf("%-50v %-10v %-10v %-10v\n", r.DisplayName, r.MissingSecurityUpdatesCount, r.MissingCriticalUpdatesCount, r.Compliance)
		}
	}

	/*
		// TEST GENERIC LAW RETURN FUNCTIONS
		var law law.LAWClient
		for _, w := range uworkspaces {
			fmt.Println(law.ReturnColumnSlice(lawAuth, w))
		}

		for _, w := range uworkspaces {
			result := law.ReturnRowSlice(lawAuth, w)
			for _, r := range result {
				fmt.Println(r)
			}
		}
	*/

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
