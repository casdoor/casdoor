# OAuth State Parameter - CSRF Protection Guide

## Overview

The OAuth 2.0 `state` parameter is a critical security mechanism used to prevent Cross-Site Request Forgery (CSRF) attacks during the OAuth authorization code flow. Casdoor now fully supports state validation to protect your applications.

## How It Works

### The OAuth Flow with State

1. **Authorization Request**: When your application initiates an OAuth flow, it should:
   - Generate a unique, random `state` value (e.g., a UUID or random string)
   - Store this state value in the user's session or a secure, short-lived storage
   - Include the state in the authorization URL

2. **Authorization Response**: After the user authorizes, Casdoor redirects back with:
   - The authorization `code`
   - The same `state` value that was sent

3. **Token Exchange**: When exchanging the code for a token:
   - Your backend extracts the `state` from the callback
   - Validates it matches the stored state from step 1
   - Sends both `code` and `state` to Casdoor's token endpoint

4. **State Validation**: Casdoor validates that:
   - The `state` sent with the token request matches the `state` stored with the authorization code
   - If they don't match, the token request is rejected with an `invalid_grant` error

## Implementation Examples

### Frontend (JavaScript/TypeScript)

```javascript
import SDK from "casdoor-js-sdk";
import { v4 as uuidv4 } from 'uuid';

const sdk = new SDK({
    serverUrl: "https://your-casdoor-server",
    clientId: "your-client-id",
    appName: "your-app",
    organizationName: "your-org",
    redirectPath: "/auth/callback",
    signinPath: "/api/auth/signin",
});

// Step 1: Generate and store state before redirecting to Casdoor
async function initiateLogin() {
    // Generate a unique state value
    const state = uuidv4();
    
    // Store state in sessionStorage or a secure cookie
    // Note: In production, use an HTTP-only cookie set by your backend
    sessionStorage.setItem('oauth_state', state);
    
    // Redirect to Casdoor with the state parameter
    const signinUrl = sdk.getSigninUrl();
    window.location.href = `${signinUrl}&state=${encodeURIComponent(state)}`;
}

// Step 2: Handle the callback
async function handleCallback() {
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');
    const returnedState = urlParams.get('state');
    
    // Retrieve stored state
    const storedState = sessionStorage.getItem('oauth_state');
    
    // Validate state
    if (!returnedState || returnedState !== storedState) {
        throw new Error('State validation failed - possible CSRF attack');
    }
    
    // Clear the used state
    sessionStorage.removeItem('oauth_state');
    
    // Send code and state to your backend
    const response = await fetch('/api/auth/signin', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code, state: returnedState }),
        credentials: 'include', // Include cookies
    });
    
    if (!response.ok) {
        throw new Error('Authentication failed');
    }
    
    const data = await response.json();
    // Handle successful authentication
}
```

### Backend (Python with Flask)

```python
from flask import Flask, request, session, jsonify
import requests
import secrets

app = Flask(__name__)
app.secret_key = 'your-secret-key-here'  # Use a strong secret in production

CASDOOR_SERVER = 'https://your-casdoor-server'
CLIENT_ID = 'your-client-id'
CLIENT_SECRET = 'your-client-secret'
REDIRECT_URI = 'https://your-app.com/auth/callback'

@app.route('/api/auth/initiate')
def initiate_auth():
    """Step 1: Generate state and redirect to Casdoor"""
    # Generate a cryptographically secure random state
    state = secrets.token_urlsafe(32)
    
    # Store state in session (server-side)
    session['oauth_state'] = state
    
    # Build authorization URL
    auth_url = (
        f"{CASDOOR_SERVER}/login/oauth/authorize"
        f"?client_id={CLIENT_ID}"
        f"&response_type=code"
        f"&redirect_uri={REDIRECT_URI}"
        f"&scope=read"
        f"&state={state}"
    )
    
    return jsonify({'auth_url': auth_url})

@app.route('/api/auth/signin', methods=['POST'])
def signin():
    """Handle signin from frontend"""
    data = request.get_json()
    code = data.get('code')
    returned_state = data.get('state')
    
    # Retrieve stored state from session
    stored_state = session.get('oauth_state')
    
    # Validate state
    if not returned_state or not stored_state or returned_state != stored_state:
        return jsonify({'error': 'State validation failed'}), 400
    
    # Clear the used state
    session.pop('oauth_state', None)
    
    # Exchange code for token
    token_url = f"{CASDOOR_SERVER}/api/login/oauth/access_token"
    payload = {
        'grant_type': 'authorization_code',
        'client_id': CLIENT_ID,
        'client_secret': CLIENT_SECRET,
        'code': code,
        'state': returned_state,  # Include state in token request
    }
    
    response = requests.post(token_url, data=payload)
    
    if response.status_code != 200:
        return jsonify({'error': 'Token exchange failed'}), 400
    
    token_data = response.json()
    session['access_token'] = token_data.get('access_token')
    
    return jsonify({'success': True, 'access_token': token_data.get('access_token')})
```

