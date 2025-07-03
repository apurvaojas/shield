# 🏗️ Multi-Module Architecture

## 📁 Project Structure

```
├── modules/
│   ├── authn/                # Authentication Module (exposes services/APIs to root main.go)
│   │   ├── internal/
│   │   │   ├── api/        # HTTP handlers (e.g., for Gin routes)
│   │   │   ├── auth/       # Auth domain logic (services, providers)
│   │   │   │   ├── provider/   # Auth providers (Cognito, Azure, etc.)
│   │   │   │   │   ├── cognito/
│   │   │   │   │   ├── azure/
│   │   │   │   │   ├── auth0/
│   │   │   │   │   └── mock/
│   │   │   │   ├── nonce/      # Nonce validation logic
│   │   │   │   └── session/    # Session management logic
│   │   │   ├── models/     # Domain-specific models for authn
│   │   │   └── config/     # Configuration structures for authn (loaded by root main.go)
│   │   └── go.mod          # Go module file for authn dependencies
│   │
│   └── authz/              # Authorization Module (OPA Control Plane)
│       ├── cmd/            # AuthZ might still have its own entry point or be integrated similarly
│       │   └── api/
│       │       └── main.go
│       ├── internal/
│       │   ├── api/       # Policy management API
│       │   ├── sync/      # Policy distribution service
│       │   ├── store/     # Policy storage
│       │   └── apps/      # Application registry
│       └── go.mod
```

## 🔄 Request Flow

```
Request → Middleware → Handler → Service → Repository → Database
   ↑          ↓          ↓         ↓          ↓           ↓
Response ← Error Handler ← Business Logic ← Data Access ← Query
```

## 🛠️ Core Components

### Middleware Layer
```go
type AuthMiddleware struct {
    sessionService *session.Service
    nonceValidator *nonce.Validator
}

func (m *AuthMiddleware) ValidateSession() gin.HandlerFunc
func (m *AuthMiddleware) ValidateNonce() gin.HandlerFunc
func (m *AuthMiddleware) DeviceFingerprint() gin.HandlerFunc
```

### Handler Layer
```go
type AuthHandler struct {
    authService  *auth.Service
    orgService   *org.Service
    errorHandler *errors.Handler
}

func (h *AuthHandler) Login(c *gin.Context)
func (h *AuthHandler) SSOCallback(c *gin.Context)
```

### Service Layer
```go
type AuthService struct {
    provider     AuthProvider
    repository   *repository.AuthRepository
    config      *config.Config
}

func NewAuthService(factory ProviderFactory, config ProviderConfig) (*AuthService, error)

func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*TokenResponse, error)
func (s *AuthService) HandleSSOCallback(ctx context.Context, code string) (*TokenResponse, error)
```

### Repository Layer
```go
type AuthRepository struct {
    db *gorm.DB
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (*User, error)
func (r *AuthRepository) SaveSession(ctx context.Context, session *Session) error
```

## 🔐 Security Components

### Nonce Validator
```go
type NonceValidator struct {
    secret []byte
    maxAge time.Duration
}

func (v *NonceValidator) Validate(signedNonce string) (*Nonce, error)
func (v *NonceValidator) ValidateWithDevice(signedNonce string, deviceFingerprint string) error
```

### Session Manager
```go
type SessionManager struct {
    store    *redis.Client
    tokenGen *jwt.TokenGenerator
}

func (m *SessionManager) CreateSession(ctx context.Context, user *User) (*Session, error)
func (m *SessionManager) ValidateSession(ctx context.Context, token string) (*Session, error)
```

## 📊 Monitoring & Observability

- Prometheus metrics for request/response timing
- Structured logging with correlation IDs
- Distributed tracing with OpenTelemetry
- Health check endpoints
- AWS CloudWatch integration

## 🔄 Rate Limiting

```go
type RateLimiter struct {
    redis      *redis.Client
    windowSize time.Duration
    maxTokens  int
}

func (r *RateLimiter) Allow(key string) bool
func (r *RateLimiter) Reset(key string)
```

## 🏭 Dependency Injection

Using Wire for compile-time DI:

```go
func InitializeAPI() (*gin.Engine, error) {
    wire.Build(
        providerSet,     // Provides configured AuthProvider
        authServiceSet,  // Uses AuthProvider to create AuthService
        handlerSet,      // Uses AuthService
        // ... other providers
    )
    return nil, nil
}

var providerSet = wire.NewSet(
    config.New,
    wire.Bind(new(AuthProvider), new(*CognitoProvider)),
    NewProviderFactory,
)
```

