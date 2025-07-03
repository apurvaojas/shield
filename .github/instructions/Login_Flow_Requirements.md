# ✅ Final Technical Requirement Specification

## 🔐 Secure OAuth2 Login with AWS Cognito (Custom UI)

### 🧱 Stack Components

| Component         | Tech                 | Description                                    |
| ----------------- | -------------------- | ---------------------------------------------- |
| Identity Provider | AWS Cognito          | User authentication, MFA, token issuer         |
| Frontend          | React SPA            | Custom login UI, nonce generation              |
| Backend           | Go Gin on AWS Lambda | Auth proxy, token exchange, session management |

---

## 🎯 Goals

* **Custom login UX** (no Cognito Hosted UI)
* **Support optional MFA**
* **Prevent replay/DDoS on `/auth/login`**
* **Stateless client-generated nonce validation**
* **Secure cookie-based session management**
* **Frontend-only credential handling via backend API**

---

## 🔁 Authentication Flow Summary

```
[1] React SPA: User enters username/password
    → generates random nonce
    → signs nonce with HMAC(secret) + timestamp

[2] POST /auth/login (username, password, signed_nonce)
    → Backend verifies nonce signature, freshness
    → Calls Cognito AdminInitiateAuth
        → If MFA required → respond with session token
        → Else → issue tokens as secure cookie

[3] POST /auth/mfa-verify (code, signed_nonce, session)
    → Backend verifies nonce, session
    → Calls Cognito RespondToAuthChallenge
    → If success → issue tokens as secure cookie

[4] GET /auth/session → return user info from token
[5] GET /auth/logout → clear secure session cookie
```

---

## 🔐 Security Controls

| Feature              | Method                                                                |
| -------------------- | --------------------------------------------------------------------- |
| No hosted UI         | Custom SPA form with backend auth proxy                               |
| Stateless nonce      | Client-generated UUID + timestamp + HMAC signature                    |
| Signature validation | HMAC-SHA256 (using shared secret known only to backend)               |
| Expiry check         | Max 5 minutes between nonce creation and usage                        |
| Replay protection    | Include `nonce`, `iat`, `exp` — optionally track used nonces in cache |
| MFA support          | Cognito session token + challenge flow                                |
| Token storage        | HttpOnly, Secure, SameSite cookies                                    |
| DDoS mitigation      | Stateless nonce + per-IP rate limiting                                |

---

## 🔧 Nonce Format (JWT-like or JSON + HMAC)

### Payload:

```json
{
  "nonce": "uuidv4-random-string",
  "iat": 1715940000,
  "exp": 1715940300
}
```

### HMAC Signature:

```ts
HMAC_SHA256(base64(payload), SECRET_KEY)
```

### Final Signed Nonce Token:

```
<base64(payload)>.<hmac_signature>
```

---

## 🔐 Updated Nonce Format with Device Fingerprint

### Payload:

```json
{
  "nonce": "uuidv4-random-string",
  "iat": 1715940000,
  "exp": 1715940300,
  "device": {
    "fingerprint": "hash-of-device-characteristics",
    "userAgent": "browser-info",
    "screenResolution": "1920x1080",
    "timezone": "UTC+0",
    "webglRenderer": "hash-of-gpu-info",
    "fonts": "hash-of-installed-fonts"
  }
}
```

### Frontend Device Fingerprint Generation:

```ts
async function generateDeviceFingerprint() {
  const characteristics = {
    userAgent: navigator.userAgent,
    screenRes: `${screen.width}x${screen.height}`,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    platform: navigator.platform,
    webgl: await getWebGLInfo(),
    fonts: await getFontList(),
    canvas: generateCanvasFingerprint()
  };
  return crypto.subtle.digest('SHA-256', 
    new TextEncoder().encode(JSON.stringify(characteristics))
  ).then(hash => btoa(String.fromCharCode(...new Uint8Array(hash))));
}

function generateSignedNonce(secret: string): Promise<string> {
  return generateDeviceFingerprint().then(fingerprint => {
    const payload = {
      nonce: crypto.randomUUID(),
      iat: Math.floor(Date.now() / 1000),
      exp: Math.floor(Date.now() / 1000) + 300,
      device: {
        fingerprint,
        userAgent: navigator.userAgent,
        screenResolution: `${screen.width}x${screen.height}`,
        timezone: Intl.DateTimeFormat().resolvedOptions().timeZone
      }
    };
    const payloadStr = btoa(JSON.stringify(payload));
    const sig = signHmac(payloadStr, secret);
    return `${payloadStr}.${sig}`;
  });
}
```

