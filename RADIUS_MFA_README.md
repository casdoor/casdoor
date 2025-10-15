# RADIUS MFA Support

This document provides a summary of the RADIUS MFA support feature added to Casdoor.

## Overview

Casdoor now supports using external RADIUS servers as Multi-Factor Authentication (MFA) providers. This allows organizations to leverage their existing RADIUS infrastructure (such as FreeRADIUS, Cisco ISE, or Microsoft NPS) for second-factor authentication.

## Key Features

- **RADIUS Client Support**: Casdoor can now act as a RADIUS client to authenticate against external RADIUS servers
- **MFA Integration**: RADIUS authentication is fully integrated with Casdoor's MFA framework
- **Provider-Based Configuration**: RADIUS servers are configured as providers, making them easy to manage
- **Per-User Configuration**: Each user can have their own RADIUS username and provider assignment
- **Standard RADIUS Protocol**: Uses standard RADIUS Access-Request/Accept/Reject flow (RFC 2865)
- **Flexible Deployment**: Works alongside existing MFA methods (SMS, Email, TOTP)

## What Changed

### New Files

1. **object/mfa_radius.go**: Core RADIUS MFA implementation
   - `RadiusMfa` struct implementing the `MfaInterface`
   - RADIUS client functionality using `layeh.com/radius` library
   - Provider-based configuration loading

2. **object/mfa_radius_test.go**: Unit tests for RADIUS MFA
   - Tests for utility creation
   - Tests for MFA type integration
   - Tests for MFA props retrieval

3. **RADIUS_MFA_SETUP.md**: Comprehensive setup guide
   - Configuration instructions
   - API usage examples
   - Troubleshooting guide
   - Security considerations

4. **examples/radius_mfa_provider.json**: Example provider configuration

### Modified Files

1. **object/mfa.go**:
   - Added `RadiusType` constant ("radius")
   - Updated `GetMfaUtil()` to support RADIUS type
   - Updated `GetAllMfaProps()` to include RADIUS
   - Updated `DisabledMultiFactorAuth()` to clear RADIUS fields
   - Updated `GetMfaProps()` to handle RADIUS configuration

2. **object/user.go**:
   - Added `RadiusSecret` field for provider reference
   - Added `RadiusProvider` field for full provider ID
   - Added `RadiusUsername` field for RADIUS authentication

3. **radius/server.go**:
   - Updated `handleAccessRequest()` to use preferred MFA type instead of hardcoding TOTP
   - Now works with any MFA type including RADIUS

4. **controllers/mfa.go**:
   - Updated `MfaSetupVerify()` to handle RADIUS parameters
   - Updated `MfaSetupEnable()` to configure RADIUS MFA for users

## How It Works

### Provider Configuration

RADIUS servers are configured as providers with:
- **Category**: "MFA"
- **Type**: "RADIUS"
- **ClientSecret**: Shared secret for RADIUS communication
- **Host**: RADIUS server IP/hostname
- **Port**: RADIUS authentication port (typically 1812)

### User Configuration

Users with RADIUS MFA have:
- **RadiusProvider**: Reference to the RADIUS provider (e.g., "built-in/radius-mfa-server")
- **RadiusUsername**: Username to use when authenticating to RADIUS
- **PreferredMfaType**: Set to "radius" to make it the default MFA method

### Authentication Flow

1. User logs in with username/password (first factor)
2. Casdoor prompts for second factor
3. User enters RADIUS password/OTP
4. Casdoor creates RADIUS Access-Request packet
5. Sends request to configured RADIUS server
6. RADIUS server validates and responds
7. Casdoor completes authentication based on RADIUS response

## API Endpoints

The existing MFA API endpoints now support RADIUS:

- `POST /api/mfa/setup/initiate` - Initialize RADIUS MFA setup
- `POST /api/mfa/setup/verify` - Verify RADIUS credentials
- `POST /api/mfa/setup/enable` - Enable RADIUS MFA for user
- `POST /api/mfa/setup/disable` - Disable MFA (including RADIUS)

See `RADIUS_MFA_SETUP.md` for detailed API usage examples.

## Testing

Unit tests are included in `object/mfa_radius_test.go`:

```bash
# Run RADIUS MFA tests
go test ./object -run ".*Radius.*" -v

# Run all MFA tests
go test ./object -run ".*Mfa.*" -v
```

## Configuration Example

### 1. Create Provider

```json
{
  "owner": "built-in",
  "name": "radius-mfa-server",
  "displayName": "RADIUS MFA Server",
  "category": "MFA",
  "type": "RADIUS",
  "clientSecret": "your-radius-shared-secret",
  "host": "10.10.10.10",
  "port": 1812
}
```

### 2. Enable for User

```bash
# Initiate setup
curl -X POST http://localhost:8000/api/mfa/setup/initiate \
  -d "owner=built-in&name=alice&mfaType=radius"

# Verify credentials
curl -X POST http://localhost:8000/api/mfa/setup/verify \
  -d "mfaType=radius&secret=built-in/radius-mfa-server&dest=alice&passcode=123456"

# Enable MFA
curl -X POST http://localhost:8000/api/mfa/setup/enable \
  -d "owner=built-in&name=alice&mfaType=radius&secret=built-in/radius-mfa-server&dest=alice&recoveryCodes=xxx"
```

## Security Considerations

- **Shared Secret**: Use strong, unique shared secrets for RADIUS communication
- **Network Security**: Consider using VPN or IPsec for RADIUS traffic
- **Access Control**: Restrict RADIUS server to accept requests only from Casdoor
- **Timeouts**: RADIUS requests timeout after 10 seconds
- **Credential Storage**: Provider credentials are stored in Casdoor database

## Comparison with Similar Systems

### CAS (Apereo)

Casdoor's RADIUS MFA implementation is inspired by [CAS RADIUS Authentication](https://apereo.github.io/cas/7.0.x/mfa/RADIUS-Authentication.html):

| Feature | CAS | Casdoor |
|---------|-----|---------|
| RADIUS Client | ✓ | ✓ |
| Shared Secret | ✓ | ✓ |
| Server Config | Properties file | Provider entity |
| User Mapping | Automatic | Per-user configurable |
| Multiple Servers | ✓ | ✓ (via multiple providers) |

### Keycloak

Keycloak supports RADIUS via third-party authenticators. Casdoor provides:
- Native integration with MFA framework
- Provider-based configuration (reusable)
- Per-user RADIUS username mapping

## Future Enhancements

Potential future improvements:

1. **Challenge-Response**: Support for RADIUS Access-Challenge messages
2. **Server Pools**: Failover between multiple RADIUS servers
3. **Custom Attributes**: Support for vendor-specific RADIUS attributes
4. **Connection Pooling**: Reuse RADIUS connections for better performance
5. **Metrics**: Track RADIUS authentication success/failure rates

## References

- [RFC 2865 - RADIUS](https://tools.ietf.org/html/rfc2865)
- [Apereo CAS RADIUS MFA](https://apereo.github.io/cas/7.0.x/mfa/RADIUS-Authentication.html)
- [FreeRADIUS Documentation](https://freeradius.org/documentation/)
- [layeh.com/radius - Go RADIUS library](https://github.com/layeh/radius)

## Support

For questions or issues:
- [GitHub Issues](https://github.com/casdoor/casdoor/issues)
- [Casdoor Discussions](https://github.com/casdoor/casdoor/discussions)
- [Documentation](https://casdoor.org/docs/)

## Contributors

This feature addresses issue: https://github.com/casdoor/casdoor/issues/[issue-number]

Implementation by: @copilot
