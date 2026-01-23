# MySQL mTLS (Mutual TLS) Support

Casdoor now supports MySQL mutual TLS authentication, allowing secure two-way authentication between the application and the MySQL database server using client certificates.

## Overview

mTLS (mutual TLS) provides enhanced security for database connections by:
- Verifying the database server's identity (server authentication)
- Verifying the client's identity using certificates (client authentication)
- Encrypting all traffic between the client and server

This is particularly important in high-security environments such as:
- Zero Trust architectures
- Compliance-driven environments (PCI-DSS, HIPAA, etc.)
- Multi-tenant cloud environments
- Enterprise security policies requiring certificate-based authentication

## Configuration

To enable MySQL mTLS, add the following optional configuration fields to your `conf/app.conf` file:

```ini
# Database driver (must be mysql)
driverName = mysql
dataSourceName = root:password@tcp(localhost:3306)/
dbName = casdoor

# MySQL TLS certificate paths (all optional)
dbCaCert = /path/to/ca.pem
dbClientCert = /path/to/client-cert.pem
dbClientKey = /path/to/client-key.pem
```

### Configuration Fields

- **`dbCaCert`** (optional): Path to the CA (Certificate Authority) certificate file in PEM format. This is used to verify the MySQL server's certificate.

- **`dbClientCert`** (optional): Path to the client certificate file in PEM format. This identifies the Casdoor application to the MySQL server.

- **`dbClientKey`** (optional): Path to the client private key file in PEM format. This must be provided together with `dbClientCert`.

### Certificate Requirements

1. **CA Certificate** (`dbCaCert`):
   - Must be in PEM format
   - Should contain the root CA that signed the MySQL server's certificate
   - Can be used alone for server-side verification only

2. **Client Certificate and Key** (`dbClientCert` and `dbClientKey`):
   - Both must be provided together (you cannot provide only one)
   - Must be in PEM format
   - Client certificate must be signed by a CA trusted by the MySQL server
   - Private key must match the client certificate

## Examples

### Example 1: CA Certificate Only (Server Verification)

This configuration verifies the MySQL server's identity but doesn't provide client authentication:

```ini
driverName = mysql
dataSourceName = root:password@tcp(mysql.example.com:3306)/
dbName = casdoor
dbCaCert = /etc/casdoor/certs/ca.pem
dbClientCert =
dbClientKey =
```

### Example 2: Full mTLS (Two-Way Authentication)

This configuration provides both server and client authentication:

```ini
driverName = mysql
dataSourceName = root:password@tcp(mysql.example.com:3306)/
dbName = casdoor
dbCaCert = /etc/casdoor/certs/ca.pem
dbClientCert = /etc/casdoor/certs/client-cert.pem
dbClientKey = /etc/casdoor/certs/client-key.pem
```

### Example 3: Client Certificates Only

You can also provide only client certificates if the server's CA is already trusted by the system:

```ini
driverName = mysql
dataSourceName = root:password@tcp(mysql.example.com:3306)/
dbName = casdoor
dbCaCert =
dbClientCert = /etc/casdoor/certs/client-cert.pem
dbClientKey = /etc/casdoor/certs/client-key.pem
```

## Environment Variables

All configuration values can also be set via environment variables, which take precedence over values in `app.conf`:

```bash
export driverName=mysql
export dataSourceName="root:password@tcp(localhost:3306)/"
export dbName=casdoor
export dbCaCert=/path/to/ca.pem
export dbClientCert=/path/to/client-cert.pem
export dbClientKey=/path/to/client-key.pem
```

## Troubleshooting

### Error: "failed to read CA certificate"

- Verify the file path is correct
- Ensure the file is readable by the Casdoor process
- Check that the file contains a valid PEM-encoded certificate

### Error: "failed to parse CA certificate"

- Ensure the file is in PEM format (contains `-----BEGIN CERTIFICATE-----`)
- Verify the certificate is not corrupted

### Error: "both dbClientCert and dbClientKey must be provided together"

- You must provide both the client certificate and key, or neither
- Set both configuration values or leave both empty

### Error: "failed to load client certificate/key"

- Verify both files exist and are readable
- Ensure the private key matches the certificate
- Check that both files are in PEM format

### Error: "x509: certificate signed by unknown authority"

- The CA certificate doesn't match the server's certificate
- Ensure you're using the correct CA certificate that signed the MySQL server's certificate

## Security Best Practices

1. **Protect Private Keys**: Ensure client private keys are only readable by the Casdoor process
   ```bash
   chmod 600 /path/to/client-key.pem
   chown casdoor:casdoor /path/to/client-key.pem
   ```

2. **Use Strong Certificates**: Use at least 2048-bit RSA keys or equivalent

3. **Rotate Certificates**: Implement a certificate rotation policy

4. **Monitor Certificate Expiration**: Set up alerts for certificate expiration

## Backward Compatibility

- If no certificate paths are configured, Casdoor behaves exactly as before
- Existing configurations without TLS will continue to work unchanged
- This feature is fully optional and backward compatible
