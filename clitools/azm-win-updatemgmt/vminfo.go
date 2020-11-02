package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest"
)

type VMInfo struct {
	ID              string // full path of Resource
	VMID            string // VMID
	Name            string // VM name
	ResourceGroup   string // resource group VM
	SubscriptionID  string // subscription id VM
	WorkspaceID     string // log analytics workspace
	OperatingSystem string // osType Windows or Linux
	ManagedBy       string // check if server is managed from tag
}

func (v *VMInfo) Get(a autorest.Authorizer, vmName, resourceGroup, subscriptionID string) VMInfo {
	// create virtualmachine type object
	var vmachine VMInfo

	// import environment variables
	managedkey := os.Getenv("AZURE_MANAGED_BY_TAGGING_KEY")
	managedval := os.Getenv("AZURE_MANAGED_BY_TAGGING_VALUE")
	client := os.Getenv("AZURE_CLIENT_NAME")

	// query virtual machines
	computeClient := compute.NewVirtualMachinesClient(subscriptionID)
	computeClient.Authorizer = a
	vm, err := computeClient.Get(context.Background(), resourceGroup, vmName, "instanceView")
	if err != nil {
		panic(err)
	}

	vmachine.ID = *vm.ID
	slid := strings.Split(vmachine.ID, "/")
	vmachine.Name = *vm.Name
	vmachine.ResourceGroup = slid[4]
	vmachine.SubscriptionID = slid[2]
	vmachine.OperatingSystem = string(vm.StorageProfile.OsDisk.OsType)
	vmachine.VMID = *vm.VMID

	// get the workspace ID with the extension client
	for _, res := range *vm.Resources {
		id := strings.Split(*res.ID, "/")
		resourcetype := id[10]
		if resourcetype == "MicrosoftMonitoringAgent" ||
			resourcetype == "Microsoft.EnterpriseCloud.Monitoring" ||
			resourcetype == "OmsAgentForLinux" ||
			resourcetype == "OMSExtension" {
			extClient := compute.NewVirtualMachineExtensionsClient(subscriptionID)
			extClient.Authorizer = a
			ext, err := extClient.Get(context.Background(), resourceGroup, vmName, *res.Name, "")
			if err != nil {
				log.Println(err)
			}

			ws := make(map[string]string)
			settings, err := json.Marshal(ext.Settings)
			if err != nil {
				log.Println(err)
			}

			str := string(settings)
			if str != "" {
				json.Unmarshal([]byte(str), &ws)
				vmachine.WorkspaceID = ws["workspaceId"]
			}
		}

		// get the ManagedBy Tag
		tags := make(map[string]string)
		managed, err := json.Marshal(vm.Tags)
		if err != nil {
			log.Println(err)
		}

		str := string(managed)
		json.Unmarshal([]byte(str), &tags)
		if tags[managedkey] == managedval {
			vmachine.ManagedBy = tags[managedkey]
		} else {
			vmachine.ManagedBy = client
		}
	}

	return vmachine
}

func (v *VMInfo) List(a autorest.Authorizer, subscriptionID string) []VMInfo {
	// create virtualmachine type objects
	var vmachine VMInfo
	var vmachines []VMInfo

	computeClient := compute.NewVirtualMachinesClient(subscriptionID)
	computeClient.Authorizer = a
	vms, err := computeClient.ListAll(context.Background(), "true")
	if err != nil {
		panic(err)
	}

	// list all virtual machines with corresponding resource group
	for _, vm := range vms.Values() {
		vmachine.ID = *vm.ID
		slid := strings.Split(vmachine.ID, "/")
		vmachine.ResourceGroup = slid[4]
		vmachine.SubscriptionID = slid[2]
		vmachine.Name = *vm.Name
		vmachine.VMID = *vm.VMID
		vmachines = append(vmachines, vmachine)
	}
	return vmachines
}
