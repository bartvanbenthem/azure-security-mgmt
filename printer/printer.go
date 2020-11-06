package printer

import (
	"fmt"

	"github.com/Azure/go-autorest/autorest"
	"github.com/bartvanbenthem/azure-update-mgmt/law"
	"github.com/bartvanbenthem/azure-update-mgmt/vm"
)

type PrintClient struct{}

func (p PrintClient) VMFormatted(auth autorest.Authorizer, subscriptionID string) {
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

func (p PrintClient) LAWTableRowsCommaSep(auth autorest.Authorizer, workspaceID string) {
	var lawclient law.LAWClient
	var q law.KustoQuery
	qresult, err := lawclient.Query(auth, workspaceID, q.ComputerUpdatesList())
	if err != nil {
		fmt.Println(err)
	}

	bresult, err := lawclient.ResultParserByte(qresult)
	if err != nil {
		fmt.Println(err)
	}

	res := lawclient.ResultParserLAWQueryResult(bresult)

	for _, table := range res.Tables {
		for _, row := range table.Rows {
			for _, item := range row {
				fmt.Printf("%v,", item)
			}
			fmt.Printf("\n")
		}
	}
}

func (p PrintClient) LAWTableColumnsCommaSep(auth autorest.Authorizer, workspaceID string) {
	var lawclient law.LAWClient
	var q law.KustoQuery
	qresult, err := lawclient.Query(auth, workspaceID, q.ComputerUpdatesList())
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
}
