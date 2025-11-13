package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/cyware-labs/cyware-mcpserver/applications/co"
	"github.com/cyware-labs/cyware-mcpserver/applications/ctix"
	"github.com/cyware-labs/cyware-mcpserver/applications/general"
	"github.com/cyware-labs/cyware-mcpserver/common"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	envPath := flag.String("config_path", "config.yaml", "Path to the .yaml file")
	flag.Parse()

	s := server.NewMCPServer(
		"CYWARE-MCP-SERVER",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
		server.WithInstructions(`
		# Cyware MCP Server 
		This server provides tools to access Cyware Products - CTIX(Cyware Threat Intel Exchange), CFTR, CSAP, CO platform functionalities and features.
		## ❗⚠️ Don't use tools where its mentioned in the tools description that it must be explicitly invoked.
		`),
	)

	cfg, err := common.Load(*envPath)

	if err != nil {
		log.Fatal(err)
	}

	ctix.Initialize(cfg, s)
	general.Initialize(s)
	co.Initialize(cfg, s)

	mcp_server_mode := cfg.Server.MCPMode
	if mcp_server_mode == "" {
		mcp_server_mode = "stdio"
	}

	switch mcp_server_mode {
	case "stdio":
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("Server error: %v\n", err)
		}
	case "sse":
		// Create SSE server with context function to inject headers
		contextFunc := func(ctx context.Context, r *http.Request) context.Context {
			// THIS SHOULD ALWAYS PRINT IF THE FUNCTION IS CALLED
			fmt.Println("!!! CONTEXT FUNCTION CALLED !!!")
			
			headers := common.ExtractHeadersFromHTTPRequest(r)
			
			// Debug logging to stdout
			fmt.Printf("\n[SSE Context] Extracting headers from request to %s\n", r.URL.Path)
			if auth, ok := headers["Authorization"]; ok {
				authPreview := auth
				if len(authPreview) > 50 {
					authPreview = authPreview[:50] + "..."
				}
				fmt.Printf("[SSE Context]   Found Authorization: %s\n", authPreview)
			}
			if ctixURL, ok := headers["X-Ctix-Base-Url"]; ok {
				fmt.Printf("[SSE Context]   Found X-Ctix-Base-Url: %s\n", ctixURL)
			}
			if coURL, ok := headers["X-Co-Base-Url"]; ok {
				fmt.Printf("[SSE Context]   Found X-Co-Base-Url: %s\n", coURL)
			}
			fmt.Printf("[SSE Context]   Total headers extracted: %d\n", len(headers))
			
			return common.InjectHeadersIntoContext(ctx, headers)
		}
		
		sseServer := server.NewSSEServer(s, server.WithSSEContextFunc(contextFunc))
		if err := sseServer.Start(":" + cfg.Server.Port); err != nil {
			log.Fatalf("Server error: %v", err)
		}
		log.Printf("MCP server listening on :%v", cfg.Server.Port)
	}

}
