# Dynamic Authentication and Backend Routing

This document describes the dynamic authentication and backend routing feature that allows the MCP server to accept client-provided authentication credentials and backend URLs via HTTP headers.

## Overview

The MCP Server now supports dynamic authentication and backend routing through client-provided HTTP headers. This allows different clients to connect to different backend instances and use different authentication credentials without requiring server-side configuration changes.

## Features

1. **Dynamic Authentication**: Client can provide authentication credentials via the `Authorization` header
2. **Dynamic Backend Routing**: Client can specify backend URLs via custom headers
3. **Header Forwarding**: All client headers are forwarded to backend services
4. **Backward Compatibility**: Static configuration in `config.yaml` still works as fallback

## Client Configuration

Clients should register the MCP server with headers in their configuration:

```json
{
  "mcpServers": {
    "cyware-ctix": {
      "url": "http://localhost:8000/sse",
      "headers": [
        "Authorization: Bearer your-token-here",
        "X-CTIX-BASE-URL: https://your-ctix-instance.com/ctixapi/"
      ]
    },
    "cyware-co": {
      "url": "http://localhost:8000/sse",
      "headers": [
        "Authorization: Bearer your-token-here",
        "X-CO-BASE-URL: https://your-co-instance.com"
      ]
    }
  }
}
```

## Supported Headers

### Authentication Headers

- **`Authorization`**: Bearer token or other authentication scheme
  - Example: `Authorization: Bearer CYW your-token-here`
  - This header overrides the static authentication configured in `config.yaml`

### Backend Routing Headers

- **`X-CTIX-BASE-URL`**: Base URL for CTIX backend
  - Example: `X-CTIX-BASE-URL: https://demo.cyware.com/ctixapi/`
  - Used for all CTIX tool operations

- **`X-CO-BASE-URL`**: Base URL for CO (Cyware Orchestrator) backend
  - Example: `X-CO-BASE-URL: https://demo.cyware.com`
  - Used for all CO tool operations

### Case Sensitivity

The server checks for headers in multiple formats:
- `X-CTIX-BASE-URL` and `X-Ctix-Base-Url`
- `X-CO-BASE-URL` and `X-Co-Base-Url`

## How It Works

### Request Flow

1. **Client Connection**: Client connects to the MCP server's SSE endpoint with headers
2. **Header Extraction**: The SSE context function stores HTTP headers in the request context
3. **Tool Invocation**: When a tool is invoked, the handler extracts headers from the context
4. **Header Forwarding**: Headers are passed to the backend API call
5. **Dynamic Routing**: If a dynamic base URL header is present, it overrides the default URL
6. **Backend Request**: The HTTP client forwards all headers to the backend service

### Code Architecture

#### 1. Context Function (cmd/main.go)

```go
sseServer := server.NewSSEServer(s,
    server.WithSSEContextFunc(func(ctx context.Context, r *http.Request) context.Context {
        return context.WithValue(ctx, common.HeadersContextKey, r.Header)
    }),
)
```

This injects HTTP headers into the request context for every incoming request.

#### 2. Header Extraction (common/utils.go)

```go
func PrepareRequestHeaders(ctx context.Context) map[string]string {
    httpHeaders := ExtractHeadersFromContext(ctx)
    headers := ConvertHeadersToMap(httpHeaders)
    // Filter out connection-specific headers
    return filteredHeaders
}
```

#### 3. Tool Handler Pattern

All tool handlers follow this pattern:

```go
s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Extract client headers
    headers := common.PrepareRequestHeaders(ctx)

    // Call API function with headers
    resp, err := GetSomeData(params, headers)

    return common.MCPToolResponse(resp, []int{200}, err)
})
```

#### 4. API Function Pattern

All API functions accept and forward headers:

```go
func GetSomeData(params map[string]string, headers map[string]string) (*common.APIResponse, error) {
    result := SomeResponse{}
    resp, err := CLIENT.MakeRequest("GET", endpoint, params, &result, nil, headers)
    return &common.APIResponse{
        FilteredReponse: common.JsonifyResponse(result),
        RawResponse:     resp,
    }, err
}
```

#### 5. HTTP Client (common/client.go)

```go
func (a *APIClient) MakeRequest(..., headers map[string]string) (*resty.Response, error) {
    // Check for dynamic base URL in headers
    if dynamicBaseURL := extractDynamicBaseURL(headers); dynamicBaseURL != "" {
        baseURL = dynamicBaseURL
    }

    // Forward all headers to backend
    r.SetHeaders(headers)

    // Make request
    resp, err := r.Get(baseURL + endpoint)
    return resp, err
}
```

