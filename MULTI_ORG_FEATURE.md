# Multi-Organization Tenancy Feature

## Overview

This feature allows a single user account to be associated with multiple organizations simultaneously. This is useful for scenarios where users need to collaborate across different organizations, such as consultants working with multiple clients or doctors serving different hospitals.

## Implementation Details

### Database Schema

A new table `user_organization` has been added to track user membership in organizations:

```sql
CREATE TABLE user_organization (
    owner VARCHAR(100) NOT NULL,        -- User's owner (original organization)
    name VARCHAR(100) NOT NULL,         -- User's name
    organization VARCHAR(100) NOT NULL, -- Organization the user is a member of
    created_time VARCHAR(100),          -- When the membership was created (stored as string for consistency with existing schema)
    is_default BOOLEAN,                 -- Is this the user's primary organization
    PRIMARY KEY (owner, name, organization)
);
```

**Note**: The `created_time` field uses VARCHAR(100) to maintain consistency with the existing Casdoor schema design pattern, where timestamps are stored as formatted strings throughout the application.

### Backend Changes

1. **UserOrganization Model** (`object/user_organization.go`):
   - Represents the many-to-many relationship between users and organizations
   - Methods for adding, removing, and querying user-organization relationships

2. **JWT Token Claims** (`object/token_jwt.go`):
   - Added `OrganizationContext` field to Claims struct
   - Tokens now include which organization context the user is operating in

3. **Invitation System** (`object/invitation_org.go`):
   - `AcceptOrganizationInvitation` allows existing users to join additional organizations
   - Validates invitation codes and creates user-organization relationships

4. **API Endpoints** (`controllers/user_organization.go`):
   - `GET /api/get-user-organizations` - Get all organizations a user belongs to
   - `POST /api/add-user-to-organization` - Add a user to an organization
   - `POST /api/remove-user-from-organization` - Remove a user from an organization
   - `POST /api/set-organization-context` - Set the active organization context
   - `GET /api/get-organization-context` - Get the active organization context
   - `POST /api/accept-organization-invitation` - Accept an invitation to join an organization

5. **Session Management** (`controllers/base.go`):
   - Extended SessionData to include `OrganizationContext`
   - Session tracks which organization the user is currently operating as

6. **Data Migration** (`object/user_organization_migration.go`):
   - Automatically creates default organization memberships for existing users
   - Runs on application startup

### Workflow

#### 1. User Creation
When a new user is created, they are automatically added to their owner organization with `is_default = true`.

#### 2. Invitation to Additional Organizations
1. Admin of Organization B creates an invitation code
2. Admin sends invitation code to user (e.g., via email)
3. User (already member of Organization A) accepts invitation using `/api/accept-organization-invitation`
4. User is now a member of both Organization A and Organization B

#### 3. Organization Context Switching
1. User logs in
2. User can call `/api/get-user-organizations` to see all organizations they belong to
3. User selects desired organization and calls `/api/set-organization-context`
4. Subsequent JWT tokens will include the selected organization in the `organization` claim
5. Applications can use this claim to enforce permissions and show organization-specific data

### JWT Token Structure

When a token is generated, it includes the organization context:

```json
{
  "owner": "org-a",
  "name": "user1",
  "organization": "org-b",
  "iss": "https://casdoor.example.com",
  "sub": "user-id",
  "aud": ["client-id"],
  "exp": 1234567890,
  ...
}
```

The `organization` claim indicates which organization the user is currently operating as.

## Frontend Integration (To Be Implemented)

A frontend UI component should be added to allow users to:
1. View all organizations they belong to
2. Switch between organizations
3. See which organization they are currently operating as
4. Accept invitations to join new organizations

## API Usage Examples

### Get User's Organizations
```bash
curl -X GET "https://casdoor.example.com/api/get-user-organizations?id=org-a/user1" \
  -H "Authorization: Bearer TOKEN"
```

### Switch Organization Context
```bash
curl -X POST "https://casdoor.example.com/api/set-organization-context?organization=org-b" \
  -H "Authorization: Bearer TOKEN"
```

### Accept Organization Invitation
```bash
curl -X POST "https://casdoor.example.com/api/accept-organization-invitation?invitationCode=ABC123&organization=org-b" \
  -H "Authorization: Bearer TOKEN"
```

## Security Considerations

1. Users can only switch to organizations they are members of
2. Primary organization (is_default=true) cannot be removed
3. Invitation codes are validated before creating organization memberships
4. Organization context is stored in secure session data
5. JWT tokens clearly identify the organization context

## Testing

Unit tests are provided in `object/user_organization_test.go` to verify:
- **UserOrganization struct creation** - Validates proper initialization of the model
- **GetId method** - Returns unique identifier for user-organization relationship in format "owner/name/organization"
- **Basic CRUD operations** - Tests for adding, retrieving, and deleting relationships (when database is available)

Integration tests can be added by uncommenting the test functions that require database connectivity.
