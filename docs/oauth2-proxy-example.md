# Example: Using Casdoor URL Protection with oauth2-proxy

This example demonstrates how to configure Casdoor's URL protection feature when using oauth2-proxy.

## Scenario

You have an application with the following requirements:
- Base URL: `https://app.example.com`
- Protected endpoints:
  - `/api/users` - Requires authentication
  - `/api/admin/*` - Requires authentication
- Public endpoints:
  - `/api/public/*` - No authentication required
  - `/health` - Health check endpoint
  - `/metrics` - Metrics endpoint (for monitoring)

## Casdoor Configuration

In your Casdoor application settings:

### Protected URIs
Add these patterns:
```
https://app\.example\.com/api/users
https://app\.example\.com/api/admin/.*
```

### Public URIs
Add these patterns:
```
https://app\.example\.com/api/public/.*
https://app\.example\.com/health$
https://app\.example\.com/metrics$
```

## oauth2-proxy Configuration

Example `oauth2-proxy.cfg`:

```ini
# OAuth2 Proxy Configuration
http_address = "0.0.0.0:4180"
upstreams = [
    "http://localhost:8080/"
]

# Casdoor Provider Settings
provider = "oidc"
oidc_issuer_url = "https://door.casdoor.com"
client_id = "your-client-id"
client_secret = "your-client-secret"
redirect_url = "https://app.example.com/oauth2/callback"

# Cookie settings
cookie_secret = "your-cookie-secret"
cookie_secure = true
cookie_httponly = true

# Skip authentication for public endpoints
skip_auth_regex = [
    "^/api/public/.*",
    "^/health$",
    "^/metrics$"
]

# Email domain restrictions (optional)
email_domains = [
    "*"
]
```

## Using the IsUriProtected Method

If you're building custom middleware or authorization logic, you can use Casdoor's `IsUriProtected` method:

```go
package main

import (
    "fmt"
    "github.com/casdoor/casdoor/object"
)

func checkAccess(uri string) {
    app, err := object.GetApplication("admin/my-app")
    if err != nil {
        fmt.Printf("Error getting application: %v\n", err)
        return
    }

    if app.IsUriProtected(uri) {
        fmt.Printf("%s requires authentication\n", uri)
    } else {
        fmt.Printf("%s is publicly accessible\n", uri)
    }
}

func main() {
    // Test various URIs
    checkAccess("https://app.example.com/api/users")        // Protected
    checkAccess("https://app.example.com/api/admin/users")  // Protected
    checkAccess("https://app.example.com/api/public/data")  // Public
    checkAccess("https://app.example.com/health")           // Public
    checkAccess("https://app.example.com/metrics")          // Public
}
```

## Testing Your Configuration

1. **Test Protected Endpoint**:
   ```bash
   curl -I https://app.example.com/api/users
   # Should redirect to login page
   ```

2. **Test Public Endpoint**:
   ```bash
   curl -I https://app.example.com/api/public/data
   # Should return 200 OK without authentication
   ```

3. **Test Health Check**:
   ```bash
   curl https://app.example.com/health
   # Should return health status without authentication
   ```

## Common Patterns

### Pattern 1: Protect Everything Except Specific Paths
```
Protected URIs: (leave empty)
Public URIs:
  - .*/health$
  - .*/metrics$
  - .*/api/public/.*
```

### Pattern 2: Protect Specific API Versions
```
Protected URIs:
  - .*/api/v1/.*
  - .*/api/v2/.*
Public URIs:
  - .*/api/v1/public/.*
  - .*/api/v2/public/.*
```

### Pattern 3: Protect Admin and User Areas
```
Protected URIs:
  - .*/admin/.*
  - .*/user/.*
Public URIs:
  - .*/login$
  - .*/signup$
  - .*/forgot-password$
```

## Troubleshooting

### Issue: Public endpoint still requires authentication

**Solution**: Check your regex patterns. Remember to:
- Escape special characters (`.` → `\.`)
- Use `$` to match the end of the URL
- Test patterns with a regex tester

### Issue: Protected endpoint is publicly accessible

**Solution**: 
1. Verify Protected URIs are configured correctly
2. Check that Public URIs don't have conflicting patterns
3. Remember: Public URIs take precedence over Protected URIs

### Issue: oauth2-proxy not respecting configuration

**Solution**:
1. Ensure `skip_auth_regex` in oauth2-proxy matches your Public URIs
2. Restart oauth2-proxy after configuration changes
3. Check oauth2-proxy logs for errors

## Best Practices

1. **Use Specific Patterns**: Be as specific as possible to avoid accidentally exposing or over-protecting endpoints
2. **Test Thoroughly**: Test each endpoint after configuration changes
3. **Document Your Patterns**: Keep a list of what each pattern is meant to protect/expose
4. **Use Anchors**: Use `^` and `$` to avoid partial matches
5. **Monitor Access**: Review access logs to ensure patterns are working as expected