## Fallback Behavior

If client headers are not provided, the server falls back to the static configuration in `config.yaml`:

```yaml
applications:
  ctix:
    base_url: "https://demo.cyware.com/ctix/"
    auth:
      type: "token"
      token: "your-static-token"
  co:
    base_url: "https://demo.cyware.com/soar/"
    auth:
      type: "token"
      token: "your-static-token"
```

## Testing

### With curl

You can test the dynamic routing with curl:

```bash
# Test CTIX with dynamic headers
curl -N -H "Authorization: Bearer CYW your-token" \
     -H "X-CTIX-BASE-URL: https://your-ctix-instance.com/ctixapi/" \
     http://localhost:8000/sse

# Test CO with dynamic headers
curl -N -H "Authorization: Bearer CYW your-token" \
     -H "X-CO-BASE-URL: https://your-co-instance.com" \
     http://localhost:8000/sse
```

### With MCP Client

Configure your MCP client (Claude Desktop, etc.) with the headers as shown in the Client Configuration section above.

## Security Considerations

1. **Header Filtering**: The server filters out connection-specific headers (Host, Connection, etc.) to prevent header injection attacks
2. **HTTPS Recommended**: Always use HTTPS in production to protect authentication tokens in transit
3. **Token Validation**: The backend services are responsible for validating authentication tokens
4. **Base URL Validation**: Clients should be trusted, as they control which backend URLs are accessed

## Updated Files

### Core Infrastructure
- `common/utils.go`: Header extraction and context utilities
- `common/client.go`: Dynamic base URL support and header forwarding
- `cmd/main.go`: SSE context function configuration

### CTIX Tools
- `applications/ctix/user_details.go`
- `applications/ctix/threat_data_list.go`
- `applications/ctix/threat_data_details.go`
- `applications/ctix/enrichment.go`
- `applications/ctix/quick_add_intel.go`
- `applications/ctix/threat_data_bulk_actions.go`
- `applications/ctix/tags.go`

### CO Tools
- `applications/co/playbooks.go`
- `applications/co/apps.go`

## Backward Compatibility

This feature is fully backward compatible:

1. **Static Configuration**: If no headers are provided, the server uses static configuration from `config.yaml`
2. **Existing Clients**: Clients that don't send headers will continue to work with static configuration
3. **Partial Headers**: If only some headers are provided, the server uses those and falls back to config for the rest

## Examples

### Example 1: Multi-Tenant Setup

Different clients connecting to different tenant backends:

**Tenant A Client:**
```json
{
  "headers": [
    "Authorization: Bearer tenant-a-token",
    "X-CTIX-BASE-URL: https://tenant-a.cyware.com/ctixapi/"
  ]
}
```

**Tenant B Client:**
```json
{
  "headers": [
    "Authorization: Bearer tenant-b-token",
    "X-CTIX-BASE-URL: https://tenant-b.cyware.com/ctixapi/"
  ]
}
```

### Example 2: Development vs Production

Same client connecting to different environments:

**Development:**
```json
{
  "headers": [
    "Authorization: Bearer dev-token",
    "X-CTIX-BASE-URL: https://dev.cyware.com/ctixapi/"
  ]
}
```

**Production:**
```json
{
  "headers": [
    "Authorization: Bearer prod-token",
    "X-CTIX-BASE-URL: https://prod.cyware.com/ctixapi/"
  ]
}
```

## Troubleshooting

### Headers Not Being Forwarded

1. Verify the server is running in SSE mode (not stdio)
2. Check that the client is sending headers correctly
3. Enable debug logging to see header extraction

### Authentication Failures

1. Verify the token format (should include "Bearer " prefix if required)
2. Check that the token is valid on the backend
3. Ensure the Authorization header is being sent by the client

### Wrong Backend URL

1. Verify the header name matches exactly: `X-CTIX-BASE-URL` or `X-CO-BASE-URL`
2. Ensure the URL includes the full path (e.g., `/ctixapi/` for CTIX)
3. Check that the URL is accessible from the MCP server

## Future Enhancements

Potential improvements for future versions:

1. **Header Validation**: Validate backend URLs against a whitelist
2. **Rate Limiting**: Per-client rate limiting based on headers
3. **Metrics**: Track usage by client/tenant based on headers
4. **Logging**: Enhanced logging of header usage for debugging
