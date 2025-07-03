-- Development Database Initialization Script
-- This script sets up the development database with initial schema

-- Create database if it doesn't exist (this might not work in init script, but included for reference)
-- The database should be created via environment variables in docker-compose

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create tables based on the architecture document
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR NOT NULL,
    sso_provider VARCHAR,
    idp_type VARCHAR,
    callback_url VARCHAR,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR NOT NULL UNIQUE,
    cognito_sub VARCHAR UNIQUE,
    org_id UUID REFERENCES organizations(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    device_fingerprint VARCHAR NOT NULL,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Applications and Roles
CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR NOT NULL,
    api_key VARCHAR UNIQUE,
    opa_endpoint VARCHAR NOT NULL,
    status VARCHAR DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS application_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_id UUID REFERENCES applications(id),
    name VARCHAR NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(app_id, name)
);

-- User/Org application roles
CREATE TABLE IF NOT EXISTS user_app_roles (
    user_id UUID REFERENCES users(id),
    app_id UUID REFERENCES applications(id),
    role_name VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, app_id, role_name)
);

-- OPA policies
CREATE TABLE IF NOT EXISTS opa_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_id UUID REFERENCES applications(id),
    name VARCHAR NOT NULL,
    rego_policy TEXT NOT NULL,
    version INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Policy distribution status
CREATE TABLE IF NOT EXISTS policy_sync_status (
    app_id UUID REFERENCES applications(id),
    version INTEGER,
    synced_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (app_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_cognito_sub ON users(cognito_sub);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_app_roles_user_id ON user_app_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_app_roles_app_id ON user_app_roles(app_id);

-- Insert sample data for development
INSERT INTO organizations (id, name, sso_provider, idp_type, callback_url) 
VALUES 
    (uuid_generate_v4(), 'Development Org', 'cognito', 'SAML', 'http://localhost:8081/auth/callback'),
    (uuid_generate_v4(), 'Test Company', 'cognito', 'OIDC', 'http://localhost:8081/auth/callback')
ON CONFLICT DO NOTHING;

-- Insert sample application
INSERT INTO applications (id, name, api_key, opa_endpoint, status)
VALUES 
    (uuid_generate_v4(), 'Demo App', 'demo-api-key-12345', 'http://localhost:8181', 'active')
ON CONFLICT DO NOTHING;

-- Log initialization completion
INSERT INTO organizations (name, sso_provider) 
SELECT 'DB_INIT_COMPLETE', 'timestamp_' || EXTRACT(epoch FROM CURRENT_TIMESTAMP)
WHERE NOT EXISTS (
    SELECT 1 FROM organizations WHERE sso_provider LIKE 'timestamp_%'
);

COMMIT;