---

## 📦 API Contract

### 1. `POST /auth/login`

**Input:**

```json
{
  "username": "user@example.com",
  "password": "your_password",
  "signed_nonce": "<base64(payload)>.<hmac_signature>"
}
```

**Validation:**

* Decode and verify HMAC
* Check `exp` not expired
* Check `iat` within allowable skew

**Outcome:**

* If MFA required: respond with MFA session + nonce
* Else: set secure cookie with Cognito ID token

---

### 2. `POST /auth/mfa-verify`

**Input:**

```json
{
  "username": "user@example.com",
  "code": "123456",
  "session": "<cognito_temp_session>",
  "signed_nonce": "<base64(payload)>.<hmac_signature>"
}
```

**Outcome:**

* Verify nonce as before
* Use `RespondToAuthChallenge` to complete auth
* On success: issue tokens via secure cookie

---

### 3. `GET /auth/session`

**Output:**

```json
{
  "sub": "...",
  "email": "...",
  "roles": [...]
}
```

* Reads token from cookie
* Validates token signature + expiry

---

### 4. `GET /auth/logout`

* Clears session cookies
* Optionally revokes refresh token (if used)

---

## 🍪 Session Cookie Configuration

| Property      | Value                            |
| ------------- | -------------------------------- |
| Name          | `id_token`, `access_token`, etc. |
| Secure        | ✅                                |
| HttpOnly      | ✅                                |
| SameSite      | `Strict` or `Lax`                |
| TTL           | 15 minutes access token          |
| Refresh token | optional, rotated or encrypted   |

---

## 💻 Frontend Logic (React)

### Nonce Generation Example:

```ts
function generateSignedNonce(secret: string): string {
  const payload = {
    nonce: crypto.randomUUID(),
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + 300
  };
  const payloadStr = btoa(JSON.stringify(payload));
  const sig = signHmac(payloadStr, secret); // from Web Crypto
  return `${payloadStr}.${sig}`;
}
```

Use this nonce for `/auth/login` and `/auth/mfa-verify`.

---

## ⚠️ Notes for Production

* HMAC secret must never be in frontend. Only used by backend to verify signatures.
* Sign nonce on frontend **only if secret is obfuscated via a secure enclave** (e.g., WebAssembly or backend-issued token).
  Otherwise, consider moving nonce signing to a **GET `/auth/login-nonce`** backend endpoint.
* Consider rotating the HMAC secret regularly.
* You may also layer in:

  * IP throttling
  * CAPTCHA (as fallback for flagged users)
  * Login delay or exponential backoff

---

## ✅ Summary

| Feature               | Implemented |
| --------------------- | ----------- |
| Custom login UX       | ✅           |
| MFA support           | ✅           |
| Secure cookie session | ✅           |
| Stateless nonce auth  | ✅           |
| Replay/DDoS-resistant | ✅           |
| SSR/SPA-compatible    | ✅           |

---


 **backend-managed SSO login flow** using **AWS Cognito Federation**, where:

* The **React SPA never directly interacts with Cognito** (even for SSO).
* All **token exchange, state management, and redirects** are handled securely via the **Go Gin backend** (AWS Lambda).
* This allows your backend to set secure HTTP-only session cookies, abstract away Cognito details, and maintain a clean, centralized login flow.

---

# ✅ Updated Final Requirement (with Backend-Handled SSO Flow)

## 🔐 Core Principles

| Area               | Strategy                                                     |
| ------------------ | ------------------------------------------------------------ |
| **SSO login**      | Initiated from SPA → handled completely by Go Gin backend    |
| **Token exchange** | Done server-side, backend fetches and stores tokens securely |
| **Session**        | Secure, HTTP-only, short-lived cookies                       |
| **Callback URL**   | Cognito → backend `/auth/callback` endpoint                  |
| **SPA role**       | Just redirects or polls `/auth/session` for status           |

