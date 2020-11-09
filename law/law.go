package law

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/operationalinsights/v1/operationalinsights"
	"github.com/Azure/go-autorest/autorest"
	"github.com/bartvanbenthem/azure-security-mgmt/law"
)

type LAWQueryResult struct {
	Tables []struct {
		Name    string `json:"name"`
		Columns []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"columns"`
		Rows [][]interface{} `json:"rows"`
	} `json:"tables"`
}

type LAWClient struct{}

func StringToPointer(s string) *string { return &s }

func (c *LAWClient) Query(a autorest.Authorizer, workspaceID, lawQuery string) (operationalinsights.QueryResults, error) {
	queryClient := operationalinsights.NewQueryClient()
	queryClient.Authorizer = a

	raw := lawQuery
	var qbody operationalinsights.QueryBody
	var qresult operationalinsights.QueryResults

	qbody.Query = StringToPointer(raw)

	qresult, err := queryClient.Execute(context.Background(), workspaceID, qbody)
	if err != nil {
		return qresult, err
	}

	return qresult, err
}

func (c *LAWClient) ReturnRowSlice(auth autorest.Authorizer, workspaceID string) []string {
	var lawclient law.LAWClient
	var q law.KustoQuery
	qresult, err := lawclient.Query(auth, workspaceID, q.ComputerUpdatesList())
	if err != nil {
		fmt.Println(err)
	}

	var result string
	var results []string
	for _, table := range *qresult.Tables {
		for _, row := range *table.Rows {
			result = fmt.Sprintf("%v", row)
			results = append(results, result)
		}
	}
	return results
}

func (c *LAWClient) ReturnColumnSlice(auth autorest.Authorizer, workspaceID string) []string {
	var lawclient law.LAWClient
	var q law.KustoQuery
	qresult, err := lawclient.Query(auth, workspaceID, q.ComputerUpdatesList())
	if err != nil {
		fmt.Println(err)
	}

	var columns []string
	for _, table := range *qresult.Tables {
		for _, col := range *table.Columns {
			columns = append(columns, *col.Name)
		}
	}
	return columns
}

func (c *LAWClient) ResultParserByte(qresult operationalinsights.QueryResults) ([]byte, error) {
	result, err := json.Marshal(qresult)
	if err != nil {
		return result, err
	}

	return result, err
}

func (c *LAWClient) ResultParserLAWQueryResult(qresult []byte) LAWQueryResult {
	var law LAWQueryResult
	json.Unmarshal(qresult, &law)
	return law
}
