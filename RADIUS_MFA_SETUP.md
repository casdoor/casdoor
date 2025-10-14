# RADIUS MFA Setup Guide

This document describes how to configure an external RADIUS server as an MFA (Multi-Factor Authentication) provider in Casdoor.

## Overview

Casdoor now supports using external RADIUS servers for MFA verification, similar to how [Apereo CAS supports RADIUS authentication](https://apereo.github.io/cas/7.0.x/mfa/RADIUS-Authentication.html). This allows you to leverage your existing RADIUS infrastructure for second-factor authentication.

## Prerequisites

- A running RADIUS server (e.g., FreeRADIUS, Cisco ISE, Microsoft NPS)
- Network connectivity between Casdoor and your RADIUS server
- RADIUS shared secret for authentication
- User accounts configured in your RADIUS server

## Configuration Steps

### 1. Create a RADIUS Provider in Casdoor

First, you need to create a provider in Casdoor to represent your RADIUS server:

1. Navigate to **Providers** in the Casdoor admin panel
2. Click **Add** to create a new provider
3. Configure the provider with the following settings:

```json
{
  "owner": "built-in",
  "name": "radius-mfa-server",
  "createdTime": "2025-01-01T00:00:00Z",
  "displayName": "RADIUS MFA Server",
  "category": "MFA",
  "type": "RADIUS",
  "clientSecret": "your-radius-shared-secret",
  "host": "10.10.10.10",
  "port": 1812
}
```

**Field Descriptions:**
- `owner`: The organization that owns this provider (typically "built-in")
- `name`: A unique identifier for this provider
- `displayName`: Human-readable name shown in the UI
- `category`: Must be "MFA" for MFA providers
- `type`: Must be "RADIUS" for RADIUS MFA
- `clientSecret`: The shared secret configured on your RADIUS server
- `host`: IP address or hostname of your RADIUS server
- `port`: Authentication port (typically 1812 for RADIUS authentication)

### 2. Configure User for RADIUS MFA

To enable RADIUS MFA for a user:

1. Navigate to **Users** in the Casdoor admin panel
2. Select or create a user
3. Enable MFA settings for the user
4. Configure the following RADIUS-specific fields:
   - `radiusProvider`: Set to the provider ID (e.g., "built-in/radius-mfa-server")
   - `radiusUsername`: The username to use when authenticating to the RADIUS server
   - `radiusSecret`: A reference to the provider configuration
   - `preferredMfaType`: Set to "radius" to make RADIUS the preferred MFA method

### 3. Enable RADIUS MFA via API

You can also configure RADIUS MFA programmatically using the Casdoor API:

#### Initiate RADIUS MFA Setup

```http
POST /api/mfa/setup/initiate
Content-Type: application/x-www-form-urlencoded

owner=built-in&name=username&mfaType=radius
```

This will return MFA properties including recovery codes.

#### Verify RADIUS MFA Setup

Before enabling, verify that the RADIUS credentials work:

```http
POST /api/mfa/setup/verify
Content-Type: application/x-www-form-urlencoded

mfaType=radius&secret=built-in/radius-mfa-server&dest=radius-username&passcode=123456
```

Parameters:
- `mfaType`: Must be "radius"
- `secret`: The provider ID in format "owner/provider-name" (e.g., "built-in/radius-mfa-server")
- `dest`: The RADIUS username
- `passcode`: The RADIUS password/OTP to verify

#### Enable RADIUS MFA for User

After successful verification:

```http
POST /api/mfa/setup/enable
Content-Type: application/x-www-form-urlencoded

owner=built-in&name=username&mfaType=radius&secret=built-in/radius-mfa-server&dest=radius-username&recoveryCodes=recovery-code-from-initiate
```

Parameters:
- `owner`: Organization owner
- `name`: Username
- `mfaType`: Must be "radius"
- `secret`: The provider ID in format "owner/provider-name"
- `dest`: The RADIUS username
- `recoveryCodes`: Recovery code obtained from the initiate step

## Usage Flow

### Authentication with RADIUS MFA

When a user with RADIUS MFA enabled attempts to log in:

1. User enters username and password (first factor)
2. Casdoor validates the credentials
3. If successful and MFA is enabled, Casdoor prompts for a second factor
4. User enters their RADIUS password/token
5. Casdoor sends an Access-Request to the configured RADIUS server with:
   - Username from `radiusUsername` field
   - Password from user input
6. RADIUS server validates the credentials and responds with:
   - `Access-Accept` - MFA verification successful
   - `Access-Reject` - MFA verification failed
7. Casdoor completes or denies the authentication based on RADIUS response

### RADIUS Server Configuration

Your RADIUS server should be configured to:

1. Accept authentication requests from Casdoor's IP address
2. Use the same shared secret configured in the Casdoor provider
3. Have user accounts or integrate with your authentication backend (LDAP, AD, database, etc.)
4. Respond to PAP (Password Authentication Protocol) requests

### Example: FreeRADIUS Configuration

For FreeRADIUS, add Casdoor as a client in `clients.conf`:

```conf
client casdoor {
    ipaddr = 192.168.1.100  # Casdoor server IP
    secret = your-radius-shared-secret
    shortname = casdoor
    nas_type = other
}
```

Configure users in `users` file or integrate with LDAP/AD.

### Example: Cisco ISE Configuration

For Cisco ISE:

1. Navigate to **Administration > Network Resources > Network Devices**
2. Add Casdoor as a network device with the shared secret
3. Configure authentication policies for the user accounts

## Implementation Details

### MFA Type

The RADIUS MFA type constant is `"radius"` and is defined in `object/mfa.go`:

```go
const RadiusType = "radius"
```

### User Fields

The following fields are added to the User struct to support RADIUS MFA:

- `RadiusSecret`: Stores the provider ID reference
- `RadiusProvider`: Stores the full provider ID (owner/name)
- `RadiusUsername`: Username to use when authenticating to RADIUS server

### RADIUS Client Implementation

The RADIUS MFA implementation uses the `layeh.com/radius` library to communicate with RADIUS servers. It:

1. Creates an Access-Request packet
2. Sets the username and password attributes
3. Sends the request to the configured server
4. Processes the response (Accept/Reject)

See `object/mfa_radius.go` for the full implementation.

## Testing

To test the RADIUS MFA configuration:

1. Set up a test RADIUS server (e.g., FreeRADIUS with test users)
2. Create a RADIUS provider in Casdoor pointing to your test server
3. Configure a test user with RADIUS MFA
4. Attempt to authenticate with the test user
5. Verify that the second factor prompt appears
6. Enter valid RADIUS credentials to complete authentication

## Troubleshooting

### Common Issues

1. **Connection Timeout**: Verify network connectivity and firewall rules between Casdoor and RADIUS server
2. **Access-Reject**: Check that the username and credentials are correct in RADIUS server
3. **Shared Secret Mismatch**: Ensure the shared secret in Casdoor provider matches RADIUS server configuration
4. **Wrong Port**: RADIUS authentication typically uses port 1812, accounting uses 1813

### Debug Logging

Enable debug logging in Casdoor to see RADIUS communication details:

```bash
# In your Casdoor configuration
logLevel = "debug"
```

Check RADIUS server logs for incoming authentication requests and their status.

## Security Considerations

1. **Shared Secret**: Use a strong, unique shared secret for RADIUS communication
2. **Network Security**: Consider using IPsec or a VPN tunnel for RADIUS traffic
3. **Access Control**: Restrict RADIUS server to only accept requests from Casdoor's IP
4. **Credential Storage**: RADIUS credentials are stored encrypted in Casdoor database
5. **Timeout**: RADIUS requests timeout after 10 seconds to prevent hanging

## Comparison with CAS RADIUS MFA

This implementation is similar to [Apereo CAS RADIUS Authentication](https://apereo.github.io/cas/7.0.x/mfa/RADIUS-Authentication.html):

| Feature | CAS | Casdoor |
|---------|-----|---------|
| RADIUS Client | ✓ | ✓ |
| Shared Secret | ✓ | ✓ |
| Server Configuration | Properties file | Provider entity |
| User Mapping | Automatic | Configurable per user |
| Protocol Support | PAP | PAP |
| Timeout Configuration | ✓ | Fixed (10s) |

## References

- [RADIUS Protocol - RFC 2865](https://tools.ietf.org/html/rfc2865)
- [Apereo CAS RADIUS MFA](https://apereo.github.io/cas/7.0.x/mfa/RADIUS-Authentication.html)
- [Casdoor RADIUS Server Documentation](https://casdoor.org/docs/radius/overview)
- [FreeRADIUS Documentation](https://freeradius.org/documentation/)

## Support

For issues or questions about RADIUS MFA:

1. Check the [Casdoor GitHub Issues](https://github.com/casdoor/casdoor/issues)
2. Join the [Casdoor Community](https://github.com/casdoor/casdoor/discussions)
3. Review RADIUS server logs for authentication failures
