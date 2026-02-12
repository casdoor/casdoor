# LDAP Group/Department Sync with Nested Hierarchies

## Overview

This feature enables Casdoor to synchronize LDAP groups and organizational units (OUs) to Casdoor groups/departments, maintaining hierarchical structures and automatically assigning users to groups based on their LDAP `memberOf` attributes.

## Features

### 1. **Hierarchical Group Syncing**
- Syncs LDAP groups (groupOfNames, groupOfUniqueNames, posixGroup, AD groups)
- Syncs organizational units (OUs) as groups
- Maintains parent-child relationships in Casdoor
- Supports unlimited nesting levels

### 2. **Automatic User-Group Assignment**
- Extracts group memberships from user's `memberOf` attribute
- Assigns users to multiple groups automatically
- Maintains backward compatibility with `DefaultGroup` configuration

### 3. **Auto-Sync Support**
- Groups sync automatically with the configured interval
- Groups are synced before users to ensure proper references
- Can be enabled/disabled via `EnableGroups` configuration flag

## Configuration

### LDAP Server Configuration

To enable group syncing, add the following field to your LDAP configuration:

```json
{
  "id": "ldap-server-id",
  "owner": "organization-name",
  "serverName": "My LDAP Server",
  "host": "ldap.example.com",
  "port": 389,
  "username": "cn=admin,dc=example,dc=com",
  "password": "password",
  "baseDn": "dc=example,dc=com",
  "filter": "(objectClass=person)",
  "autoSync": 60,
  "enableGroups": true    // Enable group syncing
}
```

### Configuration Fields

| Field | Type | Description |
|-------|------|-------------|
| `enableGroups` | boolean | Enable/disable LDAP group syncing. Default: `false` |
| `autoSync` | integer | Auto-sync interval in minutes. Set to 0 to disable |
| `defaultGroup` | string | Default group assigned to all synced users (still supported) |

## How It Works

### Group Discovery

The system discovers groups through two methods:

1. **LDAP Groups**: Searches for objects with these classes:
   - `groupOfNames`
   - `groupOfUniqueNames`
   - `posixGroup`
   - `group` (Active Directory)

2. **Organizational Units**: Searches for objects with:
   - `organizationalUnit`

### Group Hierarchy

Groups are organized hierarchically based on their Distinguished Name (DN):

**Example LDAP Structure:**
```
dc=example,dc=com
├── ou=Departments
│   ├── ou=Engineering
│   │   ├── ou=Backend
│   │   └── ou=Frontend
│   └── ou=Sales
│       ├── ou=NA
│       └── ou=EMEA
└── ou=Groups
    ├── cn=Admins
    └── cn=Developers
```

**Resulting Casdoor Groups:**
```
Departments (top-level)
├── Departments_Engineering (parent: Departments)
│   ├── Departments_Engineering_Backend (parent: Departments_Engineering)
│   └── Departments_Engineering_Frontend (parent: Departments_Engineering)
└── Departments_Sales (parent: Departments)
    ├── Departments_Sales_NA (parent: Departments_Sales)
    └── Departments_Sales_EMEA (parent: Departments_Sales)

Groups (top-level)
├── Groups_Admins (parent: Groups)
└── Groups_Developers (parent: Groups)
```

### Group Naming

Groups are named based on their DN path, excluding domain components (DC):

- **DN**: `OU=Backend,OU=Engineering,OU=Departments,DC=example,DC=com`
- **Casdoor Group Name**: `Departments_Engineering_Backend`
- **Display Name**: `Backend` (the leaf name)

### User-Group Assignment

Users are assigned to groups based on their `memberOf` attribute:

**Example User:**
```ldap
dn: cn=john.doe,ou=Users,dc=example,dc=com
memberOf: cn=Developers,ou=Groups,dc=example,dc=com
memberOf: ou=Backend,ou=Engineering,ou=Departments,dc=example,dc=com
```

**Resulting User Groups in Casdoor:**
```json
{
  "groups": [
    "Groups_Developers",
    "Departments_Engineering_Backend"
  ]
}
```

## API Reference

### Core Functions

#### `GetLdapGroups(ldapServer *Ldap) ([]LdapGroup, error)`
Fetches all LDAP groups and OUs from the LDAP server.

**Returns:**
- Array of `LdapGroup` objects containing DN, name, description, members, and parent DN

#### `SyncLdapGroups(owner string, ldapGroups []LdapGroup, ldapId string) (newGroups int, updatedGroups int, err error)`
Syncs LDAP groups to Casdoor groups, creating or updating as needed.

