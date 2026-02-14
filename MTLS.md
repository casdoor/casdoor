# mTLS (Mutual TLS) Client Authentication

This document describes the mTLS (Mutual TLS) client authentication support in Casdoor, implementing [RFC 8705 - OAuth 2.0 Mutual-TLS Client Authentication and Certificate-Bound Access Tokens](https://datatracker.ietf.org/doc/html/rfc8705).

## Overview

mTLS provides strong cryptographic client authentication using X.509 certificates as an alternative to traditional client secret-based authentication. This feature enables:

- **Strong Authentication**: Certificate-based client identity at the transport level
- **Certificate-Bound Tokens**: Proof-of-possession for access tokens (RFC 8705)
- **Financial-Grade Security**: Compliance with FAPI and other high-security requirements
- **PKI Integration**: Support for existing Public Key Infrastructure
- **Self-Signed Support**: Flexibility for development and internal services

## Features

### Supported Authentication Methods

1. **`tls_client_auth`**: PKI-based client authentication
   - Validates certificate chain against trusted CAs
   - Checks certificate expiration
   - Verifies against allowed issuers (optional)

2. **`self_signed_tls_client_auth`**: Self-signed certificate authentication
   - Accepts self-signed certificates
   - Validates certificate expiration
   - Suitable for development and internal services

### Certificate-Bound Access Tokens

When mTLS is enabled, access tokens can be bound to client certificates:
- Token fingerprint is stored with the access token
- Token usage requires presentation of the same certificate
- Provides proof-of-possession security

## Configuration

### Application Settings

Add the following fields to your Application configuration:

```go
{
  "enableMtls": true,                          // Enable mTLS for this application
  "mtlsAuthMethod": "tls_client_auth",         // Authentication method
  "allowedClientCertIssuers": [                // Optional: restrict certificate issuers
    "CN=My CA,O=My Organization,C=US"
  ]
}
```

### Server Configuration

To enable mTLS at the server level, configure TLS to request client certificates:

```go
import (
	"crypto/tls"
	"github.com/casdoor/casdoor/object"
)

// Get TLS configuration for mTLS
tlsConfig := object.GetTLSConfig()

// Use with your HTTPS server
server := &http.Server{
	TLSConfig: tlsConfig,
	// ... other configuration
}
```

The default TLS configuration:
- Requests client certificates (`tls.RequestClientCert`)
- Requires TLS 1.2 or higher
- Allows connections without certificates (mTLS is per-application)

## Usage Examples

### 1. Client Credentials Flow with mTLS

```bash
# Traditional client_secret authentication
curl -X POST https://your-casdoor/api/login/oauth/access_token \
  -d "grant_type=client_credentials" \
  -d "client_id=your-client-id" \
  -d "client_secret=your-client-secret" \
  -d "scope=read write"

# mTLS authentication (with client certificate)
curl -X POST https://your-casdoor/api/login/oauth/access_token \
  --cert client.crt \
  --key client.key \
  -d "grant_type=client_credentials" \
  -d "client_id=your-client-id" \
  -d "scope=read write"
```

### 2. Using Certificate-Bound Tokens

Once you have a certificate-bound token, you must present the same certificate when using it:

```bash
# Use the token with the same certificate
curl https://your-casdoor/api/some-endpoint \
  --cert client.crt \
  --key client.key \
  -H "Authorization: Bearer your-access-token"
```

### 3. Auto Sign-in with mTLS

mTLS authentication works with AutoSigninFilter for API access:

```bash
# Access protected API with mTLS
curl https://your-casdoor/api/protected-resource \
  --cert client.crt \
  --key client.key \
  -d "clientId=your-client-id"
```

## Certificate Requirements

### For `tls_client_auth` Method

- Valid X.509 certificate
- Not expired (within NotBefore and NotAfter dates)
- Extended Key Usage must include Client Authentication
- Certificate chain must be valid (if not self-signed)
- Issuer must match allowed issuers (if configured)

### For `self_signed_tls_client_auth` Method

- Valid X.509 certificate
- Not expired (within NotBefore and NotAfter dates)
- Can be self-signed

## Security Considerations

### Defense in Depth

Even with mTLS enabled, client secrets are still validated when configured. This provides:
- Protection against certificate compromise
- Backward compatibility
- Flexibility in security models

To use mTLS without client secret, leave the client secret empty in the application configuration.

### Certificate Validation

The system performs the following validations:

1. **Certificate Presence**: Ensures client certificate is provided
2. **Expiration Check**: Validates NotBefore and NotAfter dates
3. **Issuer Validation**: Checks against allowed issuers (if configured)
4. **Fingerprint Binding**: Validates token-certificate binding

### Best Practices

1. **Use PKI Certificates**: For production, use certificates from a trusted CA
2. **Restrict Issuers**: Configure `allowedClientCertIssuers` to limit trusted CAs
3. **Rotate Certificates**: Implement certificate rotation before expiration
4. **Monitor Certificate Expiration**: Set up alerts for expiring certificates
5. **Use Strong Keys**: Use at least 2048-bit RSA or 256-bit ECDSA keys

## API Endpoints

### OAuth Token Endpoint

**POST** `/api/login/oauth/access_token`

Supports mTLS authentication for:
- `client_credentials` grant type
- Other grant types (with client certificate binding)

### Token Introspection

**POST** `/api/login/oauth/introspect`

Validates certificate-bound tokens:
- Checks if token requires certificate binding
- Validates presented certificate matches token fingerprint
- Returns `active: false` if certificate doesn't match

## Integration Examples

### Kubernetes Service Mesh

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: client-cert
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-cert>
  tls.key: <base64-encoded-key>
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-service
spec:
  template:
    spec:
      containers:
      - name: app
        volumeMounts:
        - name: client-cert
          mountPath: /etc/ssl/client
      volumes:
      - name: client-cert
        secret:
          secretName: client-cert
```

### Go Client Example

```go
package main

import (
	"crypto/tls"
	"net/http"
)

func main() {
	// Load client certificate
	cert, err := tls.LoadX509KeyPair("client.crt", "client.key")
	if err != nil {
		panic(err)
	}

	// Create HTTPS client with mTLS
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	// Make authenticated request
	resp, err := client.Get("https://your-casdoor/api/endpoint")
	// ... handle response
}
```

### Node.js Client Example

```javascript
const https = require('https');
const fs = require('fs');

const options = {
  hostname: 'your-casdoor',
  port: 443,
  path: '/api/endpoint',
  method: 'GET',
  cert: fs.readFileSync('client.crt'),
  key: fs.readFileSync('client.key'),
};

const req = https.request(options, (res) => {
  // Handle response
});

req.end();
```

## Troubleshooting

### Common Issues

#### 1. "Certificate-bound token requires client certificate"

**Cause**: Token was issued with certificate binding, but no certificate was presented.

**Solution**: Present the same client certificate used when obtaining the token.

#### 2. "Client certificate does not match token binding"

**Cause**: A different certificate was presented than the one used to obtain the token.

**Solution**: Use the exact same certificate for all requests with the token.

#### 3. "mTLS certificate validation failed"

**Cause**: Certificate validation failed (expired, wrong issuer, etc.).

**Solution**: 
- Check certificate expiration dates
- Verify certificate issuer matches allowed issuers
- Ensure certificate is properly formatted

#### 4. "client_secret is invalid"

**Cause**: Client secret validation failed even with mTLS.

**Solution**: 
- Provide correct client secret (defense in depth)
- Or configure the application without a client secret

## Migration Guide

### From client_secret to mTLS

1. **Generate Client Certificates**
   ```bash
   # Generate private key
   openssl genrsa -out client.key 2048
   
   # Generate certificate signing request
   openssl req -new -key client.key -out client.csr
   
   # Generate self-signed certificate (for testing)
   openssl x509 -req -days 365 -in client.csr -signkey client.key -out client.crt
   ```

2. **Enable mTLS on Application**
   - Set `enableMtls: true`
   - Choose `mtlsAuthMethod`: `tls_client_auth` or `self_signed_tls_client_auth`
   - Optionally configure `allowedClientCertIssuers`

3. **Update Clients**
   - Configure HTTP client to present client certificate
   - Update token requests to use certificate authentication
   - Test with development environment first

4. **Monitor and Rollback Plan**
   - Monitor authentication failures
   - Keep client secrets as fallback during transition
   - Remove client secrets only after successful migration

## References

- [RFC 8705 - OAuth 2.0 Mutual-TLS Client Authentication and Certificate-Bound Access Tokens](https://datatracker.ietf.org/doc/html/rfc8705)
- [RFC 6749 - OAuth 2.0 Authorization Framework](https://datatracker.ietf.org/doc/html/rfc6749)
- [RFC 5280 - X.509 Public Key Infrastructure Certificate](https://datatracker.ietf.org/doc/html/rfc5280)

## Support

For issues or questions about mTLS support:
- Check the troubleshooting section above
- Review the test files in `object/mtls_test.go` for examples
- Open an issue on the Casdoor GitHub repository
