# Custom OIDC Claims in Casdoor

This guide explains how to configure custom OIDC (OpenID Connect) claims in Casdoor, including dynamic claims that change based on user data such as roles, permissions, and groups.

## Table of Contents

- [Overview](#overview)
- [Token Formats](#token-formats)
- [Configuration Methods](#configuration-methods)
  - [Method 1: Using TokenFields](#method-1-using-tokenfields)
  - [Method 2: Using TokenAttributes](#method-2-using-tokenattributes)
- [Dynamic Claim Variables](#dynamic-claim-variables)
- [Common Use Cases](#common-use-cases)
- [Examples](#examples)

## Overview

Casdoor supports customizing the claims included in JWT tokens through the Application configuration. There are two primary ways to add custom claims:

1. **TokenFields**: Select which user properties to include in the token
2. **TokenAttributes**: Define custom claims with template values that can dynamically reference user data

## Token Formats

Casdoor supports four token formats, configured via the `TokenFormat` field in the Application:

- **`JWT`**: Full user object with all fields including roles, permissions, and groups
- **`JWT-Empty`**: Minimal token with only essential claims
- **`JWT-Custom`**: Customizable token using `TokenFields` and `TokenAttributes`
- **`JWT-Standard`**: Standard OIDC token format following OpenID Connect specification

For custom claims, use **`JWT-Custom`** format, which allows you to specify exactly which claims to include.

## Configuration Methods

### Method 1: Using TokenFields

`TokenFields` is an array of user property names that should be included as claims in the token. This is useful when you want to include standard user fields.

**Available User Fields:**

Standard fields:
- `Id`, `Owner`, `Name`
- `DisplayName`, `FirstName`, `LastName`
- `Email`, `Phone`, `CountryCode`
- `Region`, `Location`, `Address`
- `Avatar`, `Gender`, `Birthday`
- `Tag`, `Language`, `Bio`
- `Affiliation`, `Title`

Special fields:
- `Roles` - Array of user's role objects
- `Permissions` - Array of user's permission objects  
- `Groups` - Array of user's group names
- `permissionNames` - Array of permission names only (special handling)
- `Properties.<fieldName>` - Access custom user properties

**Example Configuration:**

```json
{
  "tokenFormat": "JWT-Custom",
  "tokenFields": [
    "Email",
    "DisplayName",
    "Roles",
    "Permissions"
  ]
}
```

This will include the user's email, display name, roles, and permissions in the JWT token.

### Method 2: Using TokenAttributes

`TokenAttributes` allows you to define custom claim names with dynamic values using template variables. This is more flexible and allows you to create claims with custom names.

Each token attribute has three properties:
- `name`: The claim name in the JWT token
- `value`: The template value (can use variables)
- `type`: Either `"String"` (single value) or `"Array"` (multiple values)

**Example Configuration:**

```json
{
  "tokenFormat": "JWT-Custom",
  "tokenAttributes": [
    {
      "name": "warpgate_roles",
      "value": "$user.roles",
      "type": "Array"
    },
    {
      "name": "user_email",
      "value": "$user.email",
      "type": "String"
    }
  ]
}
```

## Dynamic Claim Variables

The following template variables can be used in `TokenAttributes` to create dynamic claims:

### Array Variables (return multiple values)

- **`$user.roles`** - List of role names assigned to the user
- **`$user.permissions`** - List of permission names assigned to the user  
- **`$user.groups`** - List of group names the user belongs to

### String Variables (return single value)

- **`$user.owner`** - Organization owner
- **`$user.name`** - Username
- **`$user.email`** - Email address
- **`$user.id`** - User ID
- **`$user.phone`** - Phone number

### Custom Properties

You can also access custom user properties:
- Use `Properties.<propertyName>` in `TokenFields`
- Properties will be included with the property name as the claim name

## Common Use Cases

### Use Case 1: Including User Roles in Token

**Problem**: You want to include the user's roles as a JSON array in the token.

**Solution using TokenFields**:
```json
{
  "tokenFormat": "JWT-Custom",
  "tokenFields": ["Email", "DisplayName", "Roles"]
}
```

**Solution using TokenAttributes**:
```json
{
  "tokenFormat": "JWT-Custom",
  "tokenAttributes": [
    {
      "name": "roles",
      "value": "$user.roles",
      "type": "Array"
    }
  ]
}
```

### Use Case 2: Custom Claim Name for Roles (e.g., warpgate_roles)

**Problem**: You need roles in a specific claim name like `warpgate_roles`.

**Solution**:
```json
{
  "tokenFormat": "JWT-Custom",
  "tokenAttributes": [
    {
      "name": "warpgate_roles",
      "value": "$user.roles",
      "type": "Array"
    }
  ]
}
```

The resulting token will include:
```json
{
  "warpgate_roles": ["admin", "developer", "viewer"],
  ...
}
```

### Use Case 3: Including Roles and Permissions

**Problem**: You need both roles and permissions in the token.

**Solution**:
```json
{
  "tokenFormat": "JWT-Custom",
  "tokenAttributes": [
    {
      "name": "roles",
      "value": "$user.roles",
      "type": "Array"
    },
    {
      "name": "permissions",
      "value": "$user.permissions",
      "type": "Array"
    }
  ]
}
```

### Use Case 4: Combining Static and Dynamic Claims

**Problem**: You want a mix of user fields and custom claims.

**Solution**:
```json
{
  "tokenFormat": "JWT-Custom",
  "tokenFields": ["Email", "DisplayName", "Id"],
  "tokenAttributes": [
    {
      "name": "app_roles",
      "value": "$user.roles",
      "type": "Array"
    },
    {
      "name": "user_groups",
      "value": "$user.groups",
      "type": "Array"
    }
  ]
}
```

### Use Case 5: Using Custom Properties

**Problem**: You have custom user properties and want to include them.

**Solution**:
```json
{
  "tokenFormat": "JWT-Custom",
  "tokenFields": ["Email", "Properties.department", "Properties.employee_id"]
}
```

This will include `department` and `employee_id` as separate claims in the token.

## Examples

### Example 1: Minimal Token with Roles

Configure your Application with:
```json
{
  "name": "my-app",
  "tokenFormat": "JWT-Custom",
  "tokenFields": ["Email"],
  "tokenAttributes": [
    {
      "name": "roles",
      "value": "$user.roles",
      "type": "Array"
    }
  ]
}
```

Resulting JWT payload:
```json
{
  "iss": "https://your-casdoor-instance",
  "sub": "user-id",
  "aud": ["my-app"],
  "exp": 1234567890,
  "iat": 1234567890,
  "email": "user@example.com",
  "roles": ["admin", "developer"]
}
```

### Example 2: Comprehensive Token

Configure your Application with:
```json
{
  "name": "my-app",
  "tokenFormat": "JWT-Custom",
  "tokenFields": ["Email", "DisplayName", "Phone"],
  "tokenAttributes": [
    {
      "name": "user_roles",
      "value": "$user.roles",
      "type": "Array"
    },
    {
      "name": "user_permissions",
      "value": "$user.permissions",
      "type": "Array"
    },
    {
      "name": "user_groups",
      "value": "$user.groups",
      "type": "Array"
    },
    {
      "name": "user_id",
      "value": "$user.id",
      "type": "String"
    }
  ]
}
```

Resulting JWT payload:
```json
{
  "iss": "https://your-casdoor-instance",
  "sub": "user-id",
  "aud": ["my-app"],
  "exp": 1234567890,
  "iat": 1234567890,
  "email": "user@example.com",
  "display_name": "John Doe",
  "phone": "+1234567890",
  "user_roles": ["admin", "developer"],
  "user_permissions": ["read:users", "write:users", "delete:users"],
  "user_groups": ["engineering", "leadership"],
  "user_id": "abc123"
}
```

### Example 3: Warpgate Integration

For the specific use case mentioned in the issue, configure your Application with:

```json
{
  "name": "warpgate-app",
  "tokenFormat": "JWT-Custom",
  "tokenFields": ["Email", "DisplayName"],
  "tokenAttributes": [
    {
      "name": "warpgate_roles",
      "value": "$user.roles",
      "type": "Array"
    }
  ]
}
```

This will create tokens with the `warpgate_roles` claim containing the user's current roles as a JSON array that updates automatically when the user's roles change in Casdoor.

## Configuration via UI

To configure custom claims via the Casdoor web UI:

1. Navigate to **Applications** in the Casdoor admin panel
2. Select or create your application
3. Set **Token Format** to `JWT-Custom`
4. In the **Token Fields** section, add the user fields you want to include
5. In the **Token Attributes** section, click "Add" to create custom claims:
   - Enter the claim **Name** (e.g., `warpgate_roles`)
   - Enter the **Value** template (e.g., `$user.roles`)
   - Select the **Type** (`String` or `Array`)
6. Save the application configuration

## Configuration via API

You can also configure custom claims programmatically using the Casdoor API:

```bash
# Update application with custom claims
curl -X POST https://your-casdoor-instance/api/update-application \
  -H "Content-Type: application/json" \
  -d '{
    "owner": "your-org",
    "name": "my-app",
    "tokenFormat": "JWT-Custom",
    "tokenFields": ["Email", "DisplayName"],
    "tokenAttributes": [
      {
        "name": "warpgate_roles",
        "value": "$user.roles",
        "type": "Array"
      }
    ]
  }'
```

## Important Notes

1. **Standard Claims**: The JWT always includes standard registered claims (`iss`, `sub`, `aud`, `exp`, `nbf`, `iat`, `jti`) regardless of configuration.

2. **Built-in Claims**: When using `JWT-Custom` format, the token also includes `tokenType`, `nonce`, `scope`, and `azp` (if applicable).

3. **Dynamic Updates**: Claims using template variables like `$user.roles` are dynamically evaluated each time a token is generated, ensuring they always reflect the current user state.

4. **Field Name Conversion**: User field names in `tokenFields` are automatically converted to snake_case in the token (e.g., `DisplayName` becomes `display_name`).

5. **Permission Names**: Use the special field `permissionNames` in `tokenFields` to get an array of just the permission names instead of full permission objects.

## Troubleshooting

**Problem**: Custom claims not appearing in token  
**Solution**: Ensure `tokenFormat` is set to `JWT-Custom`, not `JWT`, `JWT-Empty`, or `JWT-Standard`.

**Problem**: Roles showing as objects instead of names  
**Solution**: Use `$user.roles` in `tokenAttributes` instead of `Roles` in `tokenFields`. The template variable extracts just the role names.

**Problem**: Claim is empty even though user has roles  
**Solution**: Verify the user actually has roles assigned in Casdoor and that the roles are enabled.

**Problem**: Array claim showing as a single string  
**Solution**: Ensure the `type` is set to `"Array"`, not `"String"` in the token attribute configuration.

## Additional Resources

- [Casdoor Documentation](https://casdoor.org)
- [OpenID Connect Specification](https://openid.net/specs/openid-connect-core-1_0.html)
- [JWT Claims](https://datatracker.ietf.org/doc/html/rfc7519#section-4)

## Source Code References

For developers wanting to understand the implementation:

- Token generation: `object/token_jwt.go`
- Custom claims processing: `getClaimsCustom()` function
- Template variable replacement: `object/user_util.go` - `replaceAttributeValue()` function
- Role/permission extraction: `getUserRoleNames()`, `getUserPermissionNames()` functions
