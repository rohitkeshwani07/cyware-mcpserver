package common

import (
	"net/http"

	"resty.dev/v3"
)

type APIClient struct {
	Client   *resty.Client
	BASE_URL string
}

// MakeRequest sends an HTTP request with support for dynamic base URLs and header forwarding.
// If headers contain a dynamic base URL (X-CO-BASE-URL or X-CTIX-BASE-URL), it overrides the default BASE_URL.
// All provided headers are forwarded to the backend service.
func (a *APIClient) MakeRequest(method string, endpoint string, queryParams map[string]string, result any, payload any, headers map[string]string) (*resty.Response, error) {
	// Use default base URL
	baseURL := a.BASE_URL

	// Check if there's a dynamic base URL in headers
	// Extract it but don't pass it as a header to the backend
	dynamicBaseURL := ""
	if headers != nil {
		// Check for CO base URL headers
		if val, ok := headers["X-Co-Base-Url"]; ok {
			dynamicBaseURL = val
			delete(headers, "X-Co-Base-Url")
		} else if val, ok := headers["X-CO-BASE-URL"]; ok {
			dynamicBaseURL = val
			delete(headers, "X-CO-BASE-URL")
		} else if val, ok := headers["X-Ctix-Base-Url"]; ok {
			dynamicBaseURL = val
			delete(headers, "X-Ctix-Base-Url")
		} else if val, ok := headers["X-CTIX-BASE-URL"]; ok {
			dynamicBaseURL = val
			delete(headers, "X-CTIX-BASE-URL")
		}

		// If a dynamic base URL is provided, use it
		if dynamicBaseURL != "" {
			baseURL = dynamicBaseURL
		}
	}

	request_url := baseURL + endpoint
	r := a.Client.R()

	var resp *resty.Response
	var err error

	r.SetQueryParams(queryParams)

	// Forward all client headers to the backend
	if headers != nil && len(headers) > 0 {
		r.SetHeaders(headers)
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
