# Casdoor MCP Server

## Overview

Casdoor now provides a standard Model Context Protocol (MCP) server implementation. MCP is a protocol that allows AI assistants and other clients to interact with Casdoor's identity and access management capabilities through a standardized interface.

## What is MCP?

The Model Context Protocol (MCP) is an open protocol that standardizes how applications provide context to AI models. It enables AI assistants to:

- Access data from various sources
- Invoke tools and functions
- Use predefined prompts

## Endpoint

The MCP server is available at:

```
POST /mcp
```

## Authentication

The MCP endpoint requires admin authentication. You must be authenticated as an administrator to use the MCP server.

## Protocol

The MCP server implements JSON-RPC 2.0 protocol. All requests and responses follow the JSON-RPC 2.0 specification.

### Request Format

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "method_name",
  "params": {
    "param1": "value1"
  }
}
```

### Response Format

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    // method-specific result
  }
}
```

## Available Methods

### 1. Initialize

Initializes the MCP connection and retrieves server capabilities.

**Method:** `initialize`

**Example Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "clientInfo": {
      "name": "my-client",
      "version": "1.0.0"
    }
  }
}
```

**Example Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {
        "listChanged": false
      },
      "resources": {
        "subscribe": false,
        "listChanged": false
      },
      "prompts": {
        "listChanged": false
      }
    },
    "serverInfo": {
      "name": "casdoor-mcp-server",
      "version": "1.0.0"
    }
  }
}
```

### 2. List Tools

Lists all available tools that can be invoked.

**Method:** `tools/list`

**Example Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list",
  "params": {}
}
```

**Example Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "get_user",
        "description": "Retrieve a user by their user ID (format: owner/username)",
        "inputSchema": {
          "type": "object",
          "properties": {
            "userId": {
              "type": "string",
              "description": "User ID in format owner/username"
            }
          },
          "required": ["userId"]
        }
      }
      // ... more tools
    ]
  }
}
```

### 3. Call Tool

Invokes a specific tool with the provided arguments.

**Method:** `tools/call`

**Example Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "get_user",
    "arguments": {
      "userId": "built-in/admin"
    }
  }
}
```

**Example Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\n  \"owner\": \"built-in\",\n  \"name\": \"admin\",\n  \"email\": \"admin@example.com\"\n  ...\n}"
      }
    ],
    "isError": false
  }
}
```

## Available Tools

### User Management

#### get_user
Retrieve a user by their user ID.

**Parameters:**
- `userId` (string, required): User ID in format `owner/username`

**Example:**
```json
{
  "name": "get_user",
  "arguments": {
    "userId": "built-in/admin"
  }
}
```

#### list_users
List all users, optionally filtered by owner/organization.

**Parameters:**
- `owner` (string, optional): Owner/organization to filter users

**Example:**
```json
{
  "name": "list_users",
  "arguments": {
    "owner": "built-in"
  }
}
```

### Organization Management

#### get_organization
Retrieve an organization by ID.

**Parameters:**
- `organizationId` (string, required): Organization ID

**Example:**
```json
{
  "name": "get_organization",
  "arguments": {
    "organizationId": "built-in"
  }
}
```

#### list_organizations
List all organizations.

**Parameters:**
- `owner` (string, optional): Owner to filter organizations

**Example:**
```json
{
  "name": "list_organizations",
  "arguments": {}
}
```

### Application Management

#### get_application
Retrieve an application by ID.

**Parameters:**
- `applicationId` (string, required): Application ID in format `owner/appname`

**Example:**
```json
{
  "name": "get_application",
  "arguments": {
    "applicationId": "built-in/app-built-in"
  }
}
```

#### list_applications
List all applications, optionally filtered by owner.

**Parameters:**
- `owner` (string, optional): Owner to filter applications

**Example:**
```json
{
  "name": "list_applications",
  "arguments": {
    "owner": "built-in"
  }
}
```

### Role Management

#### get_role
Retrieve a role by ID.

**Parameters:**
- `roleId` (string, required): Role ID in format `owner/rolename`

**Example:**
```json
{
  "name": "get_role",
  "arguments": {
    "roleId": "built-in/admin"
  }
}
```

#### list_roles
List all roles, optionally filtered by owner.

**Parameters:**
- `owner` (string, optional): Owner to filter roles

**Example:**
```json
{
  "name": "list_roles",
  "arguments": {
    "owner": "built-in"
  }
}
```

## Error Handling

Errors are returned following JSON-RPC 2.0 error format:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32600,
    "message": "Invalid request"
  }
}
```

### Error Codes

- `-32700`: Parse error - Invalid JSON
- `-32600`: Invalid request - Missing required fields
- `-32601`: Method not found - Unknown method
- `-32602`: Invalid params - Invalid parameters
- `-32603`: Internal error - Server error

## Usage with AI Assistants

The MCP server can be used with AI assistants that support the Model Context Protocol. Configure your AI assistant to connect to the Casdoor MCP endpoint with appropriate authentication.

## Example: Complete Flow

1. **Initialize the connection:**
```bash
curl -X POST https://your-casdoor-instance.com/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05"
    }
  }'
```

2. **List available tools:**
```bash
curl -X POST https://your-casdoor-instance.com/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list",
    "params": {}
  }'
```

3. **Call a tool:**
```bash
curl -X POST https://your-casdoor-instance.com/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "list_users",
      "arguments": {
        "owner": "built-in"
      }
    }
  }'
```

## Example Client

An example Go client is provided in the `examples/` directory. You can use it to test the MCP server:

```bash
# Run the example client
cd mcp/examples
go run example_client.go http://localhost:8000/mcp YOUR_TOKEN
```

The example client demonstrates:
- Initializing the MCP connection
- Listing available tools
- Calling tools with parameters

## Security Considerations

- The MCP endpoint requires admin authentication
- All requests must be authenticated
- Ensure proper access control is configured for your Casdoor instance
- Use HTTPS in production environments
- Implement rate limiting if needed

## References

- [Model Context Protocol Specification](https://spec.modelcontextprotocol.io/)
- [Casdoor Documentation](https://casdoor.org)
