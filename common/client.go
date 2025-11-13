package common

import (
	"context"
	"net/http"

	"resty.dev/v3"
)

type APIClient struct {
	Client   *resty.Client
	BASE_URL string
}

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// HeadersContextKey is the key used to store incoming client headers in context
	HeadersContextKey ContextKey = "client-headers"
)

// MakeRequest performs an HTTP request with support for dynamic headers and base URL from context
func (a *APIClient) MakeRequest(method string, endpoint string, queryParams map[string]string, result any, payload any, headers map[string]string) (*resty.Response, error) {
	request_url := a.BASE_URL + endpoint
	r := a.Client.R()

	var resp *resty.Response
	var err error

	r.SetQueryParams(queryParams)
	r.SetHeaders(headers)
	r.SetResult(result)

	switch method {
	case http.MethodGet:
		resp, err = r.Get(request_url)
	case http.MethodPut:
		resp, err = r.SetBody(payload).Put(request_url)
	case http.MethodPost:
		resp, err = r.SetBody(payload).Post(request_url)
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// MakeRequestWithContext performs an HTTP request with support for dynamic headers and base URL from context
// It extracts client headers from context and applies them to the request, overriding any static configuration
func (a *APIClient) MakeRequestWithContext(ctx context.Context, method string, endpoint string, queryParams map[string]string, result any, payload any, headers map[string]string, appType string) (*resty.Response, error) {
	// Start with the configured base URL
	baseURL := a.BASE_URL
	
	// Extract client headers from context if available
	clientHeaders := make(map[string]string)
	if ctx != nil {
		if ctxHeaders, ok := ctx.Value(HeadersContextKey).(map[string]string); ok {
			// Copy context headers
			for k, v := range ctxHeaders {
				clientHeaders[k] = v
			}
			
			// Check for dynamic base URL based on app type
			if appType == "ctix" {
				if ctixBaseURL, exists := ctxHeaders["X-Ctix-Base-Url"]; exists && ctixBaseURL != "" {
					baseURL = ctixBaseURL
					if baseURL[len(baseURL)-1] != '/' {
						baseURL += "/"
					}
					baseURL += "ctixapi/"
				}
			} else if appType == "co" {
				if coBaseURL, exists := ctxHeaders["X-Co-Base-Url"]; exists && coBaseURL != "" {
					baseURL = coBaseURL
					if baseURL[len(baseURL)-1] != '/' {
						baseURL += "/"
					}
				}
			}
		}
	}
	
	request_url := baseURL + endpoint
	r := a.Client.R()

	var resp *resty.Response
	var err error

	r.SetQueryParams(queryParams)
	
	// First set the function-provided headers
	if headers != nil {
		r.SetHeaders(headers)
	}
	
	// Then apply client headers from context (these will override any conflicts)
	// This ensures Authorization and other client headers take precedence
	if len(clientHeaders) > 0 {
		r.SetHeaders(clientHeaders)
	}
	
	r.SetResult(result)

	switch method {
	case http.MethodGet:
		resp, err = r.Get(request_url)
	case http.MethodPut:
		resp, err = r.SetBody(payload).Put(request_url)
	case http.MethodPost:
		resp, err = r.SetBody(payload).Post(request_url)
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}
