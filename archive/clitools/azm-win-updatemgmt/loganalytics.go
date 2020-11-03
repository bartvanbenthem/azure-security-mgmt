package main

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/operationalinsights/v1/operationalinsights"
	"github.com/Azure/go-autorest/autorest"
)

// for api calls to log analytics (non sdk)
type LogAnalyticsTable struct {
	Tables []struct {
		Name    string `json:"name"`
		Columns []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"columns"`
		Rows [][]interface{} `json:"rows"`
	} `json:"tables"`
}

type Windows struct{}

func (w *Windows) Assessed(a autorest.Authorizer, vmId, workspaceID string) bool {
	queryClient := operationalinsights.NewQueryClient()
	queryClient.Authorizer = a

	kustoQuery := fmt.Sprintf(`Heartbeat
	| where OSType == "Windows" and VMUUID=="%v" 
	| summarize arg_max(TimeGenerated, *) by SourceComputerId 
	| count`, vmId)

	var qb operationalinsights.QueryBody
	qb.Query = StringToPointer(kustoQuery)

	query, err := queryClient.Execute(context.Background(), workspaceID, qb)
	if err != nil {
		panic(err)
	}

	var num float64
	for _, table := range *query.Tables {
		for _, row := range *table.Rows {
			for _, i := range row {
				num = i.(float64)
			}
		}
	}
	var assessed bool
	if num == 1 {
		assessed = true
	} else {
		assessed = false
	}

	return assessed
}

func (w *Windows) UpdatesCount(a autorest.Authorizer, vmId, workspaceID string) float64 {
	queryClient := operationalinsights.NewQueryClient()
	queryClient.Authorizer = a

	kustoQuery := fmt.Sprintf(`Update
		| where TimeGenerated>ago(14h) and OSType!="Linux" and Optional==false and SourceComputerId in ((Heartbeat
		| where TimeGenerated>ago(12h) and OSType=~"Windows" and notempty(Computer)
		| summarize arg_max(TimeGenerated, Solutions) by SourceComputerId
		| where Solutions has "updates"
		| distinct SourceComputerId))
		| summarize hint.strategy=partitioned arg_max(TimeGenerated, *) by Computer, SourceComputerId, UpdateID
		| where UpdateState=~"Needed" and Approved!=false and VMUUID=="%v"
		| count`, vmId)

	var qb operationalinsights.QueryBody
	qb.Query = StringToPointer(kustoQuery)
	qb.Timespan = StringToPointer("PT12H")

	query, err := queryClient.Execute(context.Background(), workspaceID, qb)
	if err != nil {
		panic(err)
	}

	var num float64
	for _, table := range *query.Tables {
		for _, row := range *table.Rows {
			for _, i := range row {
				num = i.(float64)
			}
		}
	}

	return num
}

func (w *Windows) UpdatesCriticalSecurityCount(a autorest.Authorizer, vmId, workspaceID string) float64 {
	queryClient := operationalinsights.NewQueryClient()
	queryClient.Authorizer = a

	kustoQuery := fmt.Sprintf(`Update
		| where TimeGenerated>ago(14h) and OSType!="Linux" and (Classification has "Critical" or Classification has "Security") and SourceComputerId in ((Heartbeat
		| where TimeGenerated>ago(12h) and OSType=~"Windows" and notempty(Computer)
		| summarize arg_max(TimeGenerated, Solutions) by SourceComputerId
		| where Solutions has "updates"
		| distinct SourceComputerId))
		| summarize hint.strategy=partitioned arg_max(TimeGenerated, *) by Computer, SourceComputerId, UpdateID
		| where UpdateState=~"Needed" and Approved!=false and VMUUID=="%v"
		| count`, vmId)

	var qb operationalinsights.QueryBody
	qb.Query = StringToPointer(kustoQuery)
	qb.Timespan = StringToPointer("PT12H")

	query, err := queryClient.Execute(context.Background(), workspaceID, qb)
	if err != nil {
		panic(err)
	}

	var num float64
	for _, table := range *query.Tables {
		for _, row := range *table.Rows {
			for _, i := range row {
				num = i.(float64)
			}
		}
	}

	return num
}
