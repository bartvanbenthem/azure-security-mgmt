package vm

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest"
)

type RmVMClient struct{}

type VMListResult struct {
	ID             string
	Name           string
	ResourceGroup  string
	SubscriptionID string
	VMID           string
}

func (c *RmVMClient) GetOSType() {
	//OSType = string(vm.StorageProfile.OsDisk.OsType)
}

func (c *RmVMClient) List(a autorest.Authorizer, subscriptionID string) []VMListResult {
	var virtualmachine VMListResult
	var virtualmachines []VMListResult

	computeClient := compute.NewVirtualMachinesClient(subscriptionID)
	computeClient.Authorizer = a
	vms, err := computeClient.ListAll(context.Background(), "true")
	if err != nil {
		panic(err)
	}

	// list all virtual machines with corresponding resource group
	for _, vm := range vms.Values() {
		virtualmachine.ID = *vm.ID
		slid := strings.Split(virtualmachine.ID, "/")
		virtualmachine.ResourceGroup = slid[4]
		virtualmachine.SubscriptionID = slid[2]
		virtualmachine.Name = *vm.Name
		virtualmachine.VMID = *vm.VMID
		virtualmachines = append(virtualmachines, virtualmachine)
	}
	return virtualmachines
}

func (c *RmVMClient) GetWorkspaceID(a autorest.Authorizer, vmName, resourceGroup, subscriptionID string) string {
	// query virtual machines
	computeClient := compute.NewVirtualMachinesClient(subscriptionID)
	computeClient.Authorizer = a
	vm, err := computeClient.Get(context.Background(), resourceGroup, vmName, "instanceView")
	if err != nil {
		panic(err)
	}

	var workspaceID string
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
				workspaceID = ws["workspaceId"]
			}
		}
	}
	return workspaceID
}

func (c *RmVMClient) GetManagedByTag(a autorest.Authorizer, vmName, resourceGroup, subscriptionID string) string {
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

	var managedBy string
	// get the ManagedBy Tag
	tags := make(map[string]string)
	managed, err := json.Marshal(vm.Tags)
	if err != nil {
		log.Println(err)
	}

	str := string(managed)
	json.Unmarshal([]byte(str), &tags)
	if tags[managedkey] == managedval {
		managedBy = tags[managedkey]
	} else {
		managedBy = client
	}

	return managedBy
}
