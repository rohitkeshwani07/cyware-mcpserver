package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

// Simple HTTP proxy/echo server to capture and display incoming requests
func main() {
	port := "9090"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("\n=== Incoming Request ===")
		log.Printf("Method: %s", r.Method)
		log.Printf("URL: %s", r.URL.String())
		log.Printf("Path: %s", r.URL.Path)
		
		log.Printf("\n--- Headers ---")
		for name, values := range r.Header {
			for _, value := range values {
				log.Printf("%s: %s", name, value)
			}
		}
		
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			log.Printf("\n--- Body ---")
			log.Printf("%s", string(body))
		}
		
		log.Printf("======================\n")
		
		// Return a mock successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		response := map[string]interface{}{
			"message": "Proxy received request successfully",
			"headers_received": r.Header,
			"path": r.URL.Path,
			"method": r.Method,
		}
		
		json.NewEncoder(w).Encode(response)
	})

	log.Printf("Test proxy server listening on http://localhost:%s", port)
	log.Printf("All incoming requests will be logged to console")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
