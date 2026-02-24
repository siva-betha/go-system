-- 0007_rbac_seed.up.sql

-- Insert default roles
INSERT INTO roles (name, description, is_system_role) VALUES
('admin', 'Full system access', true),
('process_engineer', 'Can modify recipes and parameters', true),
('maintenance_tech', 'Can view status and clear alarms', true),
('operator', 'Can start/stop processes and view dashboards', true),
('quality_engineer', 'Can view SPC and create golden prints', true),
('auditor', 'Read-only access with audit view', true),
('viewer', 'Basic read-only access', true);

-- Insert default permissions
INSERT INTO permissions (resource, action, description) VALUES
('recipes', 'create', 'Create new recipes'),
('recipes', 'read', 'View recipes'),
('recipes', 'update', 'Modify recipes'),
('recipes', 'delete', 'Delete recipes'),
('recipes', 'approve', 'Approve recipe changes'),
('chambers', 'read', 'View chamber data'),
('chambers', 'write', 'Control chamber parameters'),
('chambers', 'configure', 'Modify chamber configuration'),
('users', 'create', 'Create users'),
('users', 'read', 'View users'),
('users', 'update', 'Modify users'),
('users', 'delete', 'Delete users'),
('audit', 'read', 'View audit logs'),
('data', 'export', 'Export data'),
('data', 'import', 'Import data'),
('system', 'configure', 'Configure system settings');

-- Assign all permissions to admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin';

-- Assign process engineer permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'process_engineer' 
AND p.resource IN ('recipes', 'chambers')
AND p.action IN ('create', 'read', 'update');

-- Assign existing users to admin role for safety
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u, roles r
WHERE r.name = 'admin';
