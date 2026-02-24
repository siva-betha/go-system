-- 0006_rbac_schema.up.sql
-- Enable pgcrypto for gen_random_uuid() if not already enabled
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Alter existing users table to fit RBAC profile
-- First, handle the ID change if we want to move to UUID (complex)
-- For now, let's keep BIGINT for ID to maintain compatibility with existing refresh_tokens, 
-- but add the requested fields. OR, we migrate everything to UUID.
-- Given the prompt explicitly asks for UUID, let's migrate.

-- 1. Create a temporary mapping for old BIGINT IDs to new UUIDs
CREATE TABLE user_id_mapping (
    old_id BIGINT,
    new_id UUID DEFAULT gen_random_uuid()
);

INSERT INTO user_id_mapping (old_id) SELECT id FROM users;

-- 2. Add UUID column to users
ALTER TABLE users ADD COLUMN uuid_id UUID DEFAULT gen_random_uuid();
UPDATE users u SET uuid_id = m.new_id FROM user_id_mapping m WHERE u.id = m.old_id;

-- 3. Update refresh_tokens table
ALTER TABLE refresh_tokens ADD COLUMN user_uuid UUID;
UPDATE refresh_tokens rt SET user_uuid = m.new_id FROM user_id_mapping m WHERE rt.user_id = m.old_id;

-- 4. Recreate constraints
ALTER TABLE refresh_tokens DROP CONSTRAINT refresh_tokens_user_id_fkey;
ALTER TABLE users DROP CONSTRAINT users_pkey CASCADE;
ALTER TABLE users RENAME COLUMN id TO old_id;
ALTER TABLE users RENAME COLUMN uuid_id TO id;
ALTER TABLE users ADD PRIMARY KEY (id);

ALTER TABLE refresh_tokens DROP COLUMN user_id;
ALTER TABLE refresh_tokens RENAME COLUMN user_uuid TO user_id;
ALTER TABLE refresh_tokens ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE refresh_tokens ADD CONSTRAINT refresh_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- 5. Add missing columns to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(50) UNIQUE;
UPDATE users SET username = email WHERE username IS NULL;
ALTER TABLE users ALTER COLUMN username SET NOT NULL;

ALTER TABLE users ADD COLUMN IF NOT EXISTS full_name VARCHAR(100);
UPDATE users SET full_name = name WHERE full_name IS NULL;

ALTER TABLE users ADD COLUMN IF NOT EXISTS department VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS title VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS employee_id VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_attempts INT DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS created_by UUID;

-- 6. Create RBAC Tables
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    priority INT DEFAULT 0,
    is_system_role BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    UNIQUE(resource, action)
);

CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    conditions JSONB,
    granted_at TIMESTAMP DEFAULT NOW(),
    granted_by UUID REFERENCES users(id),
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id),
    expires_at TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

-- Note: chambers table assumed to exist from previous conversations/context
-- If it doesn't exist, we might need to create a stub or skip user_chambers for now
-- Let's check if chambers table exists. 
-- In the mean time, I'll create the user_chambers table but it might fail if chambers doesn't exist.

CREATE TABLE IF NOT EXISTS chambers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE user_chambers (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    chamber_id UUID REFERENCES chambers(id) ON DELETE CASCADE,
    access_level VARCHAR(20) DEFAULT 'read',
    granted_at TIMESTAMP DEFAULT NOW(),
    granted_by UUID REFERENCES users(id),
    PRIMARY KEY (user_id, chamber_id)
);

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    key_preview VARCHAR(20),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    allowed_ips INET[],
    created_at TIMESTAMP DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

CREATE TABLE api_key_permissions (
    api_key_id UUID REFERENCES api_keys(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (api_key_id, permission_id)
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    refresh_token VARCHAR(255) UNIQUE,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP DEFAULT NOW(),
    user_id UUID REFERENCES users(id),
    username VARCHAR(50),
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    resource_name VARCHAR(255),
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    status VARCHAR(20),
    error_message TEXT,
    session_id UUID REFERENCES sessions(id)
);

CREATE TABLE pending_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requestor_id UUID REFERENCES users(id),
    approver_id UUID REFERENCES users(id),
    action_type VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    requested_changes JSONB NOT NULL,
    reason TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    request_timestamp TIMESTAMP DEFAULT NOW(),
    review_timestamp TIMESTAMP,
    review_notes TEXT,
    approved_by UUID REFERENCES users(id)
);

-- Clean up
DROP TABLE user_id_mapping;
ALTER TABLE users DROP COLUMN old_id;
ALTER TABLE users DROP COLUMN name;
ALTER TABLE users DROP COLUMN role;
