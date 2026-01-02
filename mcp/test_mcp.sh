#!/bin/bash

# Test script for Casdoor MCP server
# This script tests the basic MCP protocol compliance

BASE_URL="http://localhost:8000/mcp"

echo "=== Testing Casdoor MCP Server ==="
echo ""

# Test 1: Initialize
echo "Test 1: Initialize"
echo "Request:"
cat << 'EOF'
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05"
  }
}
EOF

echo ""
echo "Expected: Should return server info and capabilities"
echo ""

# Test 2: List Tools
echo "Test 2: List Tools"
echo "Request:"
cat << 'EOF'
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list",
  "params": {}
}
EOF

echo ""
echo "Expected: Should return list of available tools"
echo ""

# Test 3: Invalid method
echo "Test 3: Invalid Method"
echo "Request:"
cat << 'EOF'
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "invalid/method",
  "params": {}
}
EOF

echo ""
echo "Expected: Should return method not found error"
echo ""

# Test 4: Invalid JSON-RPC version
echo "Test 4: Invalid JSON-RPC Version"
echo "Request:"
cat << 'EOF'
{
  "jsonrpc": "1.0",
  "id": 4,
  "method": "initialize",
  "params": {}
}
EOF

echo ""
echo "Expected: Should return invalid request error"
echo ""

echo "=== Tests completed ==="
echo "Note: To actually run these tests, you need to:"
echo "1. Start the Casdoor server"
echo "2. Obtain admin authentication credentials"
echo "3. Use curl or another HTTP client to send these requests"
echo ""
echo "Example curl command:"
echo 'curl -X POST http://localhost:8000/mcp \'
echo '  -H "Content-Type: application/json" \'
echo '  -H "Authorization: Bearer YOUR_TOKEN" \'
echo '  -d '"'"'{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}'"'"