**Parameters:**
- `owner`: Organization name
- `ldapGroups`: Array of LDAP groups to sync
- `ldapId`: LDAP server ID

**Returns:**
- `newGroups`: Number of newly created groups
- `updatedGroups`: Number of updated groups
- `err`: Error if sync fails

#### `SyncLdapUsers(owner string, syncUsers []LdapUser, ldapId string) (existUsers, failedUsers []LdapUser, err error)`
Enhanced to assign users to groups based on `memberOf` attribute.

### Data Structures

#### `LdapGroup`
```go
type LdapGroup struct {
    Dn          string   // Distinguished Name
    Cn          string   // Common Name
    Name        string   // Group name
    Description string   // Description
    Member      []string // Group members
    ParentDn    string   // Parent DN
}
```

#### `LdapUser` (Enhanced)
```go
type LdapUser struct {
    // ... existing fields ...
    MemberOf []string // Changed from string to []string
}
```

## Usage Examples

### Manual Sync via API

```bash
# Sync LDAP users (will also sync groups if enableGroups is true)
curl -X POST "https://your-casdoor.com/api/sync-ldap-users?id=organization/ldap-id" \
  -H "Content-Type: application/json" \
  -d '[...]'
```

### Programmatic Configuration

```go
ldap := &object.Ldap{
    Owner:        "my-org",
    ServerName:   "Corporate LDAP",
    Host:         "ldap.company.com",
    Port:         389,
    Username:     "cn=admin,dc=company,dc=com",
    Password:     "secret",
    BaseDn:       "dc=company,dc=com",
    Filter:       "(objectClass=person)",
    AutoSync:     60,
    EnableGroups: true, // Enable group syncing
}

affected, err := object.AddLdap(ldap)
```

## Supported LDAP Servers

- **OpenLDAP**
- **Microsoft Active Directory**
- **FreeIPA**
- **389 Directory Server**
- Any LDAP v3 compliant server

## Migration Guide

### Upgrading from Previous Versions

The new feature is **backward compatible**. Existing LDAP configurations will continue to work without changes:

1. Groups syncing is **disabled by default** (`enableGroups: false`)
2. `DefaultGroup` configuration still works
3. Existing users are not affected

### Enabling Group Sync

To enable group syncing for existing LDAP servers:

1. Update the LDAP configuration to set `enableGroups: true`
2. Groups will be synced on the next auto-sync cycle
3. Newly synced users will be assigned to groups based on `memberOf`
4. Existing users can be re-synced to update their group memberships

## Troubleshooting

### Groups Not Syncing

**Check:**
- `enableGroups` is set to `true`
- LDAP user has permissions to read group objects
- Groups exist within the configured `baseDn`

**View Logs:**
```bash
# Look for group sync messages
grep "ldap group sync" /path/to/casdoor.log
```

### User Not Assigned to Groups

**Check:**
- User has `memberOf` attribute populated in LDAP
- Groups were synced before the user
- Group DN matches synced group names

### Duplicate Groups

Group names are generated from DN paths. If you have:
- `ou=Sales,ou=Dept1,dc=example,dc=com`
- `ou=Sales,ou=Dept2,dc=example,dc=com`

They will create distinct groups:
- `Dept1_Sales`
- `Dept2_Sales`

## Performance Considerations

- **Large Directories**: Group sync uses paging (100 entries per page) to handle large directories efficiently
- **Hierarchy Processing**: Groups are processed in dependency order (parents before children)
- **Sync Frequency**: Consider your directory size when setting `autoSync` interval
- **First Sync**: Initial sync may take longer due to group creation

## Security Considerations

- LDAP credentials are stored encrypted
- Group membership is read-only from LDAP
- Only LDAP users can sync groups (requires valid LDAP connection)
- Group hierarchy changes are applied atomically

## Limitations

- Groups are read-only (synced from LDAP, not bi-directional)
- Deleting LDAP groups doesn't auto-delete Casdoor groups
- Group names cannot contain "/" characters
- Maximum DN depth is limited by database field size (500 characters for BaseDn)

## Future Enhancements

Potential improvements for future versions:

- Bi-directional group sync
- Group deletion sync
- Custom group naming patterns
- Group attribute mapping
- Selective group sync (filter by DN pattern)

## Support

For issues, questions, or feature requests:
- GitHub Issues: https://github.com/casdoor/casdoor/issues
- Discord: https://discord.gg/5rPsrAzK7S
- Documentation: https://casdoor.org/docs

## License

This feature is part of Casdoor and is licensed under the Apache License 2.0.
