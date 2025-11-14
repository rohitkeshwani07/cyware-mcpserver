package common

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"resty.dev/v3"
)

const retry = 4

var failed_status = []int{400, 401}

func FormatCywareToken(rawToken string) string {
	const prefix = "CYW "

	if rawToken == "" {
		return ""
	}

	if strings.HasPrefix(rawToken, prefix) {
		return rawToken
	}

	return prefix + rawToken
}

// Base64Encode encodes the input string to Base64 format
func Base64Encode(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// GenerateAuthParams generates authentication parameters
func GenerateAuthParams(accessID, secretKey string) map[string]string {
	// Generating unix timestamp
	unixTimestamp := time.Now().Unix()

	// Adding 20 seconds for expires
	expires := unixTimestamp + 20

	// Creating the string to sign
	toSign := accessID + "\n" + strconv.FormatInt(expires, 10)

	// Generating HMAC-SHA1 hash
	h := hmac.New(sha1.New, []byte(secretKey))
	h.Write([]byte(toSign))
	hash := h.Sum(nil)

	// Converting to base64
	hashInBase64 := base64.StdEncoding.EncodeToString(hash)

	params := map[string]string{
		"Expires":   strconv.FormatInt(expires, 10),
		"AccessID":  accessID,
		"Signature": hashInBase64,
	}
	return params
}

// ExtractParams extracts params key from the tool call request and convert them into a map
func ExtractParams(request mcp.CallToolRequest, params_list []string) map[string]string {
	params := map[string]string{}
	mp, ok := request.Params.Arguments["params"].(map[string]interface{})
	if !ok {
		return params
	}

	for _, v := range params_list {
		if _, ok := mp[v]; ok {
			params[v] = mp[v].(string)
		}
	}
	return params
}

func GetRestyClient(retryHook func(r *resty.Response, err error)) *resty.Client {
	c := resty.New()
	c.SetAllowNonIdempotentRetry(true)
	c.SetRetryCount(retry)
	c.SetRetryWaitTime(1 * time.Second)

	// Retry condition
	c.AddRetryConditions(func(r *resty.Response, err error) bool {
		return r != nil && ContainsStatusCode(failed_status, r.StatusCode())
	})
	c.AddRetryHooks(retryHook)
	return c
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// HeadersContextKey is the key used to store HTTP headers in context
	HeadersContextKey contextKey = "http-headers"
)

// ExtractHeadersFromContext extracts HTTP headers from the request context.
// The mcp-go SSE server stores incoming request headers in the context.
func ExtractHeadersFromContext(ctx context.Context) http.Header {
	if headers, ok := ctx.Value(HeadersContextKey).(http.Header); ok {
		return headers
	}
	// Also check for standard http.Header context key used by mcp-go
	if headers, ok := ctx.Value("headers").(http.Header); ok {
		return headers
	}
	return http.Header{}
}

// ConvertHeadersToMap converts http.Header to a simple map[string]string
// by taking the first value of each header
func ConvertHeadersToMap(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// GetDynamicBaseURL extracts the base URL from client-provided headers.
// It checks for X-CO-BASE-URL and X-CTIX-BASE-URL headers.
// Returns empty string if no dynamic base URL header is found.
func GetDynamicBaseURL(headers map[string]string, appType string) string {
	// Check for application-specific base URL headers
	if appType == "co" {
		if baseURL, ok := headers["X-Co-Base-Url"]; ok && baseURL != "" {
			return baseURL
		}
		if baseURL, ok := headers["X-CO-BASE-URL"]; ok && baseURL != "" {
			return baseURL
		}
	} else if appType == "ctix" {
		if baseURL, ok := headers["X-Ctix-Base-Url"]; ok && baseURL != "" {
			return baseURL
		}
		if baseURL, ok := headers["X-CTIX-BASE-URL"]; ok && baseURL != "" {
			return baseURL
		}
	}

	// Fallback: check both headers regardless of app type
	if baseURL, ok := headers["X-Co-Base-Url"]; ok && baseURL != "" {
		return baseURL
	}
	if baseURL, ok := headers["X-CO-BASE-URL"]; ok && baseURL != "" {
		return baseURL
	}
	if baseURL, ok := headers["X-Ctix-Base-Url"]; ok && baseURL != "" {
		return baseURL
	}
	if baseURL, ok := headers["X-CTIX-BASE-URL"]; ok && baseURL != "" {
		return baseURL
	}

	return ""
}

// PrepareRequestHeaders extracts headers from context and prepares them for forwarding to backend.
// This helper is used by tool handlers to enable dynamic authentication and routing.
func PrepareRequestHeaders(ctx context.Context) map[string]string {
	httpHeaders := ExtractHeadersFromContext(ctx)
	if httpHeaders == nil || len(httpHeaders) == 0 {
		return nil
	}

	// Convert to map[string]string for easier handling
	headers := ConvertHeadersToMap(httpHeaders)

	// Filter out headers that should not be forwarded
	// (e.g., internal MCP headers, connection-specific headers)
	filteredHeaders := make(map[string]string)

	// List of headers to exclude from forwarding
	excludeHeaders := map[string]bool{
		"Host":              true,
		"Connection":        true,
		"Accept-Encoding":   true,
		"Content-Length":    true,
		"Transfer-Encoding": true,
		"Upgrade":           true,
	}

	for key, value := range headers {
		if !excludeHeaders[key] {
			filteredHeaders[key] = value
		}
	}

	return filteredHeaders
}
