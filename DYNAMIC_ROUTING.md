# Dynamic Authentication and Backend Routing

This document describes the implementation of dynamic authentication and backend routing for the Cyware MCP Server via HTTP mode.

## Overview

The MCP Server now supports client-side configuration for authentication and backend routing via a streamable HTTP endpoint. This allows clients to specify authentication tokens and backend URLs via HTTP headers instead of relying solely on the server's configuration file.

**Note**: This feature requires using `mcp_mode: "http"` in the configuration.

## Features

### 1. Client-Side Authentication

Clients can now provide authentication credentials via the `Authorization` header in every request. This header will override any static authentication configured in the server's `config.yaml` file.

### 2. Dynamic Backend Routing

The server can now route requests to different backend instances based on headers provided by the client:

- **CTIX Backend**: Use the `X-Ctix-Base-Url` header to specify the CTIX backend URL
- **CO Backend**: Use the `X-Co-Base-Url` header to specify the CO backend URL

### 3. Header Forwarding

All incoming HTTP headers from the client are automatically extracted and forwarded to the backend services when making API calls.

## Client Configuration

### HTTP Mode (Recommended)

Configure the server to use HTTP mode in `config.yaml`:

```yaml
server:
  mcp_mode: "http"
  port: "8000"
```

Then make requests to the `/mcp` endpoint with custom headers:

```bash
curl -X POST http://localhost:8000/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -H "X-Ctix-Base-Url: https://ctix-backend.example.com" \
  -H "X-Co-Base-Url: https://co-backend.example.com" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "logged-in-user-details",
      "arguments": {}
    }
  }'
```

## Implementation Details

### Architecture

1. **HTTP Handler**: The `/mcp` endpoint receives POST requests with JSON-RPC payloads and extracts all HTTP headers.

2. **Context Injection**: All HTTP headers are extracted from the request and injected into the request context before processing.

3. **Dynamic Header Resolution**: The `MakeRequestWithContext` method in `APIClient` extracts headers from the context and applies them to outgoing requests.

4. **Base URL Override**: If a client provides `X-Ctix-Base-Url` or `X-Co-Base-Url` headers, these will override the configured base URLs for that specific request.

5. **Authentication Priority**: Client-provided `Authorization` headers take precedence over server-configured authentication.

### Key Components

#### `common/client.go`

- `MakeRequestWithContext()`: Enhanced method that accepts context and extracts client headers
- Supports both static (config file) and dynamic (client headers) configuration
- Automatically applies the correct base URL based on the app type (ctix/co)

#### `common/utils.go`

- `ExtractHeadersFromHTTPRequest()`: Extracts all headers from an HTTP request
- `InjectHeadersIntoContext()`: Injects headers into a context
- `GetHeadersFromContext()`: Retrieves headers from a context

#### `cmd/main.go`

- Configures HTTP server with `/mcp` endpoint for JSON-RPC requests
- Supports `http`, `sse` (legacy), and `stdio` modes

#### `cmd/http_handler.go`

- HTTP handler that extracts headers and injects them into context
- Delegates JSON-RPC processing to the MCP server's `HandleMessage` method

### Tool Handler Updates

All tool handlers in both CTIX and CO applications have been updated to:

1. Accept `context.Context` as the first parameter
2. Pass the context through to API client methods
3. Use `MakeRequestWithContext()` instead of `MakeRequest()`

## Backward Compatibility

The implementation maintains full backward compatibility:

- Static configuration in `config.yaml` continues to work
- If no client headers are provided, the server falls back to configured values
- The original `MakeRequest()` method is still available for internal use

## Header Normalization

All headers are normalized to their canonical form (e.g., `x-ctix-base-url` becomes `X-Ctix-Base-Url`) to ensure consistent header matching regardless of the client's header casing.

## Testing

To test the dynamic routing:

1. **Start the server in HTTP mode**:
   ```bash
   ./cyware-mcp-server -config_path config.yaml
   ```
   Make sure `mcp_mode: "http"` and a port is configured in `config.yaml`.

2. **Make a test request** with custom headers:
   ```bash
   curl -X POST http://localhost:8000/mcp \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer test-token-123" \
     -H "X-Ctix-Base-Url: https://your-ctix.com" \
     -d '{
       "jsonrpc": "2.0",
       "id": 1,
       "method": "tools/list"
     }'
   ```

3. **Verify**:
   - The `Authorization` header is forwarded to the backend
   - Requests are routed to the URLs specified in `X-Ctix-Base-Url` or `X-Co-Base-Url`
   - All other headers are also forwarded
   - See [TESTING_RESULTS.md](TESTING_RESULTS.md) for detailed test results

4. **Test fallback behavior**:
   - Remove client headers and verify the server uses `config.yaml` settings
   - Mix client and config settings to verify priority handling

## Security Considerations

1. **Header Validation**: The server forwards all headers as-is. Ensure backend services validate incoming requests.

2. **Token Security**: Client tokens are transmitted in every request. Always use HTTPS in production.

3. **Base URL Validation**: The server does not validate base URLs provided by clients. Ensure proper network-level controls are in place.

## Troubleshooting

### Headers Not Being Forwarded

- Verify you're using SSE mode (`mcp_mode: "sse"` in config.yaml)
- Check that the client is sending headers in the correct format
- Ensure the MCP client library supports custom headers

### Base URL Not Being Applied

- Verify the header name matches exactly: `X-Ctix-Base-Url` or `X-Co-Base-Url`
- Check that the URL includes the scheme (http:// or https://)
- Ensure the base URL does not have a trailing slash (the server adds it automatically)

### Authentication Failures

- Verify the `Authorization` header format matches what the backend expects
- For Cyware tokens, ensure they're prefixed with `CYW ` (the server adds this automatically)
- Check backend logs for authentication errors

## Example Usage

### Using curl to test SSE endpoint

```bash
# List CTIX users with custom authentication and base URL
curl -X POST http://localhost:8000/message \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -H "X-Ctix-Base-Url: https://your-ctix-instance.com" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "get-ctix-user-list",
      "arguments": {
        "params": {
          "page": "1",
          "page_size": "10"
        }
      }
    }
  }'
```

## Future Enhancements

Potential future improvements:

1. Header validation and sanitization
2. URL allowlist for security
3. Per-tool header configuration
4. Header-based rate limiting
5. Audit logging of client-provided configurations
