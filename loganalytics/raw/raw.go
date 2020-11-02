package raw

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/operationalinsights/v1/operationalinsights"
	"github.com/Azure/go-autorest/autorest"
)

// for api calls to log analytics (non sdk)
type Output struct {
	Tables []struct {
		Name    string `json:"name"`
		Columns []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"columns"`
		Rows [][]interface{} `json:"rows"`
	} `json:"tables"`
}

type QueryResultRaw struct {
	Header string
	Rows   []string
}

type RawClient struct{}

func StringToPointer(s string) *string { return &s }

func (c *RawClient) Query(a autorest.Authorizer, workspaceID, lawQuery string) (operationalinsights.QueryResults, error) {
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

// returns log analytics query result as a comma sepparated object
func (c *RawClient) ReturnQueryResultCommaSep(qresult operationalinsights.QueryResults) QueryResultRaw {
	var result QueryResultRaw
	for _, table := range *qresult.Tables {
		for _, column := range *table.Columns {
			str := fmt.Sprintf("%v,", *column.Name)
			result.Header += str
		}
	}
	for _, table := range *qresult.Tables {
		for _, r := range *table.Rows {
			var row string
			for _, item := range r {
				str := fmt.Sprintf("%v,", item)
				row += str
			}
			result.Rows = append(result.Rows, row)
		}
	}

	return result
}

func (c *RawClient) PrintQueryResultCommaSep(qresult operationalinsights.QueryResults) {
	for _, table := range *qresult.Tables {
		for _, column := range *table.Columns {
			fmt.Printf("%v,", *column.Name)
		}
	}
	for _, table := range *qresult.Tables {
		for _, row := range *table.Rows {
			fmt.Printf("\n")
			for _, item := range row {
				fmt.Printf("%v,", item)
			}
		}
		fmt.Printf("\n")
	}
}
