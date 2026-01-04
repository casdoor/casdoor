# Session-Level Single Logout (SLO) Implementation

## Overview

This document describes the session-level Single Logout (SLO) feature that enables synchronized logout across all subsystems in an SSO ecosystem.

## Problem Statement

Previously, when a user logged out from one subsystem using the `sso-logout` API, other subsystems would receive notifications containing `sessionId` but couldn't map this to their active access tokens. This prevented synchronized logout across all subsystems sharing the same SSO session.

## Solution

The implementation adds session tracking to access tokens and provides mechanisms for subsystems to identify and expire tokens associated with specific sessions.

### Key Components

#### 1. Token-Session Mapping

- **New Field**: Added `SessionId` field to the `Token` model
- **Automatic Tracking**: When tokens are created during OAuth flows, they are automatically associated with the current session ID
- **Database Index**: The `SessionId` field is indexed for efficient lookups

#### 2. API Endpoint for Token Query

A new API endpoint allows subsystems to query tokens by session IDs:

```
GET /api/get-tokens-by-session-ids?sessionIds=session1,session2,session3
```

**Response**: Returns a list of tokens (with sensitive data masked) associated with the provided session IDs.

Example response:
```json
{
  "status": "ok",
  "data": [
    {
      "owner": "built-in",
      "name": "token-abc123",
      "application": "my-app",
      "organization": "my-org",
      "user": "alice",
      "accessTokenHash": "abc123hash...",
      "refreshTokenHash": "xyz789hash...",
      "sessionId": "session1",
      "expiresIn": 3600,
      "scope": "read write",
      "createdTime": "2024-01-04T00:00:00Z"
    }
  ]
}
```

#### 3. Enhanced SSO Logout Notifications

The `SsoLogoutNotification` structure now includes:

- **SessionIds**: List of session IDs being logged out
- **AccessTokenHashes**: List of access token hashes being expired
- **SessionTokenMap**: Direct mapping from `sessionId` to `accessTokenHashes`

Example notification:
```json
{
  "owner": "my-org",
  "name": "alice",
  "displayName": "Alice Smith",
  "email": "alice@example.com",
  "event": "sso-logout",
  "sessionIds": ["session1", "session2"],
  "accessTokenHashes": ["hash1", "hash2", "hash3"],
  "sessionTokenMap": {
    "session1": ["hash1", "hash2"],
    "session2": ["hash3"]
  },
  "nonce": "unique-nonce-123",
  "timestamp": 1704326400,
  "signature": "hmac-sha256-signature"
}
```

## Usage Guide

### For Subsystem Developers

#### Receiving Logout Notifications

1. **Configure Notification Provider**: Set up a notification provider in your Casdoor application
2. **Handle Logout Events**: Listen for SSO logout notifications
3. **Process Session Mapping**: Use the `sessionTokenMap` to identify tokens to expire

Example pseudocode:
```javascript
function handleSsoLogoutNotification(notification) {
  // Verify the signature first
  if (!verifySsoLogoutSignature(notification, clientSecret)) {
    console.error('Invalid notification signature');
    return;
  }

  // Get the current user's session ID
  const currentSessionId = getCurrentSessionId();

  // Check if current session is being logged out
  if (notification.sessionIds.includes(currentSessionId)) {
    // Get the token hashes for this session
    const tokenHashes = notification.sessionTokenMap[currentSessionId];
    
    // Expire local tokens matching these hashes
    for (const hash of tokenHashes) {
      expireTokenByHash(hash);
    }
    
    // Clear local session
    clearUserSession();
    
    // Redirect to logout page
    redirectToLogoutPage();
  }
}
```

#### Querying Tokens by Session ID

Alternatively, you can query Casdoor to get tokens for specific sessions:

```javascript
async function getTokensForSessions(sessionIds) {
  const response = await fetch(
    `/api/get-tokens-by-session-ids?sessionIds=${sessionIds.join(',')}`,
    {
      headers: {
        'Authorization': `Bearer ${adminToken}`
      }
    }
  );
  
  const result = await response.json();
  return result.data;
}
```

### For Administrators

#### Enabling Session-Level SLO

1. **Update Casdoor**: Ensure you're running a version that includes the session-level SLO feature
2. **Configure Notification Providers**: Set up notification providers in your applications for receiving logout events
3. **Update Subsystems**: Ensure all subsystems are updated to handle the enhanced notification format

#### Logout Behavior

The `/api/sso-logout` endpoint supports two modes:

- **Logout All Sessions** (default): `GET /api/sso-logout` or `GET /api/sso-logout?logoutAll=true`
  - Expires all tokens for the user
  - Deletes all user sessions
  - Sends notifications with all session IDs and token hashes

- **Logout Current Session Only**: `GET /api/sso-logout?logoutAll=false`
  - Deletes only the current session
  - Sends notification with only the current session ID
  - Other sessions remain active

## Security Considerations

### Signature Verification

All SSO logout notifications include an HMAC-SHA256 signature to prevent malicious logout requests. The signature is computed over:

- Owner
- Username
- Nonce (for replay protection)
- Timestamp
- Session IDs
- Access token hashes

**Always verify the signature** before processing logout notifications using the `VerifySsoLogoutSignature` function with your application's client secret.

### Token Hash Protection

Access tokens are never sent in plaintext in notifications. Only SHA-256 hashes are included, allowing subsystems to verify they have the corresponding token without exposing the actual token value.

## Migration Notes

### Database Schema Changes

A new `session_id` column has been added to the `token` table with an index for performance. Existing tokens will have a `null` or empty `SessionId` until they are refreshed or new tokens are created.

### Backward Compatibility

The implementation is backward compatible:

- Existing subsystems can continue to use `sessionIds` and `accessTokenHashes` arrays
- The new `sessionTokenMap` provides a convenience mapping but is optional
- The signature verification remains compatible with existing implementations

## API Reference

### GET /api/get-tokens-by-session-ids

Query tokens by session IDs.

**Parameters:**
- `sessionIds` (string, required): Comma-separated list of session IDs

**Response:**
- `status`: "ok" or "error"
- `data`: Array of token objects (with sensitive data masked)

**Authorization:** Requires valid authentication

**Example:**
```bash
curl -X GET "https://your-casdoor.com/api/get-tokens-by-session-ids?sessionIds=session1,session2" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Troubleshooting

### Tokens Not Associated with Sessions

**Symptom**: New tokens have empty `sessionId`

**Solution**: Ensure the OAuth flow is initiated from an authenticated session. Server-to-server flows (client credentials grant) don't have user sessions.

### Logout Notifications Not Received

**Symptom**: Subsystems don't receive logout notifications

**Solution**: 
1. Verify notification providers are configured in the user's signup application
2. Check notification provider credentials and endpoints
3. Review notification provider logs for errors

### Session-Token Mapping is Empty

**Symptom**: `sessionTokenMap` is empty in notifications

**Solution**: This can occur if:
1. Tokens were created before the session tracking was implemented (old tokens)
2. Tokens were created via server-to-server flows without user sessions
3. Database migration hasn't been applied

## Support

For issues or questions:
- GitHub Issues: https://github.com/casdoor/casdoor/issues
- Discord: https://discord.gg/5rPsrAzK7S
- Documentation: https://casdoor.org
