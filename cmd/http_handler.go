package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/cyware-labs/cyware-mcpserver/common"
	"github.com/mark3labs/mcp-go/server"
)

// HTTPHandler creates an HTTP handler for MCP tool calls
// It extracts headers from the request and injects them into the context
// before passing the request to the MCP server
func HTTPHandler(mcpServer *server.MCPServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract all headers from the request
		headers := common.ExtractHeadersFromHTTPRequest(r)
		
		// Log request (can be disabled in production)
		log.Printf("MCP HTTP request from %s to %s", r.RemoteAddr, r.URL.Path)

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()


		// Create context with headers
		ctx := common.InjectHeadersIntoContext(context.Background(), headers)

		// Handle the message through the MCP server
		response := mcpServer.HandleMessage(ctx, json.RawMessage(body))

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		responseBytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("Failed to marshal response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Write(responseBytes)
	}
}
