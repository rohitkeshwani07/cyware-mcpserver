package ctix

import (
	"context"

	"github.com/cyware-labs/cyware-mcpserver/applications/ctix/helpers"
	"github.com/cyware-labs/cyware-mcpserver/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type BulkActionResponse struct {
	Message string `json:"message"`
}

func ThreatDataListBulkAction(ctx context.Context, endpoint string, payload any) (*common.APIResponse, error) {
	bulkResp := BulkActionResponse{}
	resp, err := CTIX_CLIENT.MakeRequestWithContext(ctx, "POST", endpoint, nil, &bulkResp, payload, nil, "ctix")
	return &common.APIResponse{
		FilteredReponse: common.JsonifyResponse(bulkResp),
		RawResponse:     resp,
	}, err
}

// This function uses an action map and registers tools for all the bulk actions of threat data
func ThreatDataListBulkActionTools(s *server.MCPServer) {
	mp := helpers.GetThreatDataBulkActionsMapping()
	for _, v := range mp {
		tool := mcp.NewToolWithRawSchema(v["tool_name"], v["tool_description"], []byte(v["schema"]))

		s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			resp, err := ThreatDataListBulkAction(ctx, v["endpoint"], request.Params.Arguments)
			return common.MCPToolResponse(resp, []int{200}, err)
		})
	}
}
