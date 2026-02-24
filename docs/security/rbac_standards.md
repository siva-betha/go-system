# RBAC & Security Standards

This document outlines the security architecture and Role-Based Access Control (RBAC) standards for the semiconductor etch monitoring platform.

## 1. Identity Management
- **User IDs**: All users are identified by immutable UUIDs (v4).
- **Authentication**: Stateless JWT-based authentication via `Authorization: Bearer <token>` header.
- **Refresh Tokens**: Database-backed refresh token rotation (single-use tokens) to prevent session hijacking.
- **Hashing**: Password hashing using Argon2id with 64MB memory, 1 iteration, and 4 degrees of parallelism.

## 2. Access Control Model
The platform uses a hybrid RBAC and Resource-Action permission model.

### 2.1 Default Roles
| Role | Description | Typical Permissions |
| :--- | :--- | :--- |
| `admin` | System administrator | Full access, user management, audit review. |
| `engineer` | Process engineer | Recipe management, data export, machine config. |
| `operator` | Tool operator | Real-time monitoring, dashboard viewing. |
| `machine` | M2M / API Service | Scoped access via API Key (e.g., `influx:read`). |

### 2.2 Permissions Structure
Permissions are stored in the JWT claims and evaluated by middleware:
- **Resource**: The logical object (e.g., `user`, `recipe`, `machine`).
- **Action**: The operation (e.g., `create`, `read`, `update`, `delete`).

## 3. Sensitive Action & Approval Workflow
Certain actions require **Dual Authorization** (Two-Person Rule). These are routed through the `pending_approvals` system.

### 3.1 Protected Actions
The following actions require explicit approval from an `admin`:
- `recipe:delete`: Permanent removal of process signatures.
- `user:delete`: Account termination.
- `machine:update`: Significant changes to PLC communication parameters.

## 4. Audit Logging
Every action that modifies state is logged to the `audit_logs` table.
- **Payload**: Includes requester UUID, action, resource, IP address, and JSON diff of changed values.
- **Retention**: Standards require 2-year retention for industrial compliance.

## 5. M2M Security (API Keys)
Service accounts and external integrations must use API Keys.
- **Format**: `prefix_secret` (e.g., `ab12c3_...`).
- **Revocation**: Keys can be instantly revoked in the Admin Center.
- **Scoping**: Keys must be restricted to minimal required scopes (e.g., `telemetry:read`).
