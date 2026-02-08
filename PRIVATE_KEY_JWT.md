# OAuth 2.0 `private_key_jwt` Client Authentication

This document describes the implementation of OAuth 2.0 `private_key_jwt` client authentication in Casdoor, following RFC 7523 (JSON Web Token (JWT) Profile for OAuth 2.0 Client Authentication and Authorization Grants) and RFC 7521 (Assertion Framework for OAuth 2.0 Client Authentication and Authorization Grants).

## Overview

The `private_key_jwt` authentication method allows OAuth 2.0 clients to authenticate using asymmetric cryptography (public/private key pairs) instead of shared secrets. This provides several security benefits:

- **No shared secrets**: Eliminates the risk of secret leakage or interception
- **Certificate-based authentication**: Leverages existing PKI infrastructure
- **Non-repudiation**: Cryptographic proof of client identity
- **Better for M2M**: Ideal for machine-to-machine communication scenarios

## Supported Grant Types

The `private_key_jwt` authentication method is supported for the following OAuth 2.0 grant types:

1. **Authorization Code Grant** (`authorization_code`)
2. **Client Credentials Grant** (`client_credentials`)
3. **Refresh Token Grant** (`refresh_token`)
4. **Token Exchange Grant** (`urn:ietf:params:oauth:grant-type:token-exchange`)

## Configuration

### Application Setup

To enable `private_key_jwt` authentication for an application:

1. **Create a Certificate** in Casdoor with the client's public key
2. **Configure the Application**:
   - Set `tokenEndpointAuthMethod` to `"private_key_jwt"`
   - Associate the certificate with the application via the `cert` field

### Client Configuration

The OAuth 2.0 client must:

1. Generate an RSA or ECDSA key pair
2. Register the public key with Casdoor as a certificate
3. Use the private key to sign JWT assertions for authentication

## Token Endpoint Request

When using `private_key_jwt` authentication, clients must include the following parameters in the token request:

- `client_id`: The OAuth 2.0 client identifier
- `client_assertion_type`: Must be `"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"`
- `client_assertion`: A signed JWT containing authentication claims

**Do NOT include** `client_secret` when using `private_key_jwt`.

### Example Token Request

```http
POST /api/login/oauth/access_token HTTP/1.1
Host: your-casdoor-instance.com
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code
&code=AUTHORIZATION_CODE
&redirect_uri=https://client.example.com/callback
&client_id=your-client-id
&client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer
&client_assertion=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

## JWT Assertion Requirements

The `client_assertion` JWT must contain the following claims (per RFC 7523):

| Claim | Required | Description |
|-------|----------|-------------|
| `iss` (Issuer) | **Yes** | Must be the `client_id` |
| `sub` (Subject) | **Yes** | Must be the `client_id` |
| `aud` (Audience) | **Yes** | Must be the token endpoint URL or authorization server identifier |
| `exp` (Expiration Time) | **Yes** | Must be in the future (Unix timestamp) |
| `iat` (Issued At) | Recommended | When the JWT was created (Unix timestamp) |
| `jti` (JWT ID) | Recommended | Unique identifier for replay protection |
| `nbf` (Not Before) | Optional | When the JWT becomes valid (Unix timestamp) |

### Example JWT Payload

```json
{
  "iss": "your-client-id",
  "sub": "your-client-id",
  "aud": "https://your-casdoor-instance.com/api/login/oauth/access_token",
  "exp": 1735689600,
  "iat": 1735686000,
  "jti": "unique-jwt-id-12345"
}
```

### Signing the JWT

The JWT must be signed with the client's private key using one of the following algorithms:

- **RS256, RS384, RS512** (RSA with SHA-256/384/512)
- **PS256, PS384, PS512** (RSA-PSS with SHA-256/384/512)
- **ES256, ES384, ES512** (ECDSA with SHA-256/384/512)

Example using the `golang-jwt/jwt` library:

```go
import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

// Create the claims
claims := jwt.RegisteredClaims{
    Issuer:    "your-client-id",
    Subject:   "your-client-id",
    Audience:  []string{"https://your-casdoor-instance.com/api/login/oauth/access_token"},
    ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
    IssuedAt:  jwt.NewNumericDate(time.Now()),
    ID:        "unique-jti-12345",
}

// Create and sign the token
token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
assertion, err := token.SignedString(privateKey)
```

## Validation Process

When a client assertion is received, Casdoor performs the following validation steps:

1. **Parse and verify the JWT signature** using the application's registered public key
2. **Validate the issuer (`iss`)** matches the `client_id`
3. **Validate the subject (`sub`)** matches the `client_id`
4. **Validate the audience (`aud`)** matches the token endpoint or issuer identifier
5. **Validate the expiration time (`exp`)** is in the future
6. **Validate the not-before time (`nbf`)** if present
7. **Check the signing algorithm** is supported and matches the certificate type

## Backward Compatibility

The implementation maintains full backward compatibility:

- Applications without `tokenEndpointAuthMethod` configured continue to use `client_secret`
- Existing client secret authentication still works
- PKCE (RFC 7636) support is preserved for authorization code flow

## Security Considerations

### Best Practices

1. **Short JWT lifetimes**: Keep `exp` within 5 minutes to minimize replay attack window
2. **Unique JTI**: Use unique `jti` values for each assertion to enable replay protection
3. **Secure key storage**: Protect private keys using hardware security modules (HSMs) or secure enclaves
4. **Key rotation**: Periodically rotate key pairs and update certificates
5. **Audience validation**: Always set `aud` to the specific token endpoint URL

### Replay Attack Protection

While Casdoor validates JWT expiration times, production deployments should implement additional replay protection by:

1. Tracking used `jti` values within the validity window
2. Rejecting JWTs with previously seen `jti` values
3. Using Redis or another distributed cache for multi-instance deployments

## Implementation Files

- `object/client_auth.go` - Core authentication logic
- `object/client_auth_test.go` - Unit tests
- `controllers/token.go` - Token endpoint integration
- `object/token_oauth.go` - Grant-specific authentication
- `object/application.go` - Application model with `tokenEndpointAuthMethod` field

## References

- [RFC 7521: Assertion Framework for OAuth 2.0 Client Authentication](https://datatracker.ietf.org/doc/html/rfc7521)
- [RFC 7523: JWT Profile for OAuth 2.0 Client Authentication](https://datatracker.ietf.org/doc/html/rfc7523)
- [OAuth 2.0 private_key_jwt](https://oauth.net/private-key-jwt/)
- [OpenID Connect Core 1.0 - Section 9](https://openid.net/specs/openid-connect-core-1_0.html#ClientAuthentication)
