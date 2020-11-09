package law

import (
	"fmt"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/bartvanbenthem/azure-security-mgmt/law"
)

type ComputerUpdatesList struct {
	Rows []ComputerListRow
}

type ComputerListRow struct {
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

func (c *ComputerUpdatesList) AddRow(row ComputerListRow) {
	c.Rows = append(c.Rows, row)
}

func (c *ComputerUpdatesList) ConvertToHumanReadable(qr ComputerUpdatesList) {}

func (c *ComputerUpdatesList) ReturnObject(auth autorest.Authorizer, workspaceID string) ComputerUpdatesList {
	var lawclient law.LAWClient
	qresult, err := lawclient.Query(auth, workspaceID, c.ComputerUpdatesListQuery())
	if err != nil {
		fmt.Println(err)
	}

	bresult, err := lawclient.ResultParserByte(qresult)
	if err != nil {
		fmt.Println(err)
	}

	res := lawclient.ResultParserLAWQueryResult(bresult)

	var computerlist ComputerUpdatesList

	// add rows to computerlist
	for _, table := range res.Tables {
		for _, row := range table.Rows {
			row := fmt.Sprintf("%v", row)
			rowItems := strings.Fields(row)
			result := ComputerListRow{ID: rowItems[0],
				DisplayName:                 rowItems[1],
				SourceComputerId:            rowItems[2],
				ScopedToUpdatesSolution:     rowItems[3],
				MissingCriticalUpdatesCount: rowItems[4],
				MissingSecurityUpdatesCount: rowItems[5],
				MissingOtherUpdatesCount:    rowItems[6],
				Compliance:                  rowItems[7],
				LastAssessedTime:            rowItems[8],
				LastUpdateAgentSeenTime:     "",
				OSType:                      rowItems[9],
				Environment:                 rowItems[10],
			}
			computerlist.AddRow(result)
		}
	}
	return computerlist
}

func (c *ComputerUpdatesList) ComputerUpdatesListQuery() string {
	const list string = `Heartbeat
	| where TimeGenerated>ago(12h) and OSType=="Linux" and notempty(Computer)
	| summarize arg_max(TimeGenerated, Solutions, Computer, ResourceId, ComputerEnvironment, VMUUID) by SourceComputerId
	| where Solutions has "updates"
	| extend vmuuId=VMUUID, azureResourceId=ResourceId, osType=1, environment=iff(ComputerEnvironment=~"Azure", 1, 2), scopedToUpdatesSolution=true, lastUpdateAgentSeenTime=""
	| join kind=leftouter
	(
		Update
		| where TimeGenerated>ago(5h) and OSType=="Linux" and SourceComputerId in ((Heartbeat
		| where TimeGenerated>ago(12h) and OSType=="Linux" and notempty(Computer)
		| summarize arg_max(TimeGenerated, Solutions) by SourceComputerId
		| where Solutions has "updates"
		| distinct SourceComputerId))
		| summarize hint.strategy=partitioned arg_max(TimeGenerated, UpdateState, Classification, Product, Computer, ComputerEnvironment) by SourceComputerId, Product, ProductArch
		| summarize Computer=any(Computer), ComputerEnvironment=any(ComputerEnvironment), missingCriticalUpdatesCount=countif(Classification has "Critical" and UpdateState=~"Needed"), missingSecurityUpdatesCount=countif(Classification has "Security" and UpdateState=~"Needed"), missingOtherUpdatesCount=countif(Classification !has "Critical" and Classification !has "Security" and UpdateState=~"Needed"), lastAssessedTime=max(TimeGenerated), lastUpdateAgentSeenTime="" by SourceComputerId
		| extend compliance=iff(missingCriticalUpdatesCount > 0 or missingSecurityUpdatesCount > 0, 2, 1)
		| extend ComplianceOrder=iff(missingCriticalUpdatesCount > 0 or missingSecurityUpdatesCount > 0 or missingOtherUpdatesCount > 0, 1, 3)
	)
	on SourceComputerId
	| project id=SourceComputerId, displayName=Computer, sourceComputerId=SourceComputerId, scopedToUpdatesSolution=true, missingCriticalUpdatesCount=coalesce(missingCriticalUpdatesCount, -1), missingSecurityUpdatesCount=coalesce(missingSecurityUpdatesCount, -1), missingOtherUpdatesCount=coalesce(missingOtherUpdatesCount, -1), compliance=coalesce(compliance, 4), lastAssessedTime, lastUpdateAgentSeenTime, osType=1, environment=iff(ComputerEnvironment=~"Azure", 1, 2), ComplianceOrder=coalesce(ComplianceOrder, 2)
	| union(Heartbeat
	| where TimeGenerated>ago(12h) and OSType=~"Windows" and notempty(Computer)
	| summarize arg_max(TimeGenerated, Solutions, Computer, ResourceId, ComputerEnvironment, VMUUID) by SourceComputerId
	| where Solutions has "updates"
	| extend vmuuId=VMUUID, azureResourceId=ResourceId, osType=2, environment=iff(ComputerEnvironment=~"Azure", 1, 2), scopedToUpdatesSolution=true, lastUpdateAgentSeenTime=""
	| join kind=leftouter
	(
		Update
		| where TimeGenerated>ago(14h) and OSType!="Linux" and SourceComputerId in ((Heartbeat
		| where TimeGenerated>ago(12h) and OSType=~"Windows" and notempty(Computer)
		| summarize arg_max(TimeGenerated, Solutions) by SourceComputerId
		| where Solutions has "updates"
		| distinct SourceComputerId))
		| summarize hint.strategy=partitioned arg_max(TimeGenerated, UpdateState, Classification, Title, Optional, Approved, Computer, ComputerEnvironment) by Computer, SourceComputerId, UpdateID
		| summarize Computer=any(Computer), ComputerEnvironment=any(ComputerEnvironment), missingCriticalUpdatesCount=countif(Classification has "Critical" and UpdateState=~"Needed" and Approved!=false), missingSecurityUpdatesCount=countif(Classification has "Security" and UpdateState=~"Needed" and Approved!=false), missingOtherUpdatesCount=countif(Classification !has "Critical" and Classification !has "Security" and UpdateState=~"Needed" and Optional==false and Approved!=false), lastAssessedTime=max(TimeGenerated), lastUpdateAgentSeenTime="" by SourceComputerId
		| extend compliance=iff(missingCriticalUpdatesCount > 0 or missingSecurityUpdatesCount > 0, 2, 1)
		| extend ComplianceOrder=iff(missingCriticalUpdatesCount > 0 or missingSecurityUpdatesCount > 0 or missingOtherUpdatesCount > 0, 1, 3)
	)
	on SourceComputerId
	| project id=SourceComputerId, displayName=Computer, sourceComputerId=SourceComputerId, scopedToUpdatesSolution=true, missingCriticalUpdatesCount=coalesce(missingCriticalUpdatesCount, -1), missingSecurityUpdatesCount=coalesce(missingSecurityUpdatesCount, -1), missingOtherUpdatesCount=coalesce(missingOtherUpdatesCount, -1), compliance=coalesce(compliance, 4), lastAssessedTime, lastUpdateAgentSeenTime, osType=2, environment=iff(ComputerEnvironment=~"Azure", 1, 2), ComplianceOrder=coalesce(ComplianceOrder, 2) )
	| order by ComplianceOrder asc, missingCriticalUpdatesCount desc, missingSecurityUpdatesCount desc, missingOtherUpdatesCount desc, displayName asc
	| project-away ComplianceOrder`

	return list
}
