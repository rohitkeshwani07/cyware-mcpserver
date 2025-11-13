#!/usr/bin/env python3
"""
Test client to verify SSE header forwarding
"""
import requests
import json
import time
import re

# Configuration
MCP_SERVER = "http://localhost:8000"
CUSTOM_TOKEN = "test-dynamic-token-12345"
CUSTOM_BASE_URL = "http://localhost:9090"

print("=== Testing SSE Header Forwarding ===\n")

# Step 1: Establish SSE connection
print("Step 1: Establishing SSE connection with custom headers...")
headers = {
    "Accept": "text/event-stream",
    "Authorization": f"Bearer {CUSTOM_TOKEN}",
    "X-Ctix-Base-Url": CUSTOM_BASE_URL,
    "X-Co-Base-Url": CUSTOM_BASE_URL,
    "X-Custom-Test-Header": "custom-value-123"
}

try:
    sse_response = requests.get(f"{MCP_SERVER}/sse", headers=headers, stream=True, timeout=5)
    
    # Extract session ID from the endpoint event
    session_id = None
    for line in sse_response.iter_lines():
        if line:
            decoded = line.decode('utf-8')
            print(f"  Received: {decoded[:100]}")
            
            # Look for endpoint event with session ID
            if decoded.startswith("event: endpoint"):
                # Next line should have the data
                continue
            if decoded.startswith("data:") and "endpoint" in decoded:
                try:
                    data_json = json.loads(decoded[5:].strip())
                    if "endpoint" in data_json:
                        # Extract session ID from endpoint URL
                        match = re.search(r'/message\?sessionId=([a-f0-9-]+)', data_json["endpoint"])
                        if match:
                            session_id = match.group(1)
                            print(f"\n✓ Session established: {session_id}\n")
                            break
                except:
                    pass
    
    if not session_id:
        print("✗ Failed to extract session ID")
        exit(1)
    
    # Step 2: Make a tool call via the message endpoint
    print("Step 2: Calling tool via message endpoint with custom headers...")
    
    message_url = f"{MCP_SERVER}/message?sessionId={session_id}"
    message_payload = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/call",
        "params": {
            "name": "logged-in-user-details",
            "arguments": {}
        }
    }
    
    # Include custom headers in the message request too
    message_headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {CUSTOM_TOKEN}",
        "X-Ctix-Base-Url": CUSTOM_BASE_URL,
        "X-Co-Base-Url": CUSTOM_BASE_URL,
        "X-Custom-Test-Header": "custom-value-123"
    }
    
    print(f"  Sending request to {message_url}")
    print(f"  With headers: {json.dumps(message_headers, indent=2)}")
    
    tool_response = requests.post(message_url, json=message_payload, headers=message_headers, timeout=10)
    
    print(f"\n  Response status: {tool_response.status_code}")
    print(f"  Response: {tool_response.text[:500]}")
    
    # Give proxy time to log
    time.sleep(1)
    
    print("\n✓ Test request sent. Check proxy.log for forwarded headers.")
    print("\nExpected headers in proxy.log:")
    print(f"  - Authorization: Bearer {CUSTOM_TOKEN} or CYW {CUSTOM_TOKEN}")
    print(f"  - X-Ctix-Base-Url: {CUSTOM_BASE_URL}")
    print(f"  - X-Co-Base-Url: {CUSTOM_BASE_URL}")
    print(f"  - X-Custom-Test-Header: custom-value-123")
    
except requests.exceptions.Timeout:
    print("✗ Request timed out")
except Exception as e:
    print(f"✗ Error: {e}")
    import traceback
    traceback.print_exc()
