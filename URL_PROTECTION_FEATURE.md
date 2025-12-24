# URL Protection Feature

This feature allows you to configure URL-level access control for your applications.

## Quick Start

1. Open your application in the Casdoor admin panel
2. Scroll to the "Protected URIs" and "Public URIs" sections
3. Add URL patterns to define which URLs should be protected

## Use Cases

### Example: oauth2-proxy Integration

If you're using oauth2-proxy with your application and want to:
- Protect: `https://app.example.com/api`
- Allow public access: `https://app.example.com/api-another`

**Configuration:**
```
Protected URIs:
  - https://app\.example\.com/api$

Public URIs:
  - https://app\.example\.com/api-another
```

### Example: Public Health Check

To allow public access to `/health` while protecting everything else:

**Configuration:**
```
Protected URIs: (leave empty)

Public URIs:
  - .*/health$
```

## Pattern Syntax

- Use regular expressions for flexible matching
- Escape special characters: `.` → `\.`
- Use `$` for exact end matching
- Use `.*` for wildcard matching

## Documentation

For detailed documentation, examples, and best practices, see [URL_PROTECTION.md](docs/URL_PROTECTION.md)

## API Usage

The `IsUriProtected(uri string) bool` method is available on Application objects to programmatically check if a URI requires protection.