### Backend (Node.js with Express)

```javascript
const express = require('express');
const session = require('express-session');
const crypto = require('crypto');
const axios = require('axios');

const app = express();
app.use(express.json());
app.use(session({
    secret: 'your-secret-key-here',
    resave: false,
    saveUninitialized: false,
    cookie: { 
        secure: true, // Use HTTPS in production
        httpOnly: true,
        sameSite: 'lax'
    }
}));

const CASDOOR_SERVER = 'https://your-casdoor-server';
const CLIENT_ID = 'your-client-id';
const CLIENT_SECRET = 'your-client-secret';

// Step 1: Generate state and redirect
app.get('/api/auth/initiate', (req, res) => {
    const state = crypto.randomBytes(32).toString('hex');
    req.session.oauthState = state;
    
    const authUrl = new URL(`${CASDOOR_SERVER}/login/oauth/authorize`);
    authUrl.searchParams.set('client_id', CLIENT_ID);
    authUrl.searchParams.set('response_type', 'code');
    authUrl.searchParams.set('redirect_uri', 'https://your-app.com/callback');
    authUrl.searchParams.set('scope', 'read');
    authUrl.searchParams.set('state', state);
    
    res.json({ authUrl: authUrl.toString() });
});

// Step 2: Handle signin
app.post('/api/auth/signin', async (req, res) => {
    const { code, state: returnedState } = req.body;
    const storedState = req.session.oauthState;
    
    // Validate state
    if (!returnedState || !storedState || returnedState !== storedState) {
        return res.status(400).json({ error: 'State validation failed' });
    }
    
    delete req.session.oauthState;
    
    // Exchange code for token
    try {
        const response = await axios.post(
            `${CASDOOR_SERVER}/api/login/oauth/access_token`,
            null,
            {
                params: {
                    grant_type: 'authorization_code',
                    client_id: CLIENT_ID,
                    client_secret: CLIENT_SECRET,
                    code: code,
                    state: returnedState,
                }
            }
        );
        
        req.session.accessToken = response.data.access_token;
        res.json({ success: true, access_token: response.data.access_token });
    } catch (error) {
        res.status(400).json({ error: 'Token exchange failed' });
    }
});
```

## Security Best Practices

1. **Always Generate Random State**: Use a cryptographically secure random generator
2. **Store State Securely**: Use server-side session storage with HTTP-only cookies
3. **One-Time Use**: Clear state after validation
4. **Validate on Backend**: Never trust frontend-only validation
5. **Use HTTPS**: Always use HTTPS in production

## Backward Compatibility

- Tokens created **without** state: validation is **skipped**
- Tokens created **with** state: validation is **required**

## Error Handling

If state validation fails:

```json
{
    "error": "invalid_grant",
    "error_description": "state parameter validation failed"
}
```

## Additional Resources

- [OAuth 2.0 RFC 6749 - State Parameter](https://datatracker.ietf.org/doc/html/rfc6749#section-4.1.1)
- [OAuth 2.0 Security Best Current Practice](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics)
