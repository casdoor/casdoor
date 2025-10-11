# LDAP Custom Attributes Mapping

## Overview

Casdoor supports syncing custom LDAP attributes into user properties. This feature allows you to map any LDAP attribute (e.g., `department`, `employeeNumber`, `jobTitle`, etc.) from your LDAP directory to Casdoor user properties.

## How It Works

When configuring an LDAP provider in Casdoor, you can define custom attribute mappings that specify:
- **LDAP Attribute Name**: The attribute name in your LDAP directory (e.g., `department`, `employeeNumber`)
- **User Property Name**: The property name in the Casdoor user object (e.g., `dept`, `empId`)

During LDAP synchronization, Casdoor will:
1. Fetch the specified custom attributes from your LDAP server
2. Map them to the corresponding property names
3. Store them in the user's `Properties` field as key-value pairs

## Configuration Steps

### 1. Add or Edit LDAP Provider

Navigate to the LDAP configuration page in Casdoor admin panel:
- Go to **Providers** → **LDAP**
- Click **Add** or select an existing LDAP provider to edit

### 2. Configure Custom Attributes

In the LDAP edit page, find the **Custom attributes** section:

1. Click the **Add** button to add a new attribute mapping
2. Enter the **LDAP attribute name** (e.g., `department`)
3. Enter the **User property name** (e.g., `department` or any custom name like `dept`)
4. Repeat for each custom attribute you want to sync

Example mappings:
| LDAP Attribute Name | User Property Name |
|---------------------|-------------------|
| department          | department        |
| employeeNumber      | employeeId        |
| title               | jobTitle          |
| physicalDeliveryOfficeName | office    |
| manager             | manager           |

### 3. Save and Sync

1. Click **Save** to store the LDAP configuration
2. Go to **LDAP Sync** page
3. Click **Sync** to synchronize users from LDAP

## Accessing Custom Properties

Once users are synchronized, their custom properties can be accessed through the user object's `Properties` field.

### Example API Response

```json
{
  "owner": "organization1",
  "name": "john.doe",
  "displayName": "John Doe",
  "email": "john.doe@example.com",
  "properties": {
    "department": "Engineering",
    "employeeId": "12345",
    "jobTitle": "Senior Software Engineer",
    "office": "Building A, Floor 3",
    "manager": "jane.smith"
  }
}
```

### Using Properties in Applications

Applications integrating with Casdoor can access these properties through:

1. **User Info Endpoint**: The properties are included in the user info response
2. **JWT Token**: Custom properties can be included in JWT claims if configured
3. **SCIM API**: Properties are accessible through the SCIM user schema

## Supported LDAP Servers

This feature works with:
- **Microsoft Active Directory**: All standard and custom attributes
- **OpenLDAP**: All standard and custom attributes
- **Other LDAP v3 compliant servers**: Any attributes exposed by the server

## Common Use Cases

### 1. Department and Team Management
Map `department`, `team`, or `division` attributes to group users by organizational structure.

### 2. Employee Information
Map `employeeNumber`, `employeeType`, `title` to maintain employee records.

### 3. Location and Office Management
Map `physicalDeliveryOfficeName`, `l` (locality), `st` (state) for office location tracking.

### 4. Manager Hierarchy
Map `manager` attribute to maintain organizational hierarchy.

### 5. Custom Business Fields
Map any custom schema extensions in your LDAP directory.

## Technical Details

### Backend Implementation

- **Ldap Struct**: Contains `CustomAttributes map[string]string` field to store mappings
- **LdapUser Struct**: Contains `Attributes map[string]string` field to store fetched values
- **User Struct**: Contains `Properties map[string]string` field to store final values
- During sync, attributes flow: LDAP → LdapUser.Attributes → User.Properties

### Database Schema

Custom attributes are stored in the `ldap` table:
```sql
custom_attributes TEXT  -- JSON-encoded map of attribute mappings
```

User properties are stored in the `user` table:
```sql
properties TEXT  -- JSON-encoded map of property key-value pairs
```

### API Endpoints

- `POST /api/add-ldap`: Add LDAP configuration with custom attributes
- `POST /api/update-ldap`: Update LDAP configuration including custom attributes
- `GET /api/get-ldap`: Retrieve LDAP configuration
- `POST /api/sync-ldap-users`: Sync users with custom attributes

## Limitations

1. **Value Types**: All custom attributes are stored as strings
2. **Multi-valued Attributes**: Only the first value is stored for multi-valued LDAP attributes
3. **Performance**: Large numbers of custom attributes may impact sync performance
4. **Naming**: Property names must be valid JSON keys (no special characters)

## Troubleshooting

### Custom attributes not syncing

1. **Check LDAP attribute names**: Ensure the attribute names match exactly (case-sensitive)
2. **Verify attribute exists**: Use LDAP browser tools to confirm attributes exist in your directory
3. **Check permissions**: Ensure the LDAP bind user has read access to the attributes
4. **Review logs**: Check Casdoor logs for any LDAP query errors

### Properties not appearing in user object

1. **Verify sync completed**: Check that users were successfully synced
2. **Check property names**: Ensure property names don't conflict with existing user fields
3. **Review database**: Query the user table to verify properties column contains data

## Example: Complete Configuration

Here's a complete example for Microsoft Active Directory:

**LDAP Configuration:**
```json
{
  "host": "ldap.company.com",
  "port": 389,
  "baseDn": "DC=company,DC=com",
  "filter": "(objectClass=person)",
  "username": "CN=admin,DC=company,DC=com",
  "password": "***",
  "customAttributes": {
    "department": "department",
    "employeeID": "employeeId",
    "title": "jobTitle",
    "physicalDeliveryOfficeName": "office",
    "manager": "managerDN",
    "extensionAttribute1": "costCenter"
  }
}
```

**Resulting User:**
```json
{
  "name": "john.doe",
  "displayName": "John Doe",
  "email": "john.doe@company.com",
  "properties": {
    "department": "Engineering",
    "employeeId": "EMP12345",
    "jobTitle": "Software Engineer",
    "office": "NYC-Building1",
    "managerDN": "CN=Jane Smith,OU=Users,DC=company,DC=com",
    "costCenter": "CC-ENG-001"
  }
}
```

## Best Practices

1. **Use meaningful property names**: Choose property names that clearly indicate their purpose
2. **Document your mappings**: Keep a record of which LDAP attributes map to which properties
3. **Test with small groups**: Test attribute mappings with a small user group first
4. **Regular audits**: Periodically verify that mappings are still accurate
5. **Avoid sensitive data**: Don't map sensitive attributes unless absolutely necessary

## Related Features

- **LDAP Auto Sync**: Automatically sync users at regular intervals
- **Default Groups**: Assign default groups to LDAP-synced users
- **LDAP Authentication**: Authenticate users against LDAP directory
- **Filter Fields**: Control which LDAP fields are used for authentication

## Version History

- **v1.x**: Initial LDAP support with standard attributes only
- **Current**: Full support for custom LDAP attribute mapping
