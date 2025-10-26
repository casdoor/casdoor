# Multi-Organization Tenancy - Usage Examples

This document provides practical examples of how to use the multi-organization tenancy feature.

## Scenario: Doctor Working for Multiple Hospitals

### Setup
- **Hospital A** (organization: `hospital-a`)
- **Hospital B** (organization: `hospital-b`)
- **Dr. Smith** (user: `hospital-a/dr-smith`, email: `dr.smith@example.com`)

### Step 1: Initial User Creation
Dr. Smith is initially created in Hospital A:

```bash
POST /api/add-user
{
  "owner": "hospital-a",
  "name": "dr-smith",
  "email": "dr.smith@example.com",
  ...
}
```

At this point:
- Dr. Smith is a member of `hospital-a` (default)
- `user_organization` table contains: `(hospital-a, dr-smith, hospital-a, ..., true)`

### Step 2: Hospital B Invites Dr. Smith

Hospital B admin creates an invitation:

```bash
POST /api/add-invitation
{
  "owner": "hospital-b",
  "name": "inv-dr-smith",
  "code": "HOSP-B-DR-SMITH-2025",
  "email": "dr.smith@example.com",
  "state": "Active",
  "quota": 1
}
```

### Step 3: Dr. Smith Accepts Invitation

Dr. Smith logs in and accepts the invitation:

```bash
POST /api/accept-organization-invitation?invitationCode=HOSP-B-DR-SMITH-2025&organization=hospital-b
Authorization: Bearer <dr-smith-token>
```

Response:
```json
{
  "status": "ok",
  "data": true
}
```

Now:
- Dr. Smith is a member of both `hospital-a` and `hospital-b`
- `user_organization` table contains:
  - `(hospital-a, dr-smith, hospital-a, ..., true)` - primary
  - `(hospital-a, dr-smith, hospital-b, ..., false)` - additional

### Step 4: View All Organizations

Dr. Smith can see all organizations they belong to:

```bash
GET /api/get-user-organizations?id=hospital-a/dr-smith
Authorization: Bearer <dr-smith-token>
```

Response:
```json
{
  "status": "ok",
  "data": [
    {
      "owner": "hospital-a",
      "name": "dr-smith",
      "organization": "hospital-a",
      "createdTime": "2025-01-15T10:00:00Z",
      "isDefault": true
    },
    {
      "owner": "hospital-a",
      "name": "dr-smith",
      "organization": "hospital-b",
      "createdTime": "2025-01-20T14:30:00Z",
      "isDefault": false
    }
  ]
}
```

### Step 5: Switch Organization Context

When Dr. Smith wants to work on Hospital B's systems:

```bash
POST /api/set-organization-context?organization=hospital-b
Authorization: Bearer <dr-smith-token>
```

Response:
```json
{
  "status": "ok",
  "data": "hospital-b"
}
```

### Step 6: Get New Token with Organization Context

After switching, when Dr. Smith requests a new access token, it includes the organization context:

```bash
POST /api/oauth/token
{
  "grant_type": "password",
  "client_id": "hospital-app",
  "username": "dr-smith",
  "password": "***"
}
```

The resulting JWT token payload includes:
```json
{
  "owner": "hospital-a",
  "name": "dr-smith",
  "organization": "hospital-b",  // <-- Current context
  "email": "dr.smith@example.com",
  "iss": "https://auth.example.com",
  "sub": "user-id-123",
  "aud": ["hospital-app"],
  "exp": 1737896400
}
```

### Step 7: Application Uses Organization Context

The hospital application can now:

```javascript
// Decode JWT token
const token = decodeJWT(accessToken);
const currentOrg = token.organization; // "hospital-b"

// Fetch hospital-specific data
if (currentOrg === "hospital-a") {
  // Show Hospital A patient records, schedules, etc.
  loadHospitalAData();
} else if (currentOrg === "hospital-b") {
  // Show Hospital B patient records, schedules, etc.
  loadHospitalBData();
}
```

## Scenario: Consultant Working with Multiple Clients

### Setup
- **Client A** (organization: `client-a`)
- **Client B** (organization: `client-b`)
- **Client C** (organization: `client-c`)
- **Consultant Jane** (user: `consulting-firm/jane`, email: `jane@consulting.com`)

### Workflow

1. **Initial Setup**: Jane is created in `consulting-firm`
2. **Client A Project**: 
   - Client A admin invites Jane with code `PROJECT-2025-CLIENT-A`
   - Jane accepts invitation
   - Jane switches context to `client-a` when working on Client A project
   
3. **Client B Project**:
   - Client B admin invites Jane with code `PROJECT-2025-CLIENT-B`
   - Jane accepts invitation
   - Jane switches context to `client-b` when working on Client B project

4. **Daily Work**:
   - Morning: Switch to `client-a`, work on Client A project
   - Afternoon: Switch to `client-b`, work on Client B project
   - Each context switch provides appropriate JWT token for accessing client resources

## Admin Operations

### Check User Memberships

```bash
GET /api/get-user-organizations?id=hospital-a/dr-smith
```

### Add User to Organization Directly (Admin)

```bash
POST /api/add-user-to-organization
{
  "owner": "hospital-a",
  "name": "dr-smith",
  "organization": "hospital-c",
  "createdTime": "2025-01-25T09:00:00Z",
  "isDefault": false
}
```

### Remove User from Organization

```bash
POST /api/remove-user-from-organization?owner=hospital-a&name=dr-smith&organization=hospital-c
```

Note: Cannot remove primary organization (isDefault=true)

## Security Best Practices

1. **Validate Organization Context**: Applications should verify the user is authorized for the organization specified in the JWT token
2. **Audit Logging**: Log organization context switches for security monitoring
3. **Data Isolation**: Use organization context to enforce strict data isolation
4. **Permission Checks**: Combine organization context with role/permission checks

## Troubleshooting

### User Cannot Accept Invitation
- Check invitation is Active
- Verify invitation email matches user's email (if specified)
- Ensure invitation quota not exceeded
- Confirm organization exists

### Organization Context Not Switching
- Verify user is member of target organization
- Check session is valid
- Ensure new token is requested after context switch

### JWT Missing Organization Claim
- Verify token was generated after implementing this feature
- Check token generation includes organization context parameter
- Ensure user has active session with organization context
