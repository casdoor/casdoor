# Direct Provider Redirect

Users can go **straight to** the chosen identity provider (e.g. Google, SAML) when the provider is specified in the URL as **`id`** and **`silentSignin=1`** is present. Casdoor’s provider selection page is **skipped** only in that case.

---

## Before

| What happened |
|---------------|
| The client application sent users to Casdoor with a URL that included the provider (e.g. `id=provider_g1_google`). |
| Casdoor **ignored** that parameter. |
| The user saw Casdoor’s login page with **all** providers listed. |
| The user had to **click the same provider again** (e.g. “Login with Google”). |

**Result:** Two steps — choose on the client app, then choose again on Casdoor.

---

## After

| What happens |
|--------------|
| The client application sends users to Casdoor with the same URL plus **`id`** (provider) and **`silentSignin=1`** (e.g. `id=provider_g1_google&silentSignin=1`). |
| Casdoor **reads** those parameters, finds the provider, and **redirects the user straight to it**. |
| The user **does not see** the provider list and **does not click again**. |

**Result:** One step — user goes from the client app directly to Google (or the SAML IdP).

---

## What works

| Provider type | Behaviour |
|---------------|-----------|
| **OAuth** (Google, GitHub, etc.) | Direct redirect to the provider’s login page. |
| **SAML** | Same flow as “click SAML” on Casdoor (API + redirect to IdP), but triggered automatically from the URL. |
| **Web3** (e.g. MetaMask) | Not supported; user still sees the login page and clicks the provider. |


---

## How we achieved it (high level)

1. **Use the URL parameters**  
   When the login page loads, we look for **`silentSignin=1`** and **`id`** (provider id). The direct redirect runs **only when both** are present; otherwise the normal login page is shown.

2. **Resolve the provider**  
   We get the application’s list of providers from the backend. We find the one whose name or id matches the parameter, and only redirect if that provider is visible for sign-in (same visibility rules as the provider buttons).

3. **Redirect instead of showing the list**  
   - **OAuth:** We build the provider’s login URL and send the user there (same as when they click “Login with Google” on Casdoor).  
   - **SAML:** We run the same process as when they click the SAML provider (call Casdoor, then send the user to the SAML IdP). We just trigger it automatically from the URL.

4. **Do this as early as possible**  
   We run this logic as soon as we have the application data (right after the API responds). If we redirect, we never draw the provider list. Redirects run in lifecycle (e.g. after API response or in componentDidUpdate), not during render.

**No backend or config changes.** All of this is in the frontend: read the parameter, find the provider, then redirect using the same flows that already exist when the user clicks a provider on the Casdoor page.

---

## Summary

| | Before | After |
|---|--------|--------|
| **URL params** | `id` ignored | `silentSignin=1` + `id` trigger direct redirect |
| **Provider list** | Always shown | Skipped when provider is in URL |
| **OAuth** | User clicks again on Casdoor | Direct redirect to provider |
| **SAML** | User clicks again on Casdoor | Automatic redirect (same flow as button) |
| **Backend / config** | — | Unchanged |

---

## Files changed

| File | What changed |
|------|----------------|
| `web/src/auth/LoginPage.js` | Added logic to read `id`, find the visible provider, and redirect (OAuth or SAML) before showing the list. |
| `web/src/auth/ProviderButton.js` | Exposed the existing SAML redirect logic so the login page can call it when the URL points to a SAML provider. |

---

## For integrators: how to use it

1. Build the authorize URL as usual (`client_id`, `redirect_uri`, `state`, etc.).
2. Add **`silentSignin=1`** to the URL so the direct-redirect behaviour is enabled.
3. Add the provider **`id`**:
   - `id=provider_g1_google`  
   - or `id=organization_4vf6hs/provider_g1_google`
4. Send the user to that URL (e.g. when they click “Login with Google” on the client app).

**Example:**

```
.../login/oauth/authorize?response_type=code&client_id=...&redirect_uri=...&state=...&id=provider_g1_google&silentSignin=1
```

If **both** `silentSignin=1` and `id` are present and the provider exists for the application, the user is sent straight to the provider and does not see Casdoor’s provider selection page. If `silentSignin=1` is omitted, the normal Casdoor login page is shown even when `id` is present.
