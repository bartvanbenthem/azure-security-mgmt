package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/Azure/go-autorest/autorest/azure/auth"
)

func StringToPointer(s string) *string { return &s }

func main() {
	// create an authorizer from the following environment variables
	// AZURE_CLIENT_ID
	// AZURE_CLIENT_SECRET
	// AZURE_TENANT_ID
	a, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		panic(err)
	}

	al, err := auth.NewAuthorizerFromEnvironmentWithResource("https://api.loganalytics.io")
	if err != nil {
		panic(err)
	}

	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")

	// list all virtual machines within the specified subscription
	var vm VMInfo
	vms := vm.List(a, subscriptionID)
	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	fmt.Fprintln(writer, "Name\tManagedBy\tAssessed\tCriticalUpdates\tOtherUpdates\tCompliant\t")
	fmt.Fprintln(writer, "----\t---------\t--------\t---------------\t------------\t---------\t")

	for _, v := range vms {
		azvm := vm.Get(a, v.Name, v.ResourceGroup, v.SubscriptionID)
		if azvm.OperatingSystem == "Windows" && len(azvm.WorkspaceID) != 0 {
			var w Windows
			var compliancy bool
			assessed := w.Assessed(al, azvm.VMID, azvm.WorkspaceID)
			critsec := w.UpdatesCriticalSecurityCount(al, azvm.VMID, azvm.WorkspaceID)
			updates := w.UpdatesCount(al, azvm.VMID, azvm.WorkspaceID)
			if assessed == true && critsec == 0 {
				compliancy = true
			} else {
				compliancy = false
			}
			fmt.Fprintln(writer, fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t",
				azvm.Name, azvm.ManagedBy, assessed,
				critsec, updates, compliancy))
		}
	}
	writer.Flush()
	fmt.Println()
}
