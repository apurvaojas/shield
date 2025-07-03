 **complete Signup Flow** for both:

* **Individual users** (email/password + optional MFA)
* **Organizations** (SAML/OIDC Federation via Cognito Identity Providers)

All handled via the **React SPA → Go Gin Backend → AWS Cognito**, maintaining **central control**, **security**, and **custom user experience**.

---

# ✅ Final Signup Flow Specification

## 📦 System Components

| Component                   | Role                                                                    |
| --------------------------- | ----------------------------------------------------------------------- |
| **React SPA**               | Handles form UI and basic validation                                    |
| **Go Gin Backend (Lambda)** | Orchestrates Cognito signup, confirmation, SSO redirect, and token flow |
| **AWS Cognito**             | Identity provider, federated IdP bridge                                 |
| **Custom DB**               | Stores organization metadata, user roles, SSO mappings                  |

---

## 🧍 Individual User Signup Flow

### ➤ Flow Overview

1. **SPA Signup Form** → POST `/auth/signup`
2. Backend registers user in Cognito (`AdminCreateUser` or `SignUp`)
3. Sends verification code (email/SMS)
4. SPA shows verification screen → submits code to `/auth/confirm`
5. Optional: user prompted for MFA setup
6. After confirmation → creates session + sets secure cookie

### ➤ API Endpoints (Go Gin)

| Endpoint           | Method | Description                              |
| ------------------ | ------ | ---------------------------------------- |
| `/auth/signup`     | POST   | Registers individual in Cognito          |
| `/auth/confirm`    | POST   | Confirms code sent via email             |
| `/auth/mfa/setup`  | POST   | (Optional) Initiate MFA setup (TOTP/SMS) |
| `/auth/mfa/verify` | POST   | Verify MFA code and complete login       |

### ➤ Backend Implementation Notes

* `SignUp()` for self-service or `AdminCreateUser()` for admin-driven flows
* Store custom attributes: `custom:user_type = "individual"`
* Use Cognito triggers (optional) for post-confirmation user setup
* Enforce strong password policy
* Hash and validate HMAC nonce to prevent abuse

---

## 🏢 Organization Signup Flow (with SSO)

### ➤ Flow Overview

1. **SPA Org Signup Form** → POST `/org/signup`
2. Backend:

   * Creates org record in custom DB
   * Creates admin user (optional: email/password or SSO user)
   * Registers identity provider in Cognito (`CreateIdentityProvider`)
   * Creates Cognito user pool domain + client (if dynamic)
3. SPA displays:

   * Org SSO Login URL (`/auth/sso/start?org=acme`)
4. User can proceed to login via SSO (handled as defined earlier)

---

### ➤ Supported SSO Types

| SSO Type | Cognito Support | Notes                            |
| -------- | --------------- | -------------------------------- |
| OIDC     | ✅ Yes           | Eg. Google Workspace, Azure AD   |
| SAML     | ✅ Yes           | Eg. Okta, PingIdentity, OneLogin |

---

### ➤ API Endpoints (Go Gin)

| Endpoint            | Method | Description                                  |
| ------------------- | ------ | -------------------------------------------- |
| `/org/signup`       | POST   | Org registration + optional admin user setup |
| `/org/sso/register` | POST   | Register SAML/OIDC IdP with Cognito          |
| `/org/settings`     | GET    | Get org SSO settings (for login entry point) |

---

### ➤ Cognito Federation Setup via Backend

Use AWS SDK in Go:

1. **CreateIdentityProvider**
2. **UpdateUserPool** → `SupportedIdentityProviders = [...]`
3. **CreateUserPoolDomain** (if needed)
4. **UpdateAppClient** to include the IdP
5. Store all mappings in your DB:

   ```json
   {
     "org": "acme",
     "idp_name": "AcmeOkta",
     "type": "saml",
     "entity_id": "...",
     "metadata_url": "...",
     "cognito_idp_name": "AcmeOkta",
     "redirect_uri": "https://api.example.com/auth/callback"
   }
   ```

---

## 🔐 Shared Security for Signup

| Feature                        | Purpose                                         |
| ------------------------------ | ----------------------------------------------- |
| HMAC-signed nonce              | Prevent replay attacks on `/auth/signup`        |
| CAPTCHA or rate limiting       | Prevent bot-driven mass signup                  |
| MFA enrollment post-confirm    | Secure high-risk accounts                       |
| Invite-based signup (optional) | For orgs to pre-register users                  |
| Email domain validation        | Enforce domain policy per organization          |
| Webhook/trigger validation     | Validate Cognito events server-side (if needed) |

---

## 📦 Optional Enhancements

* **Pre-approval workflows**: Org admins approve users after email verification
* **Org dashboard**: View, invite, deactivate users
* **Custom email branding** via Cognito Lambda triggers or SES
* **SSO fallback login**: Email/password fallback for SSO failure

---

