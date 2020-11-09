package law

import (
	"fmt"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/bartvanbenthem/azure-security-mgmt/law"
	"github.com/bartvanbenthem/azure-update-mgmt/vm"
)

type UpdateMgmtQuery struct{}

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

func (q *UpdateMgmtQuery) ComputerUpdatesList() string {
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

func (q *UpdateMgmtQuery) MissingUpdatesList() string {
	const list string = `Update
	| where TimeGenerated>ago(5h) and OSType=="Linux" and SourceComputerId in ((Heartbeat
	| where TimeGenerated>ago(12h) and OSType=="Linux" and notempty(Computer)
	| summarize arg_max(TimeGenerated, Solutions) by SourceComputerId
	| where Solutions has "updates"
	| distinct SourceComputerId))
	| summarize hint.strategy=partitioned arg_max(TimeGenerated, UpdateState, Classification, BulletinUrl, BulletinID) by SourceComputerId, Product, ProductArch
	| where UpdateState=~"Needed"
	| project-away UpdateState, TimeGenerated
	| summarize computersCount=dcount(SourceComputerId, 2), ClassificationWeight=max(iff(Classification has "Critical", 4, iff(Classification has "Security", 2, 1))) by id=strcat(Product, "_", ProductArch), displayName=Product, productArch=ProductArch, classification=Classification, InformationId=BulletinID, InformationUrl=tostring(split(BulletinUrl, ";", 0)[0]), osType=1
	| union(Update
	| where TimeGenerated>ago(14h) and OSType!="Linux" and (Optional==false or Classification has "Critical" or Classification has "Security") and SourceComputerId in ((Heartbeat
	| where TimeGenerated>ago(12h) and OSType=~"Windows" and notempty(Computer)
	| summarize arg_max(TimeGenerated, Solutions) by SourceComputerId
	| where Solutions has "updates"
	| distinct SourceComputerId))
	| summarize hint.strategy=partitioned arg_max(TimeGenerated, UpdateState, Classification, Title, KBID, PublishedDate, Approved) by Computer, SourceComputerId, UpdateID
	| where UpdateState=~"Needed" and Approved!=false
	| project-away UpdateState, Approved, TimeGenerated
	| summarize computersCount=dcount(SourceComputerId, 2), displayName=any(Title), publishedDate=min(PublishedDate), ClassificationWeight=max(iff(Classification has "Critical", 4, iff(Classification has "Security", 2, 1))) by id=strcat(UpdateID, "_", KBID), classification=Classification, InformationId=strcat("KB", KBID), InformationUrl=iff(isnotempty(KBID), strcat("https://support.microsoft.com/kb/", KBID), ""), osType=2)
	| sort by ClassificationWeight desc, computersCount desc, displayName asc
	| extend informationLink=(iff(isnotempty(InformationId) and isnotempty(InformationUrl), toobject(strcat('{ "uri": "', InformationUrl, '", "text": "', InformationId, '", "target": "blank" }')), toobject('')))
	| project-away ClassificationWeight, InformationId, InformationUrl`

	return list
}

func (q *UpdateMgmtQuery) MissingUpdatesSingleWinList(VMUUID string) string {
	list := fmt.Sprintf(`Update
	| where TimeGenerated>ago(14h) and OSType!="Linux" and (Optional==false or Classification has "Critical" or Classification has "Security") and VMUUID=~"%v"
	| summarize hint.strategy=partitioned arg_max(TimeGenerated, UpdateState, Classification, Title, KBID, PublishedDate, Approved) by Computer, SourceComputerId, UpdateID
	| where UpdateState=~"Needed" and Approved!=false
	| project-away UpdateState, Approved, TimeGenerated
	| summarize computersCount=dcount(SourceComputerId, 2), displayName=any(Title), publishedDate=min(PublishedDate), ClassificationWeight=max(iff(Classification has "Critical", 4, iff(Classification has "Security", 2, 1))) by id=strcat(UpdateID, "_", KBID), classification=Classification, InformationId=strcat("KB", KBID), InformationUrl=iff(isnotempty(KBID), strcat("https://support.microsoft.com/kb/", KBID), ""), osType=2
	| sort by ClassificationWeight desc, computersCount desc, displayName asc
	| extend informationLink=(iff(isnotempty(InformationId) and isnotempty(InformationUrl), toobject(strcat('{ "uri": "', InformationUrl, '", "text": "', InformationId, '", "target": "blank" }')), toobject('')))
	| project-away ClassificationWeight, InformationId, InformationUrl`, VMUUID)

	return list
}

func (q *UpdateMgmtQuery) MissingUpdatesSingleLinuxList(VMUUID01, VMUUID02 string) string {
	list := fmt.Sprintf(`Update
	| where TimeGenerated>ago(5h) and OSType=="Linux" and (VMUUID=~"%v" or VMUUID=~"%v")
	| summarize hint.strategy=partitioned arg_max(TimeGenerated, UpdateState, Classification, BulletinUrl, BulletinID) by Computer, SourceComputerId, Product, ProductArch
	| where UpdateState=~"Needed"
	| project-away UpdateState, TimeGenerated
	| summarize computersCount=dcount(SourceComputerId, 2), ClassificationWeight=max(iff(Classification has "Critical", 4, iff(Classification has "Security", 2, 1))) by id=strcat(Product, "_", ProductArch), displayName=Product, productArch=ProductArch, classification=Classification, InformationId=BulletinID, InformationUrl=tostring(split(BulletinUrl, ";", 0)[0]), osType=1
	| sort by ClassificationWeight desc, computersCount desc, displayName asc
	| extend informationLink=(iff(isnotempty(InformationId) and isnotempty(InformationUrl), toobject(strcat('{ "uri": "', InformationUrl, '", "text": "', InformationId, '", "target": "blank" }')), toobject('')))
	| project-away ClassificationWeight, InformationId, InformationUrl`, VMUUID01, VMUUID02)

	return list
}