## 💾 Database Schema

Key tables:

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR NOT NULL UNIQUE,
    cognito_sub VARCHAR UNIQUE,
    org_id UUID REFERENCES organizations(id),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    device_fingerprint VARCHAR NOT NULL,
    last_seen TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP
);

CREATE TABLE organizations (
    id UUID PRIMARY KEY,
    name VARCHAR NOT NULL,
    sso_provider VARCHAR,
    idp_type VARCHAR,
    callback_url VARCHAR,
    created_at TIMESTAMP
);

-- Applications and Roles
CREATE TABLE applications (
    id UUID PRIMARY KEY,
    name VARCHAR NOT NULL,
    api_key VARCHAR UNIQUE,
    opa_endpoint VARCHAR NOT NULL,  -- OPA sidecar endpoint
    status VARCHAR,
    created_at TIMESTAMP
);

CREATE TABLE application_roles (
    id UUID PRIMARY KEY,
    app_id UUID REFERENCES applications(id),
    name VARCHAR NOT NULL,
    description TEXT,
    created_at TIMESTAMP,
    UNIQUE(app_id, name)
);

-- User/Org application roles
CREATE TABLE user_app_roles (
    user_id UUID REFERENCES users(id),
    app_id UUID REFERENCES applications(id),
    role_name VARCHAR NOT NULL,
    created_at TIMESTAMP,
    PRIMARY KEY (user_id, app_id, role_name)
);

-- OPA policies
CREATE TABLE opa_policies (
    id UUID PRIMARY KEY,
    app_id UUID REFERENCES applications(id),
    name VARCHAR NOT NULL,
    rego_policy TEXT NOT NULL,
    version INTEGER,
    created_at TIMESTAMP
);

-- Policy distribution status
CREATE TABLE policy_sync_status (
    app_id UUID REFERENCES applications(id),
    version INTEGER,
    synced_at TIMESTAMP,
    PRIMARY KEY (app_id)
);
```

## 🔐 Authorization Components

### Application Registration
```go
type Application struct {
    ID           uuid.UUID
    Name         string
    APIKey       string
    OPAEndpoint  string    // OPA sidecar URL
    Roles        []string  // Available roles
    CreatedAt    time.Time
}

type PolicyService struct {
    store    PolicyStore
    sync     PolicySyncService
}

func (s *PolicyService) RegisterApplication(ctx context.Context, app *Application) error
func (s *PolicyService) UpdatePolicy(ctx context.Context, policy *OPAPolicy) error
```

### OPA Policy Management
```go
type OPAPolicy struct {
    ID        uuid.UUID
    AppID     uuid.UUID
    Name      string
    Policy    string    // Rego policy
    Version   int
}

type PolicySyncService struct {
    store     PolicyStore
    notifier  PolicyChangeNotifier
}

func (s *PolicySyncService) DistributePolicy(ctx context.Context, policy *OPAPolicy) error
func (s *PolicySyncService) GetPolicyBundle(ctx context.Context, appID uuid.UUID) (*PolicyBundle, error)
```

### Example Application OPA Integration
```go
// In application services with OPA sidecar
type OPAClient struct {
    client    *opa.Client
    bundleURL string    // URL to fetch policies from AuthZ service
}

type AccessRequest struct {
    User      UserContext     
    Action    string
    Resource  string
    Context   map[string]interface{}
}

func (c *OPAClient) CheckAccess(ctx context.Context, req *AccessRequest) (bool, error)
```

### Example OPA Policy
```rego
package app.myapp

# Define role capabilities
roles := {
    "admin": ["read", "write", "delete"],
    "editor": ["read", "write"],
    "viewer": ["read"]
}

# Evaluate access
default allow = false

allow {
    # Get user role
    role := input.user.roles[_]
    
    # Check if role allows action
    roles[role][_] == input.action

    # Custom business logic
    not is_restricted_resource
    within_working_hours
}

# Custom rules
is_restricted_resource {
    input.resource.confidential == true
    not input.user.has_clearance
}

within_working_hours {
    time.now_ns() >= input.user.shift_start
    time.now_ns() <= input.user.shift_end
}
```

## 🔄 Authorization Flow

1. Application registers with AuthZ service
2. Application defines roles during registration
3. OPA sidecar fetches policies from AuthZ service
4. On each request:
   - Application gets user context from JWT
   - Calls OPA sidecar with user context + request details
   - OPA evaluates against synced policies
   
```
Request → App → OPA Sidecar → Policy Evaluation
                    ↑
                    └── Policies synced from AuthZ service
```
