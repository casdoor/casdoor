# Kubernetes Ingress-Nginx Authentication with Casdoor

This guide explains how to configure authentication for applications on a Kubernetes cluster using ingress-nginx with Casdoor as your Identity and Access Management (IAM) provider.

## Table of Contents

- [Authentication Approaches](#authentication-approaches)
- [Approach 1: Using oauth2-proxy (Recommended for Most Cases)](#approach-1-using-oauth2-proxy-recommended-for-most-cases)
- [Approach 2: Direct Integration with Ingress Annotations](#approach-2-direct-integration-with-ingress-annotations)
- [Approach 3: Native Application Integration](#approach-3-native-application-integration)
- [Security Comparison](#security-comparison)
- [Choosing the Right Approach](#choosing-the-right-approach)

## Authentication Approaches

There are three main approaches to integrate Casdoor authentication with applications on Kubernetes using ingress-nginx:

1. **Using oauth2-proxy** - A reverse proxy that handles OAuth2/OIDC authentication
2. **Direct integration with ingress annotations** - Using nginx ingress auth annotations
3. **Native application integration** - Applications integrate directly with Casdoor

## Approach 1: Using oauth2-proxy (Recommended for Most Cases)

### Overview

oauth2-proxy acts as a middleware between ingress-nginx and your application, handling all authentication logic. This is the **safest and most flexible approach** for most use cases.

### Advantages

- ✅ **No application code changes required** - Works with any application
- ✅ **Centralized authentication** - Single point of authentication management
- ✅ **Session management** - Handles session cookies and token refresh
- ✅ **Multiple authentication providers** - Easy to switch or add providers
- ✅ **Battle-tested** - Widely used in production environments
- ✅ **Security headers** - Automatically adds security headers

### Disadvantages

- ⚠️ Additional component to maintain
- ⚠️ Slight increase in latency (typically negligible)
- ⚠️ Additional resource usage (minimal)

### Implementation

#### Step 1: Create Casdoor Application

1. Log into Casdoor admin panel
2. Create a new application with the following settings:
   - **Name**: `my-app-oauth2-proxy`
   - **Redirect URLs**: `https://my-app.example.com/oauth2/callback`
   - **Grant types**: Select `authorization_code`
   - **Note down** the Client ID and Client Secret

#### Step 2: Deploy oauth2-proxy

Create `oauth2-proxy-deployment.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: oauth2-proxy-config
  namespace: default
data:
  oauth2_proxy.cfg: |
    # OAuth2 Provider settings
    provider = "oidc"
    provider_display_name = "Casdoor"
    redirect_url = "https://my-app.example.com/oauth2/callback"
    oidc_issuer_url = "https://casdoor.example.com"
    
    # Client credentials (use secrets in production)
    client_id = "your-client-id"
    # client_secret set via environment variable
    
    # Email domains to allow (optional)
    email_domains = [ "*" ]
    
    # Cookie settings
    cookie_name = "_oauth2_proxy"
    cookie_secret = "random-32-char-string-here"
    cookie_secure = true
    cookie_httponly = true
    
    # Upstream and listening
    http_address = "0.0.0.0:4180"
    upstreams = [ "http://my-app-service:8080" ]
    
    # Security settings
    pass_access_token = true
    pass_authorization_header = true
    set_authorization_header = true
    set_xauthrequest = true
    
    # Session settings
    session_store_type = "cookie"
    cookie_expire = "168h"  # 7 days
    
    # Skip authentication for specific paths (optional)
    skip_auth_regex = [
      "^/health",
      "^/metrics"
    ]
---
apiVersion: v1
kind: Secret
metadata:
  name: oauth2-proxy-secret
  namespace: default
type: Opaque
stringData:
  client-secret: "your-client-secret-from-casdoor"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: oauth2-proxy
  namespace: default
  labels:
    app: oauth2-proxy
spec:
  replicas: 2
  selector:
    matchLabels:
      app: oauth2-proxy
  template:
    metadata:
      labels:
        app: oauth2-proxy
    spec:
      containers:
      - name: oauth2-proxy
        image: quay.io/oauth2-proxy/oauth2-proxy:v7.5.1
        args:
        - --config=/etc/oauth2-proxy/oauth2_proxy.cfg
        env:
        - name: OAUTH2_PROXY_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: oauth2-proxy-secret
              key: client-secret
        ports:
        - containerPort: 4180
          protocol: TCP
          name: http
        volumeMounts:
        - name: config
          mountPath: /etc/oauth2-proxy
        livenessProbe:
          httpGet:
            path: /ping
            port: http
            scheme: HTTP
          initialDelaySeconds: 10
          timeoutSeconds: 1
        readinessProbe:
          httpGet:
            path: /ping
            port: http
            scheme: HTTP
          initialDelaySeconds: 10
          timeoutSeconds: 1
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
      volumes:
      - name: config
        configMap:
          name: oauth2-proxy-config
---
apiVersion: v1
kind: Service
metadata:
  name: oauth2-proxy
  namespace: default
  labels:
    app: oauth2-proxy
spec:
  type: ClusterIP
  ports:
  - port: 4180
    targetPort: http
    protocol: TCP
    name: http
  selector:
    app: oauth2-proxy
```

#### Step 3: Configure Ingress

Create `ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app-ingress
  namespace: default
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod  # Optional: for TLS
    # Optional: increase body size if needed
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
spec:
  tls:
  - hosts:
    - my-app.example.com
    secretName: my-app-tls
  rules:
  - host: my-app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: oauth2-proxy
            port:
              number: 4180
```

Apply the configurations:

```bash
kubectl apply -f oauth2-proxy-deployment.yaml
kubectl apply -f ingress.yaml
```

#### Step 4: Production Considerations

For production deployments, use Kubernetes Secrets instead of ConfigMaps for sensitive data:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: oauth2-proxy-secrets
  namespace: default
type: Opaque
stringData:
  client-id: "your-client-id"
  client-secret: "your-client-secret"
  cookie-secret: "your-random-32-char-string"  # Generate with: openssl rand -base64 32 | head -c 32 or openssl rand -hex 16
```

Update the deployment to use environment variables:

```yaml
env:
- name: OAUTH2_PROXY_CLIENT_ID
  valueFrom:
    secretKeyRef:
      name: oauth2-proxy-secrets
      key: client-id
- name: OAUTH2_PROXY_CLIENT_SECRET
  valueFrom:
    secretKeyRef:
      name: oauth2-proxy-secrets
      key: client-secret
- name: OAUTH2_PROXY_COOKIE_SECRET
  valueFrom:
    secretKeyRef:
      name: oauth2-proxy-secrets
      key: cookie-secret
```

## Approach 2: Direct Integration with Ingress Annotations

### Overview

This approach uses nginx ingress controller's built-in authentication annotations to authenticate requests directly at the ingress level.

### Advantages

- ✅ No additional deployment (oauth2-proxy)
- ✅ Lightweight solution
- ✅ Built into ingress-nginx

### Disadvantages

- ⚠️ **Less flexible** than oauth2-proxy
- ⚠️ Limited session management capabilities
- ⚠️ Requires Casdoor to validate every request (no session caching)
- ⚠️ More complex token validation setup

### Implementation

This approach requires Casdoor to expose a validation endpoint and is less commonly used. For most use cases, **oauth2-proxy is recommended** instead.

Example ingress with basic auth annotations:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app-ingress
  namespace: default
  annotations:
    kubernetes.io/ingress.class: nginx
    # External authentication via Casdoor
    nginx.ingress.kubernetes.io/auth-url: "https://casdoor.example.com/api/userinfo"
    nginx.ingress.kubernetes.io/auth-signin: "https://casdoor.example.com/login/oauth/authorize"
    nginx.ingress.kubernetes.io/auth-response-headers: "X-Auth-Request-User,X-Auth-Request-Email"
spec:
  rules:
  - host: my-app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-app-service
            port:
              number: 8080
```

**Note**: This approach requires additional configuration on the Casdoor side and is more complex. oauth2-proxy is recommended for better reliability and maintainability.

## Approach 3: Native Application Integration

### Overview

Your application integrates directly with Casdoor using OAuth2/OIDC/SAML SDKs.

### Advantages

- ✅ **Maximum control** over authentication flow
- ✅ Access to full user information within the application
- ✅ Can implement custom authorization logic
- ✅ No additional components needed

### Disadvantages

- ⚠️ **Requires application code changes**
- ⚠️ Each application must implement authentication
- ⚠️ More complex to maintain across multiple applications
- ⚠️ Application must handle session management

### Implementation

#### Step 1: Create Casdoor Application

Same as Approach 1, but set the redirect URL to your application's callback endpoint:
- **Redirect URLs**: `https://my-app.example.com/auth/callback`

#### Step 2: Integrate with Your Application

Use Casdoor SDKs for your application language:

- **Go**: [casdoor-go-sdk](https://github.com/casdoor/casdoor-go-sdk)
- **Java**: [casdoor-java-sdk](https://github.com/casdoor/casdoor-java-sdk)
- **Node.js**: [casdoor-nodejs-sdk](https://github.com/casdoor/casdoor-nodejs-sdk)
- **Python**: [casdoor-python-sdk](https://github.com/casdoor/casdoor-python-sdk)
- **PHP**: [casdoor-php-sdk](https://github.com/casdoor/casdoor-php-sdk)
- **.NET**: [casdoor-dotnet-sdk](https://github.com/casdoor/casdoor-dotnet-sdk)

Example with Node.js:

```javascript
const CasdoorSDK = require('casdoor-nodejs-sdk');

const casdoorConfig = {
  endpoint: 'https://casdoor.example.com',
  clientId: 'your-client-id',
  clientSecret: 'your-client-secret',
  certificate: '', // Your certificate content
  orgName: 'your-org',
  appName: 'your-app'
};

const sdk = new CasdoorSDK(casdoorConfig);

// Redirect to login
app.get('/login', (req, res) => {
  const redirectUrl = sdk.getSigninUrl('https://my-app.example.com/auth/callback');
  res.redirect(redirectUrl);
});

// Handle callback
app.get('/auth/callback', async (req, res) => {
  const { code, state } = req.query;
  const token = await sdk.getOAuthToken(code, state);
  const user = await sdk.parseJwtToken(token);
  
  // Store user session
  req.session.user = user;
  res.redirect('/dashboard');
});
```

#### Step 3: Configure Ingress (Simple)

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app-ingress
  namespace: default
  annotations:
    kubernetes.io/ingress.class: nginx
spec:
  rules:
  - host: my-app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-app-service
            port:
              number: 8080
```

## Security Comparison

| Feature | oauth2-proxy | Ingress Annotations | Native Integration |
|---------|--------------|---------------------|-------------------|
| **Application Changes Required** | ❌ None | ❌ None | ✅ Yes |
| **Session Management** | ✅ Built-in | ⚠️ Limited | ✅ Custom |
| **Token Refresh** | ✅ Automatic | ❌ No | ✅ Manual |
| **CSRF Protection** | ✅ Yes | ⚠️ Depends | ✅ Manual |
| **Security Headers** | ✅ Automatic | ⚠️ Manual | ✅ Manual |
| **Centralized Auth** | ✅ Yes | ✅ Yes | ❌ Per-app |
| **Complexity** | 🟢 Low | 🟡 Medium | 🔴 High |
| **Production Ready** | ✅ Yes | ⚠️ Depends | ✅ Yes (if done right) |
| **Maintenance** | 🟢 Easy | 🟡 Medium | 🔴 Complex |
| **Performance Impact** | 🟢 Minimal | 🟢 Minimal | 🟢 None |
| **Fine-grained Control** | 🟡 Medium | 🟡 Medium | 🟢 Full |

### Security Best Practices

Regardless of the approach chosen:

1. **Always use HTTPS/TLS** in production
   - Use cert-manager for automatic certificate management
   - Configure TLS in your ingress resources

2. **Secure cookie settings**
   - Set `Secure`, `HttpOnly`, and `SameSite` flags
   - Use encrypted session storage

3. **Implement CSRF protection**
   - oauth2-proxy includes this by default
   - Native apps should implement CSRF tokens

4. **Regular security updates**
   - Keep Casdoor, oauth2-proxy, and SDKs updated
   - Monitor security advisories

5. **Use Kubernetes Secrets**
   - Never hardcode credentials
   - Use sealed-secrets or external secret managers

6. **Network policies**
   - Restrict traffic between pods
   - Only allow necessary communications

7. **Rate limiting**
   - Configure rate limiting on ingress
   - Protect against brute force attacks

## Choosing the Right Approach

### Use **oauth2-proxy** (Approach 1) when:

- ✅ You want authentication for legacy applications
- ✅ You need a quick, battle-tested solution
- ✅ You want centralized authentication management
- ✅ You don't want to modify application code
- ✅ You need to protect multiple applications consistently

**This is the recommended approach for most use cases.**

### Use **Ingress Annotations** (Approach 2) when:

- ⚠️ You want to minimize additional components
- ⚠️ You have simple authentication requirements
- ⚠️ You're comfortable with limited session management

**Only recommended for simple use cases; oauth2-proxy is usually better.**

### Use **Native Integration** (Approach 3) when:

- ✅ You need deep integration with application logic
- ✅ You need custom authorization rules
- ✅ You need access to full user profile in the application
- ✅ You're building a new application
- ✅ You need fine-grained control over authentication flow

## Complete Example: oauth2-proxy with Casdoor

Here's a complete working example:

### 1. Deploy Casdoor on Kubernetes

```yaml
# casdoor-deployment.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: casdoor-config
  namespace: casdoor
data:
  app.conf: |
    appname = casdoor
    httpport = 8000
    runmode = prod
    copyrequestbody = true
    sessionOn = true
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: casdoor
  namespace: casdoor
spec:
  replicas: 2
  selector:
    matchLabels:
      app: casdoor
  template:
    metadata:
      labels:
        app: casdoor
    spec:
      containers:
      - name: casdoor
        image: casbin/casdoor:v1.577.0  # Pin to specific version in production
        ports:
        - containerPort: 8000
        env:
        - name: RUNNING_IN_DOCKER
          value: "true"
        volumeMounts:
        - name: config
          mountPath: /conf
        resources:
          requests:
            cpu: 200m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi
      volumes:
      - name: config
        configMap:
          name: casdoor-config
---
apiVersion: v1
kind: Service
metadata:
  name: casdoor-service
  namespace: casdoor
spec:
  selector:
    app: casdoor
  ports:
  - port: 8000
    targetPort: 8000
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: casdoor-ingress
  namespace: casdoor
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - casdoor.example.com
    secretName: casdoor-tls
  rules:
  - host: casdoor.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: casdoor-service
            port:
              number: 8000
```

### 2. Deploy Your Application with oauth2-proxy

Use the oauth2-proxy configuration from Approach 1 above.

### 3. Access Your Application

1. Visit `https://my-app.example.com`
2. You'll be redirected to Casdoor login
3. After successful authentication, you'll be redirected back to your app
4. oauth2-proxy will set authentication headers for your application

## Troubleshooting

### Common Issues

1. **Redirect loop**
   - Check that `redirect_url` in oauth2-proxy matches the callback URL in Casdoor
   - Ensure `cookie_secure` is set correctly for your setup

2. **401 Unauthorized**
   - Verify client ID and secret are correct
   - Check that the application is enabled in Casdoor
   - Verify OIDC issuer URL is accessible

3. **Cookie issues**
   - Ensure cookie domain is set correctly
   - Check cookie size limits (use Redis for large sessions)
   - Verify `cookie_secret` is set (32 characters)

4. **CORS errors**
   - Configure CORS properly in Casdoor
   - Check ingress CORS annotations if needed

### Debugging

Enable debug logging in oauth2-proxy:

```yaml
args:
- --config=/etc/oauth2-proxy/oauth2_proxy.cfg
- --logging-filename=
- --standard-logging
- --request-logging
- --auth-logging
```

Check logs:

```bash
kubectl logs -n default deployment/oauth2-proxy
kubectl logs -n casdoor deployment/casdoor
```

## Additional Resources

- [Casdoor Documentation](https://casdoor.org)
- [oauth2-proxy Documentation](https://oauth2-proxy.github.io/oauth2-proxy/)
- [ingress-nginx Documentation](https://kubernetes.github.io/ingress-nginx/)
- [Casdoor SDKs](https://casdoor.org/docs/how-to-connect/overview)
- [Kubernetes Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)

## Conclusion

For most use cases, **using oauth2-proxy (Approach 1) is the recommended and safest approach**. It provides:

- Enterprise-grade security
- No application changes
- Easy maintenance
- Battle-tested in production
- Great community support

Only use native integration (Approach 3) if you have specific requirements that oauth2-proxy cannot fulfill, such as deep application integration or custom authorization logic.

The ingress annotations approach (Approach 2) is generally not recommended unless you have very specific constraints that prevent using oauth2-proxy.
