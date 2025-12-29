# Testing Password Reset via Email Link

## Prerequisites

1. Casdoor instance running
2. Email provider configured (SMTP recommended for testing)
3. User account with email address

## Test Plan

### Test 1: Magic Link Password Reset (Happy Path)

**Setup:**
1. Configure an Email provider with the following template:

```html
<p>Hi %{user.friendlyName},</p>
<p>You're receiving this email because we received a password reset request for your account.</p>
<p><reset-link>Reset your password</reset-link></p>
<p>This link expires in 24 hours.</p>
```

2. Assign the Email provider to your application

**Steps:**
1. Navigate to login page
2. Click "Forgot Password"
3. Enter username
4. Click "Next Step"
5. Verify email is sent with magic link
6. Click the magic link in email
7. Verify you're taken directly to password reset form (Step 3)
8. Enter new password and confirm
9. Click "Change Password"
10. Verify password is updated
11. Login with new password

**Expected Results:**
- ✅ Email received with clickable link
- ✅ Link redirects to password reset page
- ✅ No verification code input required
- ✅ Password successfully updated
- ✅ Login works with new password

### Test 2: Verification Code Flow (Backward Compatibility)

**Setup:**
1. Update Email provider template to remove `<reset-link>` tags:

```html
<p>Hi %{user.friendlyName},</p>
<p>Your verification code is: %s</p>
<p>Please enter this code within 5 minutes.</p>
```

**Steps:**
1. Follow steps 1-4 from Test 1
2. Verify email is sent with verification code
3. Copy the verification code
4. Return to browser and select email/phone
5. Paste verification code
6. Click "Next Step"
7. Enter new password
8. Click "Change Password"

**Expected Results:**
- ✅ Email received with 6-digit code
- ✅ Code can be entered manually
- ✅ Password successfully updated
- ✅ Original flow still works

### Test 3: Token Expiration

**Setup:**
1. Use magic link template from Test 1

**Steps:**
1. Request password reset
2. Receive email with magic link
3. Extract token from URL
4. Manually modify verification record in database to set time to 25 hours ago:
   ```sql
   UPDATE verification_record 
   SET time = UNIX_TIMESTAMP(NOW() - INTERVAL 25 HOUR) 
   WHERE code = 'your-token-here';
   ```
5. Click the magic link

**Expected Results:**
- ✅ Error message: "The verification link is invalid or has expired!"
- ✅ Redirected back to Step 1

### Test 4: Token Single Use

**Setup:**
1. Use magic link template from Test 1

**Steps:**
1. Request password reset
2. Receive email with magic link
3. Click the link
4. Complete password reset
5. Click the same link again

**Expected Results:**
- ✅ First click: Successfully reset password
- ✅ Second click: Error message "The verification link has already been used!"

### Test 5: Invalid Token

**Steps:**
1. Manually construct URL with fake token:
   `http://localhost:7001/forget/app-built-in?token=invalid-token-12345`
2. Navigate to the URL

**Expected Results:**
- ✅ Error message: "The verification link is invalid or has expired!"

### Test 6: Security - Token Format

**Steps:**
1. Request password reset with magic link
2. Inspect the token in the email link
3. Verify token characteristics:
   - Length > 40 characters (32 bytes base64-encoded)
   - Contains URL-safe characters
   - Appears random

**Expected Results:**
- ✅ Token is cryptographically random
- ✅ Token is URL-safe base64 encoded
- ✅ Token is sufficiently long

### Test 7: Concurrent Requests

**Steps:**
1. Request password reset
2. Receive first email
3. Request password reset again immediately
4. Receive second email
5. Click link from first email
6. Click link from second email

**Expected Results:**
- ✅ Both emails received
- ✅ Both tokens are different
- ✅ First token works
- ✅ Second token shows error (already used) OR works if code verification flow allows multiple

### Test 8: Cross-Browser Session

**Steps:**
1. Request password reset in Browser A
2. Receive email with magic link
3. Open link in Browser B (different session)
4. Complete password reset in Browser B

**Expected Results:**
- ✅ Password reset works in different browser
- ✅ Session is properly created in Browser B

## Manual Testing Checklist

- [ ] Test 1: Magic Link Happy Path
- [ ] Test 2: Verification Code Backward Compatibility
- [ ] Test 3: Token Expiration
- [ ] Test 4: Token Single Use
- [ ] Test 5: Invalid Token
- [ ] Test 6: Token Security Format
- [ ] Test 7: Concurrent Requests
- [ ] Test 8: Cross-Browser Session

## Email Provider Configuration

### SMTP Provider Example (for testing)

```json
{
  "owner": "admin",
  "name": "provider_email_test",
  "type": "SMTP",
  "category": "Email",
  "host": "smtp.gmail.com",
  "port": 587,
  "clientId": "your-email@gmail.com",
  "clientSecret": "your-app-password",
  "title": "Reset your password",
  "content": "<p>Hi %{user.friendlyName},</p><p>Click below to reset your password:</p><p><reset-link>Reset Password</reset-link></p><p>This link expires in 24 hours.</p>"
}
```

### Gmail App Password Setup

1. Enable 2-factor authentication
2. Go to Google Account > Security > 2-Step Verification > App passwords
3. Generate new app password
4. Use this password as `clientSecret`

## Common Issues

### Email Not Received
- Check spam folder
- Verify SMTP credentials
- Check email provider logs
- Verify user has email address

### Link Not Working
- Check if `<reset-link>` tags are present in template
- Verify application has email provider assigned
- Check browser console for errors
- Verify token hasn't expired

### Token Validation Fails
- Check database for verification_record entry
- Verify token matches exactly
- Check token hasn't been marked as used
- Verify time hasn't exceeded 24 hours

## Database Inspection

To inspect verification records:

```sql
-- View recent verification records
SELECT * FROM verification_record 
WHERE type = 'Email' 
ORDER BY time DESC 
LIMIT 10;

-- Check specific token
SELECT *, FROM_UNIXTIME(time) as created_at 
FROM verification_record 
WHERE code = 'your-token-here';

-- Check token age
SELECT *, 
  (UNIX_TIMESTAMP(NOW()) - time) / 3600 as hours_old,
  is_used
FROM verification_record 
WHERE code = 'your-token-here';
```

## Success Criteria

All tests should pass with:
- ✅ Magic links work correctly
- ✅ Verification codes still work (backward compatibility)
- ✅ Security features enforced (expiration, single-use)
- ✅ Error messages are clear and helpful
- ✅ No security vulnerabilities introduced
