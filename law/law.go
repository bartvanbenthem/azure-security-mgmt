package law

import (
	"context"
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/services/operationalinsights/v1/operationalinsights"
	"github.com/Azure/go-autorest/autorest"
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
