# Profile Completion After SSO Login

## Overview

Casdoor now supports requiring users to complete missing profile information after Single Sign-On (SSO) login before being redirected to the application. This is useful when OAuth/SAML providers return partial user information (e.g., only email) and you need additional fields like phone number, display name, etc.

## How It Works

When a user signs in via an OAuth or SAML provider:

1. Casdoor receives user information from the provider (email, name, etc.)
2. If the application has "prompted" signup items configured, the user is redirected to a profile completion page
3. The user must fill in all required prompted fields before being redirected to the application
4. Once completed, the user is redirected with the OAuth authorization code

## Configuration

### Step 1: Configure Signup Items in Application

1. Navigate to **Applications** in the Casdoor admin panel
2. Select the application you want to configure
3. Go to the **Signup items** section
4. For each field you want to prompt after SSO login:
   - Check **Visible** - Makes the field available
   - Check **Required** - Makes it mandatory (optional)
   - Check **Prompted** - Shows it on the profile completion page after SSO

### Example Configuration

To require phone number after SSO login when only email is provided:

| Name  | Visible | Required | Prompted |
|-------|---------|----------|----------|
| Phone | ✓       | ✓        | ✓        |

### Supported Signup Items for Prompting

The following signup items can be prompted after SSO login:

- **Display name** - User's display name
- **First name** - User's first name
- **Last name** - User's last name  
- **Email** - User's email address (with validation)
- **Phone** - User's phone number (with country code selector and validation)
- **Affiliation** - User's organization/affiliation
- **Country/Region** - User's country or region
- **ID card** - User's ID card number

## User Experience Flow

### Standard OAuth Flow (without profile completion)
```
User → OAuth Provider → Casdoor Login → Application Redirect
```

### OAuth Flow with Profile Completion
```
User → OAuth Provider → Casdoor Login → Profile Completion Page → Application Redirect
```

### Profile Completion Page Features

- Clean, user-friendly interface with form validation
- Required fields are marked and validated before submission
- Phone number includes country code selector
- Email validation ensures proper format
- Custom labels and placeholders can be configured per field
- Regex validation support for custom field rules

## Example Scenarios

### Scenario 1: Collect Phone Number After Email-Only SSO

**Problem**: Google OAuth only returns email, but you need phone numbers for 2FA.

**Solution**: 
1. Enable Google OAuth provider
2. Set Phone signup item as: Visible=✓, Required=✓, Prompted=✓
3. Users will be asked to enter their phone number after Google login

### Scenario 2: Complete User Profile After Corporate SAML

**Problem**: Corporate SAML only provides username, need full profile.

**Solution**:
1. Configure SAML provider
2. Set multiple signup items as Prompted:
   - Display name: Visible=✓, Required=✓, Prompted=✓
   - Email: Visible=✓, Required=✓, Prompted=✓
   - Phone: Visible=✓, Required=✓, Prompted=✓
3. Users complete their full profile after SAML login

### Scenario 3: Optional Regional Information

**Problem**: Want to collect user's country/region but don't want to make it mandatory.

**Solution**:
1. Set Country/Region signup item as: Visible=✓, Required=✗, Prompted=✓
2. Users can optionally provide their region, or skip and continue

## Technical Details

### Backend

The backend already supports profile completion through:
- `SignupItem.Prompted` field in application configuration
- `HasPromptPage()` function checks if any items are prompted
- `getAllPromptedSignupItems()` returns list of prompted items
- Session management allows the prompt page to access user context

### Frontend

The frontend implementation includes:
- **PromptPage Component** (`web/src/auth/PromptPage.js`)
  - Renders form with all prompted signup items
  - Handles form validation before submission
  - Updates user profile via API
  - Redirects to application after completion

- **Field Validation** (`web/src/Setting.js`)
  - `isSignupItemPrompted()` - Checks if item should be shown on prompt page
  - `isSignupItemAnswered()` - Validates if required fields are filled
  - `isPromptAnswered()` - Checks if all prompted fields are completed

### API Endpoints

- `GET /api/get-account` - Retrieves current user information
- `POST /api/update-user` - Updates user profile with prompted fields
- Existing OAuth/SAML callback endpoints handle redirect to prompt page

## Best Practices

1. **Keep it minimal** - Only prompt for essential information to avoid user friction
2. **Use clear labels** - Configure custom labels that clearly explain what's needed
3. **Combine with email/phone verification** - Can be used with verification codes if needed
4. **Test the flow** - Always test the complete OAuth flow after configuration changes
5. **Consider user experience** - Too many required fields may discourage signups

## Troubleshooting

### Users Not Seeing Prompt Page

**Check:**
1. Signup items have `Prompted` checkbox enabled
2. Application has `EnableSignUp` enabled
3. Provider has `CanSignUp` enabled in provider configuration
4. User doesn't already have the prompted fields filled

### Validation Errors

**Check:**
1. Required fields are properly configured
2. Regex patterns (if configured) are valid
3. Email/phone format matches validation rules
4. Country codes are properly configured for phone validation

### Redirect Issues

**Check:**
1. OAuth redirect URI is properly configured in application
2. State parameter matches between login and prompt page
3. User session is active during prompt page interaction

## Migration from Previous Versions

If you were previously using custom profile completion logic:

1. Remove custom code/redirects
2. Configure signup items with `Prompted` flag as described above
3. Test the built-in prompt page functionality
4. Customize field labels and placeholders as needed

## Future Enhancements

Potential future improvements to this feature:

- Support for custom signup item types
- Multi-step profile completion wizard
- Conditional field prompting based on provider type
- Profile completion progress indicator
- Skip option for optional fields with clear UI

## Support

For issues or questions:
- GitHub Issues: https://github.com/casdoor/casdoor/issues
- Discord: https://discord.gg/5rPsrAzK7S
- Documentation: https://casdoor.org/docs
