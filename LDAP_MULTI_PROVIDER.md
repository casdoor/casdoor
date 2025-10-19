# Multiple LDAP Provider Support

## Overview

Casdoor now supports configuring multiple LDAP providers within a single organization, with intelligent handling of duplicate usernames and proper authentication routing.

## Problem Statement

When multiple LDAP directories are configured (e.g., separate LDAP servers for different departments or offices), users with the same username may exist in different directories. This creates several challenges:

1. **Authentication Ambiguity**: Which LDAP provider should be used to authenticate a user?
2. **Username Conflicts**: How to handle users with the same username from different providers?
3. **Synchronization Conflicts**: How to prevent data overwrites during user synchronization?

## Solution

Casdoor implements a provider-aware system that:

1. **Tracks Provider Association**: Each user is associated with their originating LDAP provider via a `LdapId` field
2. **Prevents Username Conflicts**: Automatically appends LDAP server name to usernames when conflicts are detected
3. **Optimizes Authentication**: Prioritizes authentication against the user's associated LDAP provider

## How It Works

### User Synchronization

When synchronizing users from multiple LDAP providers:

1. **First Provider**: Users are created with their original username (e.g., `john`)
2. **Subsequent Providers**: If a username already exists from a different provider, the LDAP server name is appended (e.g., `john_ServerName`)
3. **Provider Tracking**: The `LdapId` field stores which LDAP provider the user belongs to

```go
// Example: User from LDAP Provider 1
User {
    Name: "john",
    Ldap: "uuid-from-ldap-1",
    LdapId: "ldap-provider-1",
}

// Example: User with same username from LDAP Provider 2
User {
    Name: "john_CompanyB",
    Ldap: "uuid-from-ldap-2",
    LdapId: "ldap-provider-2",
}
```

### Authentication

When a user attempts to log in:

1. **Primary Provider Check**: The system first attempts authentication against the user's associated LDAP provider (stored in `LdapId`)
2. **Fallback**: If the primary provider fails or isn't set, the system attempts authentication against all configured LDAP providers
3. **First Match Wins**: Authentication succeeds with the first provider that validates the credentials

This approach:
- **Improves Performance**: Reduces authentication latency by checking the correct provider first
- **Enhances Security**: Ensures users are authenticated against their intended directory
- **Maintains Compatibility**: Existing users without `LdapId` still work

### Password Changes

Password change operations work similarly:
- The system attempts to change the password in the user's associated LDAP provider
- If the provider is not specified, it searches all providers

## Configuration

### Setting Up Multiple LDAP Providers

1. Navigate to the LDAP configuration page in Casdoor
2. Add multiple LDAP providers, each with:
   - **Server Name**: A unique identifier (e.g., "Office-NYC", "Office-London")
   - **Host**: LDAP server hostname
   - **Port**: LDAP server port
   - **Base DN**: Base distinguished name
   - **Filter**: LDAP search filter
   - **Auto Sync**: Enable automatic synchronization if needed

### Conflict Resolution Strategy

The system automatically resolves username conflicts by appending the LDAP server name. For example:

- User `alice` from "LDAP-Main" → Username: `alice`
- User `alice` from "LDAP-Branch" → Username: `alice_LDAP-Branch`

## Best Practices

### 1. Use Descriptive Server Names

Choose clear, descriptive names for your LDAP servers:
- ✅ Good: "Office-NYC", "Office-London", "Partners-LDAP"
- ❌ Bad: "LDAP1", "Server2", "Test"

### 2. Plan for Username Conflicts

If you know users will have the same usernames across providers:
- Consider using separate organizations for each LDAP provider
- Use email addresses as usernames if possible
- Document the naming convention for users

### 3. Monitor Synchronization

- Enable auto-sync carefully to avoid unexpected user creation
- Review sync logs regularly
- Test with a small subset before full deployment

### 4. Test Authentication

After configuring multiple providers:
- Test authentication with users from each provider
- Verify that users are authenticated against the correct directory
- Check that password changes work correctly

## Backward Compatibility

This feature is fully backward compatible:

- **Existing Users**: Users without `LdapId` will continue to work
- **Existing Authentication**: The fallback mechanism ensures all providers are checked
- **Database Migration**: The `LdapId` field is automatically added to the schema

## Technical Details

### Database Schema

New field added to the `user` table:
```sql
ldap_id VARCHAR(100) INDEX
```

### Modified Functions

1. **SyncLdapUsers**: Sets `LdapId` when creating users
2. **buildLdapUserName**: Checks for conflicts across providers
3. **CheckLdapUserPassword**: Prioritizes user's associated provider

## Troubleshooting

### Issue: Users can't authenticate after adding a second LDAP provider

**Solution**: Ensure the new provider's `BaseDn` and `Filter` are correctly configured. Check logs for authentication errors.

### Issue: Duplicate users created during sync

**Solution**: This should not happen with the new implementation. If it does:
1. Check that `LdapId` is being set correctly
2. Verify the LDAP UUID is unique across providers
3. Review the sync logs for errors

### Issue: Username has unexpected suffix

**Solution**: This indicates a username conflict. The suffix is the LDAP server name. To avoid:
1. Ensure usernames are unique across providers
2. Use separate organizations if needed
3. Consider using email-based usernames

## API Changes

### User Object

The User object now includes:
```json
{
  "ldap": "uuid-from-ldap",
  "ldapId": "ldap-provider-id"
}
```

### LDAP Sync Response

No changes to the API response format. The `LdapId` is set internally.

## Security Considerations

1. **Provider Isolation**: Users are authenticated only against their associated provider, reducing attack surface
2. **Credential Verification**: Each LDAP provider's credentials are validated independently
3. **Audit Trail**: The `LdapId` field provides clear tracking of user origin for auditing

## Future Enhancements

Potential improvements for future versions:

1. **Conflict Resolution Policies**: Configurable strategies for handling username conflicts (append, prefix, reject)
2. **Provider Priority**: Allow administrators to set authentication priority order
3. **Cross-Provider Search**: Search for users across all LDAP providers in the admin UI
4. **Provider Failover**: Automatic failover to backup providers if primary is unavailable

## Questions and Support

If you encounter issues or have questions:

1. Check the Casdoor logs for detailed error messages
2. Review this documentation and best practices
3. Open an issue on GitHub with:
   - Casdoor version
   - LDAP provider details (sanitized)
   - Steps to reproduce
   - Expected vs actual behavior

## Example Configuration

### Scenario: Company with Two Offices

**Setup:**
- Main office in New York with primary LDAP
- Branch office in London with separate LDAP
- Some employees have the same first name

**Configuration:**

LDAP Provider 1:
```yaml
Server Name: Office-NYC
Host: ldap-nyc.company.com
Port: 389
Base DN: dc=nyc,dc=company,dc=com
Filter: (objectClass=person)
```

LDAP Provider 2:
```yaml
Server Name: Office-London
Host: ldap-london.company.com
Port: 389
Base DN: dc=london,dc=company,dc=com
Filter: (objectClass=person)
```

**Result:**
- User "john" from NYC → Username: `john`
- User "john" from London → Username: `john_Office-London`

Both users can authenticate with their credentials, and the system routes authentication to the correct directory.
