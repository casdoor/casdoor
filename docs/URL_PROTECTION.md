# URL Protection Configuration

This document explains how to configure URL-level protection for applications in Casdoor.

## Overview

The URL protection feature allows you to specify which URLs within your application should require authentication and which should be publicly accessible. This is particularly useful when using Casdoor with reverse proxies like oauth2-proxy.

## Configuration Fields

### Protected URIs

A list of URL patterns that require authentication. These URLs will be protected by Casdoor authentication.

### Public URIs

A list of URL patterns that are publicly accessible without authentication. This takes precedence over Protected URIs.

## Behavior

1. **No Configuration**: If both Protected URIs and Public URIs are empty, all URIs are protected by default (backward compatibility).

2. **Protected URIs Only**: If only Protected URIs is configured:
   - URIs matching the protected patterns require authentication
   - URIs not matching any protected pattern are NOT protected

3. **Public URIs Only**: If only Public URIs is configured:
   - URIs matching the public patterns are NOT protected
   - URIs not matching any public pattern require authentication

4. **Both Configured**: If both are configured:
   - Public URIs takes precedence
   - URIs matching public patterns are NOT protected
   - URIs matching protected patterns (and not public) require authentication
   - URIs not matching any pattern are NOT protected

## Pattern Matching

Patterns support both exact string matching and regular expressions:

- **Exact Match**: `https://app.example.com/api` will match exactly that URL
- **Regex Pattern**: `https://app\.example\.com/api.*` will match `https://app.example.com/api/users`, `https://app.example.com/api/posts`, etc.
- **Ending Anchor**: Use `$` to match exact paths: `https://app\.example\.com/api$` will match `https://app.example.com/api` but NOT `https://app.example.com/api-another`

### Pattern Tips

1. Escape special regex characters (`.`, `?`, `*`, etc.) if you want literal matching
2. Use `.*` for wildcard matching
3. Use `^` and `$` for start and end anchors
4. Test your patterns carefully

## Examples

### Example 1: Protect specific API endpoints

**Scenario**: You have an application with:
- `https://app.example.com/api` - should be protected
- `https://app.example.com/api-public` - should NOT be protected

**Configuration**:
- Protected URIs: 
  - `https://app\.example\.com/api$`
- Public URIs:
  - `https://app\.example\.com/api-public`

### Example 2: Protect everything except health check

**Scenario**: You want to protect all endpoints except `/health`

**Configuration**:
- Protected URIs: (leave empty)
- Public URIs:
  - `.*/health`

### Example 3: Protect specific paths with wildcards

**Scenario**: Protect all admin endpoints

**Configuration**:
- Protected URIs:
  - `.*/admin/.*`
- Public URIs: (leave empty)

### Example 4: Complex scenario

**Scenario**: 
- Protect `/api/*` endpoints
- But allow public access to `/api/public/*`
- Also allow public access to `/health` and `/status`

**Configuration**:
- Protected URIs:
  - `.*/api/.*`
- Public URIs:
  - `.*/api/public/.*`
  - `.*/health$`
  - `.*/status$`

## Integration with oauth2-proxy

When using Casdoor with oauth2-proxy, you can use the `IsUriProtected` method in your application code to determine whether a specific URI should be protected. This method is available on the Application object.

Example in Go:
```go
app, err := object.GetApplication("admin/my-app")
if err != nil {
    // handle error
}

uri := "https://app.example.com/api"
if app.IsUriProtected(uri) {
    // Require authentication
} else {
    // Allow public access
}
```

## Notes

- Changes to URL protection configuration require updating the application settings
- URL patterns are evaluated in the order: Public URIs first, then Protected URIs
- Invalid regex patterns will fall back to exact string matching
- Empty or blank patterns are ignored

## Performance Considerations

The `IsUriProtected` method compiles regex patterns on each call. This design choice was made because:

1. **Dynamic Configuration**: Application settings can be updated at runtime without server restart
2. **Typical Usage**: This method is called during the authentication flow (e.g., when a user tries to access a protected resource), not on every request
3. **Small Pattern Sets**: Most applications have a small number of URL protection patterns (typically < 10)

For high-traffic scenarios with many patterns, consider:
- Keeping pattern lists small and specific
- Using exact string matching where possible
- Implementing pattern caching at the application integration layer if needed

In practice, the performance impact is negligible for typical use cases.
