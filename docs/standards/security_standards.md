# Security Standards

Guidelines for ensuring data integrity and system security in the semiconductor fab environment.

## 1. Authentication
- **Mechanism**: JWT (JSON Web Token) with RS256 signing.
- **Rotation**: Refresh tokens required; 1-hour expiry for access tokens.
- **MFA**: Recommended for Administrative accounts.

## 2. Authorization (RBAC)
| Role | Permissions |
|---|---|
| **Admin** | Full system configuration, user management, audit logs. |
| **Engineer** | Recipe management, alert configuration, analytics tuning. |
| **Technician** | Live monitoring, manual overrides, shift reports. |
| **Viewer** | Read-only access to dashboards and reports. |

## 3. Data Security
- **Encryption**: AES-256 at rest for database files.
- **Transmission**: TLS 1.3 enforced for all Edge-to-Cloud traffic.
- **Secrets**: No secrets in source code; use encrypted environment variables.

## 4. Audit Logging
- All configuration changes (recipes/settings) must be logged with:
  - User ID
  - Timestamp
  - Before/After state
  - Rationale
