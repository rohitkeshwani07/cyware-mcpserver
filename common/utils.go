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

// ExtractHeadersFromHTTPRequest extracts all headers from an HTTP request
// and returns them as a map[string]string
func ExtractHeadersFromHTTPRequest(r *http.Request) map[string]string {
	headers := make(map[string]string)
	if r == nil {
		return headers
	}
	
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0] // Take the first value if multiple exist
		}
	}
	return headers
}

// InjectHeadersIntoContext creates a new context with the given headers
func InjectHeadersIntoContext(ctx context.Context, headers map[string]string) context.Context {
	return context.WithValue(ctx, HeadersContextKey, headers)
}

// GetHeadersFromContext extracts headers from the given context
func GetHeadersFromContext(ctx context.Context) map[string]string {
	if ctx == nil {
		return make(map[string]string)
	}
	
	if headers, ok := ctx.Value(HeadersContextKey).(map[string]string); ok {
		return headers
	}
	return make(map[string]string)
}

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
