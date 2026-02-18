# Verification Guide: Invitation Links with OAuth Providers

## Issue Description
Previously, when users signed up using invitation links with OAuth providers (like Google), they were not automatically assigned to the specified signup group. This worked correctly for email+password signup but failed for OAuth signup.

## Root Cause
When `SetUserOAuthProperties` called `UpdateUserForAllFields` during OAuth signup, groups were written to the database but not synchronized to the Casbin enforcer. This caused the groups to appear in the database but not be properly enforced by the permission system.

## Fix Implementation (PR #5123)
Added group enforcer synchronization in `UpdateUserForAllFields` function (object/user.go lines 957-962):

```go
if len(user.Groups) > 0 {
    _, err = userEnforcer.UpdateGroupsForUser(user.GetId(), user.Groups)
    if err != nil {
        return false, err
    }
}
```

This ensures that groups are properly synchronized to the Casbin enforcer whenever user data is updated, including during OAuth signup.

## Verification Steps

### Prerequisites
1. Casdoor v2.324.0 or later (includes PR #5123 fix)
2. An OAuth provider configured (e.g., Google)
3. An invitation with a signup group configured

### Test Scenario 1: Email + Password Signup (Should Work)
1. Create an invitation with a signup group (e.g., "TestGroup")
2. Get the invitation link (e.g., `/signup/app?invitationCode=ABC123`)
3. Sign up using email and verification code
4. Verify the user is assigned to "TestGroup"

### Test Scenario 2: OAuth Provider Signup (Should Now Work)
1. Use the same invitation link from Scenario 1
2. Click on the OAuth provider button (e.g., "Sign in with Google")
3. Complete OAuth authentication
4. Verify the user is created and assigned to "TestGroup"

### Verification Methods

#### Method 1: Check User Groups in UI
1. Log in as admin
2. Go to Users page
3. Find the newly created user
4. Check the "Groups" field - should show "TestGroup"

#### Method 2: Check Casbin Enforcer
```go
// Verify group is in enforcer
groups := userEnforcer.GetGroupsForUser(user.GetId())
// Should include "TestGroup"
```

#### Method 3: Check Database
```sql
-- Check user's groups in database
SELECT owner, name, groups FROM user WHERE name = 'oauth_user_name';
-- groups column should contain ["TestGroup"]
```

## Code Flow

### OAuth Signup with Invitation Code
1. **controllers/auth.go:876** - Invitation code is validated, invitation object retrieved
2. **controllers/auth.go:958-962** - Groups assigned from invitation.SignupGroup or providerItem.SignupGroup
3. **controllers/auth.go:965** - `AddUser(user)` called
   - **object/user.go:1069-1074** - Groups synced to Casbin enforcer
   - **object/user.go:1081** - User inserted to database
4. **controllers/auth.go:988** - `SetUserOAuthProperties()` called
   - **object/user_util.go:203-283** - OAuth properties updated
   - **object/user_util.go:285** - `UpdateUserForAllFields()` called
     - **object/user.go:957-962** - Groups synced to enforcer again (PR #5123 fix)
     - **object/user.go:964** - All columns updated in database

### Key Points
- Groups are set before `AddUser` is called
- The same user object (with groups) is passed through all function calls
- `SetUserOAuthProperties` does not modify the Groups field
- `UpdateUserForAllFields` now properly syncs groups to the enforcer

## Expected Behavior After Fix
- ✅ Email+password signup with invitation: User assigned to signup group
- ✅ OAuth provider signup with invitation: User assigned to signup group
- ✅ Groups visible in UI
- ✅ Groups enforced by permission system
- ✅ Groups persisted in database

## Troubleshooting

### If groups are still not assigned after OAuth signup:
1. Verify you're running Casdoor v2.324.0 or later
2. Check that the invitation has a signup group configured
3. Verify the OAuth provider is properly configured in the application
4. Check server logs for any errors during signup
5. Ensure the Casbin enforcer is properly initialized

### Debug Logging
Add logging to verify the fix:
```go
// In UpdateUserForAllFields (object/user.go:957)
log.Printf("UpdateUserForAllFields: user=%s, groups=%v", user.GetId(), user.Groups)
```

## Related Issues
- Issue #5122: Groups are still not assigned when registering through external providers using an invitation code
- PR #5123: feat: fix group assignment for OAuth provider signup with invitation codes
- Issue #4960: Original report of invitation code group assignment issue
- PR #4961: Initial attempted fix

## Conclusion
PR #5123 should fully resolve the issue where OAuth provider signups with invitation codes don't assign users to the specified signup group. The fix ensures that groups are properly synchronized to the Casbin enforcer whenever user data is updated, which was the missing piece in the previous implementation.
