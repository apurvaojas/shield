# Shield App Data Model

```mermaid
graph TB
    Client --> AuthN[AuthN Service]
    AuthN[AuthN Service] --> IdentityProvider[AWS Cognito]
    InternalAdmin --> AuthZ[AuthZ Service]
    AuthZ --> OPA[OPA Agents] // sits as a sidecar of applications
    AuthN --> DB[(User DB)] // signup
    AuthZ --> PolicyDB[(Policy DB)]
    App[Application] --> OPA[OPA Agents] // sits as a sidecar of applications
```

**AuthN**

Identity provider can be AWS Cognito, Azure Identity, Okta, Auth0, Keycloak, etc.

Two types of users: Individual and Organization.

- **Individual User**: Can sign up and log in using AuthN as a proxy to the identity provider.
- **Organizational User**: Will sign up and onboard Organization SSO. If SSO is unavailable, they will leverage the identity provider.

**AuthZ**

Internal admins can log in to an internal portal where applications can be registered with supported roles in the application and OPA policies specific to that application. OPA policies should be versioned and unit testable.

Once published, OPA policy bundles should be synced to OPA Agents of the application.