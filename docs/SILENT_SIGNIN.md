# Silent Sign-In Configuration

## Overview

The `silentSignin` feature allows users who are already authenticated to automatically proceed with login when accessing a Casdoor login page, typically in an iframe context. This is useful for seamless integration with applications that need to obtain tokens without showing the login UI.

## How It Works

1. User is already signed in to Casdoor (has a valid session cookie)
2. Application opens Casdoor login page in an iframe with `?silentSignin=1` parameter
3. Casdoor detects the existing session and automatically completes the login flow
4. The result is communicated back to the parent window via `postMessage`

## Configuration

### Enabling SameSite=None for Session Cookies

Modern browsers implement third-party cookie restrictions that block cookies in cross-origin iframe contexts by default. To enable `silentSignin` to work properly, you need to configure session cookies to use `SameSite=None`.

**In `conf/app.conf`, set:**

```
enableSessionCookieSameSiteNone = true
```

### Requirements

⚠️ **Important:** Setting `SameSite=None` requires HTTPS. Browsers will reject cookies with `SameSite=None` over HTTP connections (except for localhost in some browsers).

#### For Production Deployment

1. Ensure your Casdoor instance is served over HTTPS
2. Set `enableSessionCookieSameSiteNone = true` in `conf/app.conf`
3. Restart Casdoor

#### For Local Development

**Option 1: Use HTTPS locally (Recommended for testing iframe scenarios)**

1. Generate self-signed certificates for localhost
2. Configure your application to use HTTPS
3. Set `enableSessionCookieSameSiteNone = true`

**Option 2: Same-origin setup (Works without HTTPS)**

If your application and Casdoor are served from the same origin (same protocol, domain, and port), you don't need `SameSite=None`:

1. Keep `enableSessionCookieSameSiteNone = false` (default)
2. Deploy both applications on the same domain (e.g., using a reverse proxy)

**Option 3: Use localhost (Limited testing)**

Some browsers (like Chrome) allow `SameSite=None` cookies on localhost without Secure flag for development purposes. However, this is not consistent across all browsers.

## Troubleshooting

### Issue: "Please login first" error when using `?silentSignin=1`

**Possible causes:**

1. **No existing session:** The user is not signed in. `silentSignin` requires an existing valid session.
   - **Solution:** Ensure the user is logged in before accessing the silentSignin URL

2. **Session cookie not sent:** The browser is blocking third-party cookies
   - **Solution:** Enable `enableSessionCookieSameSiteNone = true` and ensure HTTPS is used

3. **HTTPS not configured:** `SameSite=None` requires HTTPS
   - **Solution:** Configure HTTPS or use same-origin deployment

4. **Session expired:** The session has expired
   - **Solution:** User needs to login again

### Issue: Login form still shows even with `?silentSignin=1`

This is expected behavior when no valid session exists. The `silentSignin` parameter is a hint to automatically proceed if a session exists, but it doesn't create a new session. If no session is found, the normal login form is displayed.

## Client-Side Integration

When using `silentSignin` in an iframe, listen for postMessage events:

```javascript
window.addEventListener("message", (event) => {
  if (event.data.tag === "Casdoor" && event.data.type === "SilentSignin") {
    switch (event.data.data) {
      case "signing-in":
        console.log("Silent sign-in successful, completing login...");
        break;
      case "user-not-logged-in":
        console.log("No session found, user needs to login");
        // Redirect to normal login or show login prompt
        break;
    }
  }
});
```

## Security Considerations

- Only enable `SameSite=None` if you need cross-origin iframe functionality
- Always use HTTPS in production when `SameSite=None` is enabled
- Implement proper CORS policies to restrict which origins can embed your Casdoor instance
- Monitor for potential CSRF attacks and ensure proper origin validation

## References

- [MDN: SameSite cookies](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite)
- [Chrome SameSite Updates](https://www.chromium.org/updates/same-site/)
- [Casdoor Documentation](https://casdoor.org)
