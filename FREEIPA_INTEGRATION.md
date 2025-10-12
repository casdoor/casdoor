# FreeIPA-Compatible API Integration

This document describes how to use Casdoor's FreeIPA-compatible JSON-RPC API for Linux machine authentication, similar to FreeIPA.

## Overview

Casdoor now provides FreeIPA-compatible JSON-RPC API endpoints that allow Linux machines to authenticate users via SSO, similar to how FreeIPA works. This enables integration with SSSD (System Security Services Daemon) and PAM (Pluggable Authentication Modules) on Linux systems.

## API Endpoints

### 1. JSON-RPC Endpoint
**URL:** `/ipa/json`
**Method:** POST
**Description:** Main JSON-RPC endpoint for FreeIPA-compatible operations

### 2. Session-based JSON-RPC Endpoint
**URL:** `/ipa/session/json`
**Method:** POST
**Description:** Authenticated JSON-RPC endpoint requiring session

### 3. Session Login
**URL:** `/ipa/session/login_password`
**Method:** POST
**Description:** Authenticate and create a session

**Parameters:**
- `username` (string): Username
- `password` (string): Password
- `organization` (string, optional): Organization name (default: "built-in")

### 4. Session Logout
**URL:** `/ipa/session/logout`
**Method:** POST
**Description:** Destroy the current session

## Supported JSON-RPC Methods

### ping
Check server availability and version.

**Request:**
```json
{
  "method": "ping",
  "params": [[], {}],
  "id": 1
}
```

**Response:**
```json
{
  "result": {
    "summary": "IPA server version 4.9.0. API version 2.245"
  },
  "error": null,
  "id": 1
}
```

### user_show
Get information about a specific user.

**Request:**
```json
{
  "method": "user_show",
  "params": [
    ["username"],
    {"organization": "built-in"}
  ],
  "id": 1
}
```

**Response:**
```json
{
  "result": {
    "result": {
      "uid": ["username"],
      "uidnumber": ["1234"],
      "gidnumber": ["1234"],
      "cn": ["Display Name"],
      "displayname": ["Display Name"],
      "mail": ["user@example.com"],
      "homedirectory": ["/home/username"],
      "loginshell": ["/bin/bash"],
      "memberof": ["group1", "group2"]
    },
    "value": "username"
  },
  "error": null,
  "id": 1
}
```

### user_find
Search for users.

**Request:**
```json
{
  "method": "user_find",
  "params": [
    ["search_term"],
    {"organization": "built-in"}
  ],
  "id": 1
}
```

**Response:**
```json
{
  "result": {
    "result": [
      {
        "uid": ["user1"],
        "uidnumber": ["1234"],
        "gidnumber": ["1234"],
        "cn": ["User One"],
        "displayname": ["User One"],
        "mail": ["user1@example.com"],
        "homedirectory": ["/home/user1"],
        "loginshell": ["/bin/bash"]
      }
    ],
    "count": 1,
    "truncated": false
  },
  "error": null,
  "id": 1
}
```

### group_show
Get information about a specific group.

**Request:**
```json
{
  "method": "group_show",
  "params": [
    ["groupname"],
    {"organization": "built-in"}
  ],
  "id": 1
}
```

**Response:**
```json
{
  "result": {
    "result": {
      "cn": ["groupname"],
      "gidnumber": ["5678"],
      "member": ["user1", "user2"]
    },
    "value": "groupname"
  },
  "error": null,
  "id": 1
}
```

## Integration with Linux Systems

### Using with SSSD

To integrate Casdoor with Linux systems using SSSD, you can configure SSSD to use the IPA provider pointing to your Casdoor server.

**Example SSSD Configuration** (`/etc/sssd/sssd.conf`):

```ini
[sssd]
config_file_version = 2
services = nss, pam
domains = casdoor

[domain/casdoor]
id_provider = ipa
auth_provider = ipa
ipa_server = casdoor.example.com
ipa_domain = example.com
ipa_hostname = casdoor.example.com
cache_credentials = True
krb5_store_password_if_offline = True
```

**Note:** Full SSSD integration may require additional configuration and Kerberos setup. The JSON-RPC API provides the foundation for user and group information queries.

### Using with PAM

For PAM-based authentication, you can use the REST API with a custom PAM module that calls the Casdoor authentication endpoints.

## Example: Testing with curl

### Login and create session:
```bash
curl -X POST http://localhost:8000/ipa/session/login_password \
  -d "username=admin" \
  -d "password=123" \
  -d "organization=built-in" \
  -c cookies.txt
```

### Get user information:
```bash
curl -X POST http://localhost:8000/ipa/session/json \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "method": "user_show",
    "params": [["admin"], {}],
    "id": 1
  }'
```

### Search users:
```bash
curl -X POST http://localhost:8000/ipa/session/json \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "method": "user_find",
    "params": [[""], {}],
    "id": 1
  }'
```

### Logout:
```bash
curl -X POST http://localhost:8000/ipa/session/logout \
  -b cookies.txt
```

## Security Considerations

1. **HTTPS Required:** Always use HTTPS in production to protect credentials
2. **Authentication:** Most operations require authentication via session or other methods
3. **Organization Scope:** Users and groups are scoped to organizations in Casdoor
4. **Access Control:** Ensure proper access controls are configured in your Casdoor application

## Limitations

This implementation provides basic FreeIPA JSON-RPC API compatibility for user and group operations. It does not include:

- Kerberos authentication (use existing authentication methods)
- DNS management
- Certificate management
- Host enrollment
- Other advanced FreeIPA features

For full Linux machine integration, combine this API with Casdoor's existing LDAP and RADIUS servers for comprehensive authentication and authorization.

## Additional Resources

- [FreeIPA JSON-RPC API Documentation](https://www.freeipa.org/page/JSON-RPC_API)
- [SSSD Documentation](https://sssd.io/)
- [Casdoor LDAP Server](./ldap/README.md) (if exists)
- [Casdoor RADIUS Server](./radius/README.md) (if exists)

## Support

For issues or questions about this integration, please open an issue on the Casdoor GitHub repository.
