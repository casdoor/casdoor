# External IDP Claim Mapping

## Overview

Casdoor now supports mapping claims from external Identity Providers (IdPs) like Azure AD, Okta, and others to additional user properties beyond the basic fields (name, email, phone, avatar). This feature allows you to automatically populate user profile fields based on the claims returned by the external IdP during authentication.

## Supported Fields

The following user fields can now be mapped from IdP claims:

### Standard Fields
- `id` - User ID
- `username` - Username
- `displayName` - Display name
- `email` - Email address
- `avatarUrl` - Avatar URL

### Extended Fields (New)
- `phone` - Phone number
- `countryCode` - Country code
- `firstName` - First name
- `lastName` - Last name
- `region` - Region/State
- `location` - Location/City
- `affiliation` - Company/Organization
- `title` - Job title
- `homepage` - Personal website
- `bio` - Biography
- `tag` - User tag
- `language` - Preferred language
- `gender` - Gender
- `birthday` - Date of birth
- `education` - Education level
- `idCard` - ID card number
- `idCardType` - ID card type

## Configuration

### Backend Configuration

When configuring an external IdP provider in Casdoor, you can now specify a `userMapping` configuration that maps IdP claims to user fields.

Example provider configuration:
```json
{
  "name": "my-okta-provider",
  "type": "Okta",
  "clientId": "your-client-id",
  "clientSecret": "your-client-secret",
  "domain": "https://your-domain.okta.com/oauth2/default",
  "userMapping": {
    "id": "sub",
    "username": "preferred_username",
    "displayName": "name",
    "email": "email",
    "avatarUrl": "picture",
    "phone": "phone_number",
    "countryCode": "country_code",
    "firstName": "given_name",
    "lastName": "family_name",
    "region": "region",
    "location": "locality",
    "affiliation": "organization",
    "title": "job_title",
    "language": "locale"
  }
}
```

### Frontend Configuration

In the Casdoor admin UI, when editing a provider:

1. Navigate to **Providers** section
2. Edit or create an OAuth/OIDC provider (e.g., Okta, Azure AD)
3. Scroll to the **User mapping** section
4. Configure the mapping between IdP claim names and Casdoor user fields

Each mapping field allows you to specify the claim name from your IdP that should be mapped to the corresponding user field.

## How It Works

1. When a user authenticates via an external IdP, Casdoor receives an access token
2. Casdoor fetches user information from the IdP's userinfo endpoint
3. The IdP returns claims in the userinfo response (e.g., `given_name`, `family_name`, `email`, etc.)
4. Casdoor captures all claims in the `Extra` field of the UserInfo structure
5. The `userMapping` configuration is applied to map specific claims to user fields
6. User fields are only updated if they are currently empty (existing values are preserved)
7. The user record is saved with the mapped information

## Common IdP Claim Names

### Okta
- `sub` - User ID
- `preferred_username` - Username
- `name` - Full name
- `given_name` - First name
- `family_name` - Last name
- `email` - Email address
- `phone_number` - Phone number
- `locale` - Preferred language
- `zoneinfo` - Time zone
- `picture` - Profile picture URL

### Azure AD
- `sub` or `oid` - User ID
- `preferred_username` or `upn` - Username
- `name` - Full name
- `given_name` - First name
- `family_name` - Last name
- `email` - Email address
- `mobile_phone` - Phone number
- `preferred_language` - Preferred language
- `country` - Country
- `city` - City
- `state` - State/Region
- `job_title` - Job title
- `company_name` - Company name

### Google
- `sub` - User ID
- `email` - Email address
- `name` - Full name
- `given_name` - First name
- `family_name` - Last name
- `picture` - Profile picture URL
- `locale` - Preferred language

## Important Notes

1. **Non-Overwriting**: The mapping only updates user fields that are currently empty. If a user already has a value for a field, it will not be overwritten by IdP claims.

2. **Standard Fields**: The standard fields (`id`, `username`, `displayName`, `email`, `avatarUrl`) are handled by the main authentication flow and should not be modified through the custom mapping function.

3. **Extra Claims**: All claims from the IdP are stored in the user's properties under `oauth_{provider}_extra` as a JSON string, allowing you to access any claim that wasn't explicitly mapped.

4. **Case Insensitive**: The field names in the user mapping are case-insensitive (e.g., `firstName`, `firstname`, and `FIRSTNAME` all work).

## Example Use Cases

### Use Case 1: Auto-populate User Profile from Okta
Configure your Okta provider to map common claims:
- Map `given_name` → `firstName`
- Map `family_name` → `lastName`
- Map `phone_number` → `phone`
- Map `locale` → `language`

When a user logs in via Okta for the first time, their profile will be automatically populated with these fields.

### Use Case 2: Sync Employee Information from Azure AD
Map Azure AD claims to employee fields:
- Map `job_title` → `title`
- Map `company_name` → `affiliation`
- Map `city` → `location`
- Map `state` → `region`
- Map `country` → `countryCode`

### Use Case 3: Custom Claims
If your IdP returns custom claims, you can map them as well:
- Map `employee_id` → `idCard`
- Map `department` → `tag`
- Map `manager_email` → Custom property (stored in Extra)

## Troubleshooting

### Claims Not Being Mapped
1. Check that the claim name in your `userMapping` matches exactly the claim name returned by your IdP
2. Verify that the IdP is actually returning the claim (check the `oauth_{provider}_extra` property)
3. Ensure the user field is empty (mapping doesn't overwrite existing values)

### Missing Claims from IdP
1. Check your IdP configuration to ensure the required scopes are requested (e.g., `profile`, `email`, `phone`)
2. Verify that the user has the claims configured in the IdP
3. Check the IdP documentation for available claims

### Field Names
1. Refer to the list of supported fields above
2. Field names are case-insensitive but should follow the camelCase convention
3. Standard fields are automatically handled by the authentication flow

## Technical Details

### Code Structure

- **IDP Providers** (`idp/okta.go`, `idp/azuread_b2c.go`, `idp/goth.go`): Enhanced to capture all claims from the IdP in the `Extra` field
- **User Util** (`object/user_util.go`): 
  - `SetUserOAuthPropertiesWithMapping` - Main function that applies user mapping
  - `applyUserMapping` - Helper function that performs the actual field mapping
- **Auth Controller** (`controllers/auth.go`): Updated to pass provider's `userMapping` configuration
- **Frontend** (`web/src/ProviderEditPage.js`): UI for configuring user field mappings

### Testing

Tests are available in `object/user_mapping_test.go` that verify:
- Basic mapping functionality
- Preservation of existing user values
- Handling of missing claims
- Proper skipping of standard fields

Run tests with:
```bash
go test ./object -run TestApplyUserMapping
```
