package law

import (
	"fmt"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/bartvanbenthem/azure-security-mgmt/law"
	"github.com/bartvanbenthem/azure-update-mgmt/vm"
)

type ComputerListQueryResult struct {
	ID                          string `json:"id"`
	DisplayName                 string `json:"displayName"`
	SourceComputerId            string `json:"sourceComputerId"`
	ScopedToUpdatesSolution     string `json:"scopedToUpdatesSolution"`
	MissingCriticalUpdatesCount string `json:"missingCriticalUpdatesCount"`
	MissingSecurityUpdatesCount string `json:"missingSecurityUpdatesCount"`
	MissingOtherUpdatesCount    string `json:"missingOtherUpdatesCount"`
	Compliance                  string `json:"compliance"`
	LastAssessedTime            string `json:"lastAssessedTime"`
	LastUpdateAgentSeenTime     string `json:"lastUpdateAgentSeenTime"`
	OSType                      string `json:"osType"`
	Environment                 string `json:"environment"`
}

func (c ComputerListQueryResult) ReturnObject(auth autorest.Authorizer, workspaceID string) []ComputerListQueryResult {
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

	var result ComputerListQueryResult
	var results []ComputerListQueryResult
	for _, table := range res.Tables {
		for _, row := range table.Rows {
			row := fmt.Sprintf("%v", row)
			rowItems := strings.Fields(row)
			result.ID = rowItems[0]
			result.DisplayName = rowItems[1]
			result.SourceComputerId = rowItems[2]
			result.ScopedToUpdatesSolution = rowItems[3]
			result.MissingCriticalUpdatesCount = rowItems[4]
			result.MissingSecurityUpdatesCount = rowItems[5]
			result.MissingOtherUpdatesCount = rowItems[6]
			result.Compliance = rowItems[7]
			result.LastAssessedTime = rowItems[8]
			result.LastUpdateAgentSeenTime = ""
			result.OSType = rowItems[9]
			result.Environment = rowItems[10]
			results = append(results, result)
		}
	}
	return results
}

func (c ComputerListQueryResult) FormattedPrint(auth autorest.Authorizer, subscriptionID string) {
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
