# Provider Category Components

This directory contains category-specific provider field components that have been extracted from the main `ProviderEditPage.js` file to improve code organization and maintainability.

## Purpose

The original `ProviderEditPage.js` was over 1900 lines and contained extensive if-else logic for different provider categories. This refactoring separates the category-specific UI rendering logic into dedicated component files while keeping the common logic (data fetching, saving, field updates) in the main page component.

## Components

### NotificationProviderFields.js
Handles UI fields for Notification category providers:
- Custom HTTP method selection
- Parameter configuration
- Metadata input
- Content editor
- Receiver configuration
- Test notification button

### EmailProviderFields.js
Handles UI fields for Email category providers:
- Host and port configuration
- SSL mode settings
- Proxy configuration
- HTTP method and headers (for Custom HTTP Email)
- Email content editor with preview
- Invitation email content
- Test email functionality

### SmsProviderFields.js
Handles UI fields for SMS category providers:
- Sign name configuration
- Template code
- HTTP method and headers (for Custom HTTP SMS)
- Phone number mapping
- Proxy settings
- Test SMS functionality

### MfaProviderFields.js
Handles UI fields for MFA (Multi-Factor Authentication) category providers:
- RADIUS server host and port
- Shared secret configuration

### SamlProviderFields.js
Handles UI fields for SAML category providers:
- Sign request toggle
- Metadata URL fetching
- Metadata parsing
- Endpoint configuration
- IdP certificate
- SP ACS URL and Entity ID

## Usage

Each component exports a render function that takes the provider data and callback functions as parameters:

```javascript
import {renderEmailProviderFields} from "./provider/EmailProviderFields";

// In the render method:
renderEmailProviderFields(
  provider,                           // provider object
  updateProviderField.bind(this),     // field update callback
  renderEmailMappingInput.bind(this), // mapping input renderer
  account                             // account object
)
```

## Benefits

1. **Reduced file size**: Main file reduced from 1971 to ~1400 lines (28% reduction)
2. **Improved maintainability**: Category-specific code is isolated and easier to find
3. **Better code organization**: Related functionality is grouped together
4. **Reduced complexity**: Fewer nested if-else statements in the main file
5. **Easier testing**: Category-specific components can be tested independently

## Future Improvements

Additional categories that could be extracted:
- OAuth provider fields (email regex, custom URLs, user mapping)
- Captcha provider fields (preview)
- Storage provider fields (type-specific configurations)
- Web3 provider fields (wallet selection)
- Payment provider fields (cert selection)
