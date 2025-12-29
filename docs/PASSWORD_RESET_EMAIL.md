# Password Reset Email Template Guide

## Overview

Casdoor supports password reset via email using either:
1. **Verification codes** (legacy method)
2. **Magic links** (new method)

## Email Template Format

### Using Magic Links (Recommended)

To enable magic link password resets, your email provider's template should include the `<reset-link>` tag:

```html
<!DOCTYPE html>
<html>
<body>
  <p>Hi %{user.friendlyName},</p>
  
  <p>You're receiving this email because we received a password reset request for your account.</p>
  
  <p>
    <reset-link>Reset password</reset-link>
  </p>
  
  <p>This password reset link will expire within 24 hours.</p>
  
  <p>If you didn't request this or need assistance, reach out to your admin or internal support team immediately.</p>
</body>
</html>
```

### Template Tags

- **`%{user.friendlyName}`**: Replaced with the user's friendly name
- **`<reset-link>...</reset-link>`**: Content within these tags becomes a clickable link
  - The system automatically wraps this content in an `<a>` tag
  - Example: `<reset-link>Reset password</reset-link>` becomes `<a href="...">Reset password</a>`
- **`%link`**: (Optional) Direct URL placeholder if you want to manually create the link

### Using Verification Codes (Legacy)

If your template doesn't include `<reset-link>` tags, Casdoor will fall back to the verification code method:

```html
<!DOCTYPE html>
<html>
<body>
  <p>Hi %{user.friendlyName},</p>
  
  <p>You have requested a verification code at Casdoor.</p>
  
  <p>Here is your code: <strong>%s</strong></p>
  
  <p>Please enter this code within 5 minutes.</p>
  
  <p>You can also use this link: %link</p>
</body>
</html>
```

- **`%s`**: Replaced with the 6-digit verification code

## How It Works

### Magic Link Flow

1. User clicks "Forgot Password" on login page
2. User enters their username/email
3. System generates a secure token (cryptographically random)
4. Email is sent with a magic link containing the token
5. User clicks the link and is taken directly to the password reset form
6. Token is validated (must be unused and within 24 hours)
7. User enters new password and submits
8. Password is updated and user is redirected to login

### Security Features

- **Secure Token**: Uses `crypto/rand` for cryptographically secure random tokens
- **24-Hour Expiration**: Tokens automatically expire after 24 hours
- **Single Use**: Tokens can only be used once
- **Session Validation**: Token validation creates a secure session for password reset

### Backward Compatibility

The implementation maintains full backward compatibility:
- Email templates without `<reset-link>` tags use verification codes
- Existing verification code flow continues to work
- Both methods can coexist in the same installation

## Configuration

### Email Provider Setup

1. Go to **Providers** in Casdoor admin panel
2. Select or create an Email provider
3. Set the **Category** to "Email"
4. Configure the **Title** (email subject)
5. Set the **Content** with your HTML template using the tags above
6. Save the provider

### Application Setup

1. Go to **Applications** in Casdoor admin panel
2. Select your application
3. In the **Providers** section, add your Email provider
4. Ensure it's enabled for password reset (forget method)

## Example Templates

### Simple Magic Link Template

```html
<p>Hello %{user.friendlyName},</p>
<p>Click the button below to reset your password:</p>
<p><reset-link>Reset My Password</reset-link></p>
<p>This link expires in 24 hours.</p>
```

### Rich HTML Template

```html
<!DOCTYPE html>
<html>
<head>
  <style>
    .button {
      background-color: #4CAF50;
      border: none;
      color: white;
      padding: 15px 32px;
      text-align: center;
      text-decoration: none;
      display: inline-block;
      font-size: 16px;
      margin: 4px 2px;
      cursor: pointer;
      border-radius: 4px;
    }
  </style>
</head>
<body>
  <h2>Password Reset</h2>
  <p>Hi %{user.friendlyName},</p>
  <p>You're receiving this email because we received a password reset request for your account.</p>
  <p>This password reset link will expire within <strong>24 hours</strong>.</p>
  <p style="text-align: center;">
    <reset-link>
      <span class="button">Reset Password</span>
    </reset-link>
  </p>
  <p>If you didn't request this or need assistance, reach out to your admin or internal support team immediately.</p>
</body>
</html>
```

## Troubleshooting

### Magic Link Not Working

1. Verify the email template contains `<reset-link>` tags
2. Check the Email provider is properly configured
3. Ensure the application has the Email provider enabled
4. Verify the token hasn't expired (24 hours)
5. Check if the token has already been used

### Token Validation Errors

- **"The verification link is invalid or has expired!"**: Token is either invalid, expired (>24 hours), or already used
- **"The verification link has already been used!"**: Token has been used previously
- **"The user does not exist"**: User associated with the token no longer exists

### Verification Code Fallback

If you want to force verification code mode instead of magic links, simply remove the `<reset-link>` tags from your email template.

## API Reference

### Verify Reset Token Endpoint

**POST** `/api/verify-reset-token`

**Request:**
```
Content-Type: application/x-www-form-urlencoded

token=<base64-encoded-token>
```

**Response (Success):**
```json
{
  "status": "ok",
  "data": "username",
  "data2": "user@email.com"
}
```

**Response (Error):**
```json
{
  "status": "error",
  "msg": "The verification link is invalid or has expired!"
}
```
