# ID Verification (IDV) Providers

This package provides ID verification functionality for Casdoor, allowing organizations to verify user identities through third-party ID verification services.

## Overview

ID Verification providers enable Casdoor to verify user identities by integrating with external identity verification services. This is useful for compliance, KYC (Know Your Customer), and enhanced security requirements.

## Supported Providers

### Jumio

Jumio is a leading identity verification service that provides real-time ID verification and authentication.

**Configuration:**
- **Client ID**: Your Jumio API Token (also known as API Key)
- **Client Secret**: Your Jumio API Secret
- **Endpoint**: Jumio API endpoint (default: `https://api.jumio.com`)

**Features:**
- Identity document verification (passport, driver's license, ID card, etc.)
- Real-time verification status
- Support for multiple document types and countries
- Test connection functionality

**Usage:**
1. Create a Jumio account and obtain API credentials
2. In Casdoor, create a new Provider with category "ID Verification"
3. Select "Jumio" as the provider type
4. Enter your API credentials
5. Use the "Test Connection" button to verify the configuration

## Adding New ID Verification Providers

To add a new ID verification provider:

1. Create a new file in the `idv` package (e.g., `newprovider.go`)
2. Implement the `IdvProvider` interface:
   ```go
   type IdvProvider interface {
       VerifyIdentity(request *VerificationRequest) (*VerificationResult, error)
       GetVerificationStatus(transactionID string) (*VerificationResult, error)
       TestConnection() error
   }
   ```
3. Add the provider to the `GetIdvProvider` function in `provider.go`
4. Update the frontend to include the new provider type in `Setting.js`

## API Usage

### Verify Identity

```go
provider, err := idv.GetIdvProvider("Jumio", clientId, clientSecret, endpoint)
if err != nil {
    // Handle error
}

request := &idv.VerificationRequest{
    FirstName:    "John",
    LastName:     "Doe",
    DateOfBirth:  "1990-01-01",
    Country:      "USA",
    IdCardType:   "PASSPORT",
    IdCardNumber: "123456789",
}

result, err := provider.VerifyIdentity(request)
if err != nil {
    // Handle error
}

// Check result.Verified status
```

### Check Verification Status

```go
result, err := provider.GetVerificationStatus(transactionID)
if err != nil {
    // Handle error
}

// Check result.Verified status
```

### Test Connection

```go
err := provider.TestConnection()
if err != nil {
    // Handle error
}
```

## License

This package is part of Casdoor and is licensed under the Apache License 2.0.
