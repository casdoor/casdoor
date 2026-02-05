# OAuth 2.0 Token Exchange (RFC 8693) Support

Casdoor now supports the OAuth 2.0 Token Exchange grant type as defined in [RFC 8693](https://www.rfc-editor.org/rfc/rfc8693.html).

## Overview

Token Exchange enables clients to exchange a valid security token (e.g., an access token, ID token, or JWT) for a new token tailored for a different audience or scope. This is particularly useful in microservices architectures where:

- A client (e.g., an API gateway) needs to forward requests to downstream services
- Services expect tokens with service-specific audiences
- Scope downscoping is required for principle of least privilege
- Token delegation and impersonation scenarios

## Usage

### Endpoint

```
POST /api/login/oauth/access_token
```

### Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `grant_type` | Yes | Must be `urn:ietf:params:oauth:grant-type:token-exchange` |
| `client_id` | Yes | The client identifier |
| `client_secret` | Yes | The client secret |
| `subject_token` | Yes | The security token being exchanged |
| `subject_token_type` | No | Token type identifier (defaults to `urn:ietf:params:oauth:token-type:access_token`) |
| `audience` | No | The intended audience for the new token |
| `scope` | No | Space-separated list of scopes (must be subset of subject token scopes) |

### Supported Token Types

- `urn:ietf:params:oauth:token-type:access_token` - Access token
- `urn:ietf:params:oauth:token-type:jwt` - JWT token
- `urn:ietf:params:oauth:token-type:id_token` - ID token

### Example Request

```bash
curl -X POST https://your-casdoor-instance/api/login/oauth/access_token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "subject_token=EXISTING_ACCESS_TOKEN" \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "scope=openid profile"
```

### Example Response

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 7200,
  "scope": "openid profile"
}
```

### Error Responses

| Error | Description |
|-------|-------------|
| `invalid_client` | Client authentication failed |
| `invalid_request` | Missing or invalid `subject_token` parameter |
| `invalid_grant` | Invalid or expired subject token |
| `invalid_scope` | Requested scope is not a subset of subject token scope |
| `unsupported_grant_type` | Token exchange not enabled for this application |

## Configuration

To enable token exchange for an application:

1. Navigate to your application settings in Casdoor
2. Add `urn:ietf:params:oauth:grant-type:token-exchange` to the allowed grant types
3. Save the application configuration

## Use Cases

### Microservices Token Delegation

```
┌─────────┐     ┌─────────────┐     ┌──────────┐     ┌──────────┐
│ Client  │────▶│ API Gateway │────▶│ Service A│────▶│ Service B│
└─────────┘     └─────────────┘     └──────────┘     └──────────┘
   Token            Exchange          Exchange
                    Token             Token
                   (Scope: A)       (Scope: B)
```

### Scope Downscoping

Exchange a token with broad permissions for one with limited scope:

```
Original Token: scope="read write delete"
Exchanged Token: scope="read"
```

## Security Considerations

1. **Client Authentication**: Always verify the client credentials before issuing a new token
2. **Scope Downscoping**: The new token's scope must be a subset of the original token's scope
3. **Token Validation**: Subject tokens are validated using the application's certificate
4. **User Status**: Forbidden users cannot exchange tokens
5. **Audience Restriction**: Use the `audience` parameter to restrict token usage to specific services

## OIDC Discovery

Token exchange support is advertised in the OIDC discovery document:

```json
{
  "grant_types_supported": [
    "authorization_code",
    "implicit",
    "password",
    "client_credentials",
    "refresh_token",
    "urn:ietf:params:oauth:grant-type:device_code",
    "urn:ietf:params:oauth:grant-type:token-exchange"
  ]
}
```

## References

- [RFC 8693 - OAuth 2.0 Token Exchange](https://www.rfc-editor.org/rfc/rfc8693.html)
- [OAuth 2.0 Specification](https://oauth.net/2/)
