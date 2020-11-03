package updatemgmt

import "fmt"

type UpdateMgmtClient struct{}

type ComputerListQueryResult struct {
	ID                          string
	DisplayName                 string
	SourceComputerId            string
	ScopedToUpdatesSolution     string
	MissingCriticalUpdatesCount string
	MissingSecurityUpdatesCount string
	MissingOtherUpdatesCount    string
	Compliance                  string
	LastAssessedTime            string
	LastUpdateAgentSeenTime     string
	OSType                      string
	Environment                 string
}

func (c *UpdateMgmtClient) ComputerList() string {
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

func (c *UpdateMgmtClient) MissingUpdatesList() string {
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

func (c *UpdateMgmtClient) MissingUpdatesSingleWinList(VMUUID string) string {
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

func (c *UpdateMgmtClient) MissingUpdatesSingleLinuxList(VMUUID01, VMUUID02 string) string {
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
