# Custom Token Attributes (JWT-Custom) - Examples

This document provides examples of how to use custom token attributes in Casdoor applications with the JWT-Custom token format.

## Overview

When you select **JWT-Custom** as the Token Format for your application, you can define custom JWT claims using the **Token Attributes** table. This allows you to include dynamic user data in your access tokens.

## Supported Dynamic Values

You can use the following placeholders in the "Value" field to include dynamic user data:

| Placeholder | Description | Returns |
|------------|-------------|---------|
| `$user.roles` | User's role names | Array of role names |
| `$user.permissions` | User's permission names | Array of permission names |
| `$user.groups` | User's groups | Array of group names |
| `$user.owner` | User's organization | String |
| `$user.name` | User's username | String |
| `$user.email` | User's email address | String |
| `$user.id` | User's unique ID | String |
| `$user.phone` | User's phone number | String |

## Example 1: Custom Roles Field (warpgate_roles)

**Use Case:** You want to include user roles in a custom claim named "warpgate_roles" as an array.

**Configuration:**
- **Name:** `warpgate_roles`
- **Value:** `$user.roles`
- **Type:** `Array`

**Result in JWT:**
```json
{
  "warpgate_roles": ["admin", "developer", "viewer"],
  "sub": "admin",
  "aud": ["your-app-client-id"],
  ...
}
```

## Example 2: Custom Permissions Field

**Use Case:** You want to include user permissions in a custom claim named "app_permissions".

**Configuration:**
- **Name:** `app_permissions`
- **Value:** `$user.permissions`
- **Type:** `Array`

**Result in JWT:**
```json
{
  "app_permissions": ["read", "write", "delete"],
  ...
}
```

## Example 3: User Groups

**Use Case:** You want to include user groups in a custom claim named "user_groups".

**Configuration:**
- **Name:** `user_groups`
- **Value:** `$user.groups`
- **Type:** `Array`

**Result in JWT:**
```json
{
  "user_groups": ["engineering", "product"],
  ...
}
```

## Example 4: Single Value Fields

**Use Case:** You want to include the user's organization as a string.

**Configuration:**
- **Name:** `org`
- **Value:** `$user.owner`
- **Type:** `String`

**Result in JWT:**
```json
{
  "org": "my-company",
  ...
}
```

## Example 5: Multiple Attributes

You can define multiple custom attributes for a single application:

| Name | Value | Type |
|------|-------|------|
| `warpgate_roles` | `$user.roles` | Array |
| `user_email` | `$user.email` | String |
| `user_org` | `$user.owner` | String |
| `user_groups` | `$user.groups` | Array |

**Result in JWT:**
```json
{
  "warpgate_roles": ["admin", "developer"],
  "user_email": "user@example.com",
  "user_org": "my-company",
  "user_groups": ["engineering"],
  ...
}
```

## Example 6: Template Strings

You can also use template strings by combining placeholders with static text:

**Configuration:**
- **Name:** `full_user_id`
- **Value:** `$user.owner/$user.name`
- **Type:** `String`

**Result in JWT:**
```json
{
  "full_user_id": "my-company/johndoe",
  ...
}
```

## How to Configure

1. Navigate to your application's edit page
2. Set **Token Format** to `JWT-Custom`
3. Scroll to the **Token Attributes** section
4. Click **Add** to create a new attribute
5. Fill in:
   - **Name**: The name of the JWT claim (e.g., `warpgate_roles`)
   - **Value**: The dynamic value using placeholders (e.g., `$user.roles`)
   - **Type**: Choose `Array` for lists or `String` for single values
6. Save your application

## Testing Your Configuration

After configuring your token attributes:

1. Obtain an access token by logging in through your application
2. Decode the JWT token (you can use https://jwt.io)
3. Verify that your custom claims appear in the token payload

## Notes

- Array types will include all values from the placeholder (e.g., all user roles)
- String types will use only the first value from array placeholders
- Static values (without placeholders) are also supported
- All custom attributes are in addition to standard JWT claims (iss, sub, aud, exp, iat, etc.)

## Related Documentation

- [Token Format Options](https://casdoor.org/docs/token/overview/#token-format-options)
- [Application Configuration](https://casdoor.org/docs/application/overview)