---

## 🧭 High-Level Flow (SSO)

### Step 1: SPA → Initiate SSO login

* `GET /auth/sso/start?provider=okta&redirect=/dashboard`
* Backend generates:

  * OAuth2 `state` and `nonce`
  * Stores it in temporary session (encrypted cookie or Redis)
* Redirects to Cognito SSO Authorization URL with:

  * `identity_provider=OktaSAML`
  * `response_type=code`
  * `state=...` (opaque)
  * `nonce=...` (optional; for OIDC)
  * `redirect_uri=https://api.example.com/auth/callback`

---

### Step 2: Cognito → Redirects to Go Gin `/auth/callback`

* URL: `GET /auth/callback?code=...&state=...`

Backend:

* Verifies `state` against stored value
* Exchanges `code` for tokens using Cognito token endpoint
* Extracts user info (`sub`, `email`, etc.)
* Maps Cognito identity to app DB user
* Creates secure session
* Redirects to original `redirect=/dashboard`

---

### Step 3: SPA Reads Session

* On landing, SPA calls `GET /auth/session`
* Backend checks session cookie and returns:

  ```json
  {
    "user": { "id": "uuid", "email": "...", "org": "acme" },
    "roles": [...],
    "expiresIn": 3600
  }
  ```

---

## 🧱 Go Gin Backend: Updated Auth Endpoints

| Endpoint           | Method | Purpose                                       |
| ------------------ | ------ | --------------------------------------------- |
| `/auth/sso/start`  | GET    | Initiate SSO login redirect via Cognito       |
| `/auth/callback`   | GET    | Cognito redirect URL (handles token exchange) |
| `/auth/session`    | GET    | Return logged-in user session                 |
| `/auth/logout`     | POST   | Clear cookies                                 |
| `/auth/login`      | POST   | Email/password login with signed nonce        |
| `/auth/mfa-verify` | POST   | MFA verification (if challenge triggered)     |

---

## 🏢 Org-Specific SSO Setup

Each org can register:

* `name`: e.g. `Acme Corp`
* `SSO config`: either **OIDC** or **SAML**

  * Setup in your DB
  * **Also configured in AWS Cognito** via:

    * `identity_provider_name`: e.g., `AcmeOkta`
    * Redirects handled by Cognito Hosted UI using `identity_provider=AcmeOkta`

Mapping Logic:

* Your DB maintains:

  ```json
  {
    "org": "acme",
    "sso_provider": "AcmeOkta",
    "idp_type": "saml",
    "cognito_provider_name": "AcmeOkta",
    "callback_url": "https://api.example.com/auth/callback"
  }
  ```

---

## 🧠 Security Best Practices (Confirmed)

| Concern                        | Implementation                                            |
| ------------------------------ | --------------------------------------------------------- |
| **Nonce HMAC validation**      | Required for all login types                              |
| **Signed `state` param**       | Prevent CSRF on SSO                                       |
| **Server-side token exchange** | Avoids token exposure to browser                          |
| **Session management**         | Signed + encrypted cookies, with refresh on re-auth       |
| **Per-org SSO config**         | Maps to Cognito IDP for multi-tenant security isolation   |
| **Callback whitelist**         | Only allow pre-registered return URLs per org             |
| **Rate limiting**              | On `/auth/login`, `/auth/sso/start`, and `/auth/callback` |
| **Federated user binding**     | Link external `sub` to internal user ID on first login    |

---

## 🔒 Optional Enhancements

| Feature                          | Benefit                                               |
| -------------------------------- | ----------------------------------------------------- |
| Device fingerprinting            | Lock session to browser/device for extra security     |
| Short-lived session + refresh    | Add `/auth/refresh` to rotate Cognito tokens silently |
| Detect suspicious login patterns | Alert admin or require re-verification                |
| OIDC ID token verification       | Verify signature of ID token on backend               |
| Refresh token rotation tracking  | Detect replay and compromise                          |

----





