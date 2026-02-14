# mTLS (Mutual TLS) Client Authentication

This document describes how to configure and use mTLS (Mutual TLS) client authentication in Casdoor, following [RFC 8705](https://datatracker.ietf.org/doc/html/rfc8705).

## Overview

mTLS client authentication allows OAuth clients to authenticate using X.509 certificates instead of client secrets. This provides:

- **Strong cryptographic authentication** using certificate-based client identity
- **Enhanced security** - certificates are harder to compromise than shared secrets
- **Compliance** with financial-grade security requirements (FAPI)
- **Better service-to-service security** in Kubernetes and service mesh environments

## Configuration

### 1. Upload Client Certificate

First, upload the client certificate that will be used for authentication:

1. Navigate to the **Certificates** section in Casdoor admin panel
2. Create a new certificate or upload an existing one
3. Ensure the certificate is in PEM format
4. Note the certificate name for the next step

### 2. Configure Application for mTLS

To enable mTLS for an OAuth application:

1. Navigate to the **Applications** section
2. Select or create the application you want to configure
3. Set the following fields:
   - **Enable Client Cert**: Set to `true` to enable mTLS authentication
   - **Client Cert**: Select the certificate uploaded in step 1

### 3. TLS Server Configuration

Ensure your Casdoor server is configured to request client certificates:

```go
tlsConfig := &tls.Config{
    ClientAuth: tls.RequestClientCert, // or tls.RequireAnyClientCert
    // ... other TLS configuration
}
```

## Authentication Flow

### RFC 8705 tls_client_auth Method

When mTLS is enabled, the authentication flow works as follows:

1. **Client connects with TLS** and presents its X.509 certificate
2. **Casdoor receives the client certificate** from the TLS handshake
3. **Client sends clientId** via HTTP Basic Auth or query parameters
4. **Casdoor validates**:
   - The application has mTLS enabled
   - A client certificate is configured for the application
   - The presented certificate matches the stored certificate
   - The certificate is not expired
5. **Authentication succeeds** if all checks pass

### Fallback to Client Secret

If mTLS validation fails or is not configured:
- The system falls back to traditional client secret authentication
- This ensures backward compatibility with existing clients

## Example Usage

### Using mTLS with cURL

```bash
# Authenticate using client certificate
curl -X POST https://your-casdoor-instance/api/your-endpoint \
  --cert client-cert.pem \
  --key client-key.pem \
  --user "your-client-id:" \
  -d "grant_type=client_credentials"
```

### Using mTLS with Go

```go
cert, err := tls.LoadX509KeyPair("client-cert.pem", "client-key.pem")
if err != nil {
    log.Fatal(err)
}

client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            Certificates: []tls.Certificate{cert},
        },
    },
}

req, _ := http.NewRequest("POST", "https://your-casdoor-instance/api/your-endpoint", nil)
req.SetBasicAuth("your-client-id", "")
resp, err := client.Do(req)
```

## Security Considerations

### Certificate Management

- **Certificate Rotation**: Regularly rotate client certificates
- **Secure Storage**: Store private keys securely
- **Certificate Revocation**: Have a process to revoke compromised certificates

### Validation

The system performs the following validations:

1. **Certificate Expiration**: Checks `NotBefore` and `NotAfter` dates
2. **Certificate Matching**: Compares the presented certificate with the stored certificate
3. **Certificate Chain**: Uses the first certificate in the chain (client certificate)

### Best Practices

- Use certificates from a trusted CA when possible
- Set appropriate certificate expiration periods
- Monitor certificate expiration and set up alerts
- Use strong key sizes (2048-bit RSA minimum, 256-bit ECDSA recommended)
- Implement proper certificate lifecycle management

## Troubleshooting

### Common Issues

**Authentication fails with "certificate validation failed"**
- Verify the certificate is not expired
- Ensure the client is presenting the correct certificate
- Check that the certificate matches the one configured in Casdoor

**"mTLS is enabled but no client certificate is configured"**
- Ensure you've set the `ClientCert` field in the application configuration
- Verify the certificate exists in Casdoor's certificate store

**No client certificate received**
- Verify the TLS server is configured to request client certificates
- Check that the client is sending its certificate in the TLS handshake
- Ensure the application's `EnableClientCert` flag is set to `true`

## API Fields

### Application Model

| Field | Type | Description |
|-------|------|-------------|
| `enableClientCert` | boolean | Enables/disables mTLS for the application |
| `clientCert` | string | Name of the certificate to validate against (format: "owner/cert-name") |

### Certificate Model

Client certificates should be configured with:
- **Type**: Can be any type, typically "x509" or "client-cert"
- **Certificate**: PEM-encoded X.509 certificate
- **Scope**: Optional scope for certificate usage

## Related Standards

- [RFC 8705: OAuth 2.0 Mutual-TLS Client Authentication and Certificate-Bound Access Tokens](https://datatracker.ietf.org/doc/html/rfc8705)
- [RFC 5246: The Transport Layer Security (TLS) Protocol Version 1.2](https://datatracker.ietf.org/doc/html/rfc5246)
- [RFC 8446: The Transport Layer Security (TLS) Protocol Version 1.3](https://datatracker.ietf.org/doc/html/rfc8446)
