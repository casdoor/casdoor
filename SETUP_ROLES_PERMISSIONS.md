# How to Setup Roles and Permissions in Casdoor

This guide explains how to properly configure roles and permissions in Casdoor to control access to your applications and resources.

## Table of Contents

- [Overview](#overview)
- [Core Concepts](#core-concepts)
- [Setting Up Roles](#setting-up-roles)
- [Setting Up Permissions](#setting-up-permissions)
- [Best Practices](#best-practices)
- [Common Use Cases](#common-use-cases)
- [Troubleshooting](#troubleshooting)

## Overview

Casdoor uses a flexible Role-Based Access Control (RBAC) system powered by [Casbin](https://casbin.org/). This system allows you to:

- Define roles and assign them to users or groups
- Create granular permissions for resources
- Control access using policies based on the Casbin model
- Support hierarchical role inheritance

## Core Concepts

### Roles

A **Role** is a collection of users or groups that share common access requirements. Roles can:
- Be assigned to individual users
- Be assigned to groups (all users in the group inherit the role)
- Inherit from other roles (hierarchical roles)
- Be associated with specific domains

**Key Fields:**
- **Owner**: The organization that owns this role
- **Name**: Unique identifier for the role
- **Display Name**: Human-readable name
- **Description**: Purpose of the role
- **Users**: List of user IDs assigned to this role (format: `owner/username`)
- **Groups**: List of group IDs whose members inherit this role
- **Roles**: List of parent roles (for role hierarchy)
- **Domains**: List of domains where this role applies
- **IsEnabled**: Whether the role is currently active

### Permissions

A **Permission** defines what actions can be performed on specific resources. Permissions specify:
- Who can access (users, groups, or roles)
- What resources they can access
- What actions they can perform
- Whether access is allowed or denied

**Key Fields:**
- **Owner**: The organization that owns this permission
- **Name**: Unique identifier for the permission
- **Display Name**: Human-readable name
- **Description**: Purpose of the permission
- **Users**: Direct user assignments (format: `owner/username` or `*` for all)
- **Groups**: Group assignments (format: `owner/groupname`)
- **Roles**: Role assignments (format: `owner/rolename`)
- **Domains**: Domain restrictions
- **Model**: The Casbin model to use (e.g., `built-in/user-model-built-in`)
- **Adapter**: The Casbin adapter for policy storage
- **ResourceType**: Type of resource (e.g., "Application", "Custom")
- **Resources**: List of resource IDs (e.g., application IDs)
- **Actions**: List of allowed actions (e.g., "Read", "Write", "Admin")
- **Effect**: "Allow" or "Deny"
- **IsEnabled**: Whether the permission is currently active

### Models

A **Model** defines the access control policy syntax using Casbin's PERM metamodel. Casdoor comes with built-in models:

#### User Model (user-model-built-in)
```
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

This model supports:
- **sub**: Subject (user or role)
- **obj**: Object (resource)
- **act**: Action (Read, Write, Admin, etc.)

## Setting Up Roles

### Step 1: Create a Role via UI

1. Log in to Casdoor as an administrator
2. Navigate to **Organizations** → Select your organization → **Roles**
3. Click **Add** to create a new role
4. Fill in the required fields:
   - **Name**: Unique identifier (e.g., `developer`, `manager`)
   - **Display Name**: User-friendly name (e.g., "Developer")
   - **Description**: Brief description of the role's purpose
5. Assign users or groups:
   - **Users**: Add individual users by their ID (format: `owner/username`)
   - **Groups**: Add groups whose members should inherit this role
6. (Optional) Set up role hierarchy:
   - **Roles**: Add parent roles that this role inherits from
7. Click **Save**

### Step 2: Create a Role via API

```bash
curl -X POST http://your-casdoor-url/api/add-role \
  -H "Content-Type: application/json" \
  -d '{
    "owner": "your-org",
    "name": "developer",
    "displayName": "Developer",
    "description": "Development team members",
    "users": ["your-org/john", "your-org/jane"],
    "groups": ["your-org/dev-team"],
    "roles": [],
    "domains": [],
    "isEnabled": true
  }'
```

### Step 3: Create a Role via init_data.json

Add role configuration to your `init_data.json`:

```json
{
  "roles": [
    {
      "owner": "your-org",
      "name": "developer",
      "displayName": "Developer",
      "description": "Development team members",
      "users": ["your-org/john", "your-org/jane"],
      "groups": ["your-org/dev-team"],
      "roles": [],
      "domains": [],
      "isEnabled": true
    }
  ]
}
```

## Setting Up Permissions

### Step 1: Ensure You Have a Model

Permissions require a Casbin model. You can use the built-in models or create custom ones:

- **user-model-built-in**: Standard RBAC model for application resources
- **api-model-built-in**: API-level access control

To create a custom model:
1. Navigate to **Models** in Casdoor
2. Click **Add** and define your Casbin model
3. Save the model

### Step 2: Create a Permission via UI

1. Navigate to **Organizations** → Select your organization → **Permissions**
2. Click **Add** to create a new permission
3. Fill in the required fields:
   - **Name**: Unique identifier (e.g., `read-user-data`)
   - **Display Name**: User-friendly name
   - **Description**: What this permission allows
4. Configure access subjects (at least one required):
   - **Users**: Direct user assignments (e.g., `owner/username` or `*`)
   - **Groups**: Group assignments
   - **Roles**: Role assignments (e.g., `owner/developer`)
5. Configure the policy:
   - **Model**: Select a Casbin model (e.g., `built-in/user-model-built-in`)
   - **Resource Type**: "Application" or "Custom"
   - **Resources**: Select resources (e.g., application IDs or resource names)
   - **Actions**: Define allowed actions (e.g., ["Read", "Write", "Admin"])
   - **Effect**: "Allow" (or "Deny" for explicit restrictions)
6. Click **Save**

### Step 3: Create a Permission via API

```bash
curl -X POST http://your-casdoor-url/api/add-permission \
  -H "Content-Type: application/json" \
  -d '{
    "owner": "your-org",
    "name": "developer-app-access",
    "displayName": "Developer App Access",
    "description": "Allows developers to read and write app data",
    "users": [],
    "groups": [],
    "roles": ["your-org/developer"],
    "domains": [],
    "model": "built-in/user-model-built-in",
    "adapter": "",
    "resourceType": "Application",
    "resources": ["your-org/app-built-in"],
    "actions": ["Read", "Write"],
    "effect": "Allow",
    "isEnabled": true,
    "submitter": "admin",
    "approver": "admin",
    "state": "Approved"
  }'
```

### Step 4: Create a Permission via init_data.json

```json
{
  "permissions": [
    {
      "owner": "your-org",
      "name": "developer-app-access",
      "displayName": "Developer App Access",
      "description": "Allows developers to read and write app data",
      "users": [],
      "groups": [],
      "roles": ["your-org/developer"],
      "domains": [],
      "model": "built-in/user-model-built-in",
      "adapter": "",
      "resourceType": "Application",
      "resources": ["your-org/your-app"],
      "actions": ["Read", "Write"],
      "effect": "Allow",
      "isEnabled": true,
      "submitter": "admin",
      "approver": "admin",
      "state": "Approved"
    }
  ]
}
```

## Best Practices

### 1. Use Role-Based Assignments

Instead of assigning permissions directly to users, assign them to roles. This makes management easier:

```
Users → Roles → Permissions → Resources
```

**Example:**
- Create a `developer` role
- Assign users to the `developer` role
- Create permissions that grant the `developer` role access to resources

### 2. Follow the Principle of Least Privilege

Grant only the minimum permissions necessary for users to perform their tasks.

### 3. Use Hierarchical Roles

For complex organizations, use role inheritance:

```json
{
  "roles": [
    {
      "name": "employee",
      "displayName": "Employee",
      "users": [],
      "roles": []
    },
    {
      "name": "developer",
      "displayName": "Developer",
      "users": ["org/john"],
      "roles": ["org/employee"]  // Inherits employee permissions
    },
    {
      "name": "senior-developer",
      "displayName": "Senior Developer",
      "users": ["org/jane"],
      "roles": ["org/developer"]  // Inherits developer permissions
    }
  ]
}
```

### 4. Use Wildcards Carefully

The `*` wildcard can be used for users, resources, or actions:
- `"users": ["*"]` - Grants permission to all users
- `"resources": ["*"]` - Grants access to all resources
- `"actions": ["*"]` - Allows all actions

Use wildcards sparingly and only when necessary.

### 5. Group Related Resources

When creating permissions, group related resources together when they share the same access control requirements.

### 6. Document Your Permissions

Always provide clear descriptions for roles and permissions to help administrators understand their purpose.

### 7. Regular Audits

Periodically review roles and permissions to ensure they remain appropriate as your organization evolves.

## Common Use Cases

### Use Case 1: Basic User Access

**Scenario:** Grant all authenticated users read access to a public application.

**Solution:**
```json
{
  "permissions": [
    {
      "name": "public-read-access",
      "displayName": "Public Read Access",
      "users": ["*"],
      "model": "built-in/user-model-built-in",
      "resourceType": "Application",
      "resources": ["org/public-app"],
      "actions": ["Read"],
      "effect": "Allow",
      "isEnabled": true
    }
  ]
}
```

### Use Case 2: Administrative Role

**Scenario:** Create an admin role with full access to all resources.

**Solution:**
1. Create an `admin` role:
```json
{
  "roles": [
    {
      "name": "admin",
      "displayName": "Administrator",
      "users": ["org/admin-user"],
      "isEnabled": true
    }
  ]
}
```

2. Create an admin permission:
```json
{
  "permissions": [
    {
      "name": "admin-full-access",
      "displayName": "Admin Full Access",
      "roles": ["org/admin"],
      "model": "built-in/user-model-built-in",
      "resourceType": "Application",
      "resources": ["*"],
      "actions": ["Read", "Write", "Admin"],
      "effect": "Allow",
      "isEnabled": true
    }
  ]
}
```

### Use Case 3: Department-Specific Access

**Scenario:** Grant different permissions to HR and Engineering departments.

**Solution:**
1. Create groups for each department
2. Create roles and assign groups:
```json
{
  "roles": [
    {
      "name": "hr-staff",
      "displayName": "HR Staff",
      "groups": ["org/hr-department"],
      "isEnabled": true
    },
    {
      "name": "engineer",
      "displayName": "Engineer",
      "groups": ["org/engineering-department"],
      "isEnabled": true
    }
  ]
}
```

3. Create permissions for each role:
```json
{
  "permissions": [
    {
      "name": "hr-access",
      "roles": ["org/hr-staff"],
      "resources": ["org/hr-app"],
      "actions": ["Read", "Write"],
      "effect": "Allow"
    },
    {
      "name": "engineering-access",
      "roles": ["org/engineer"],
      "resources": ["org/dev-app"],
      "actions": ["Read", "Write", "Admin"],
      "effect": "Allow"
    }
  ]
}
```

### Use Case 4: Temporary Access

**Scenario:** Grant temporary elevated permissions for a specific task.

**Solution:**
1. Create a temporary role
2. Assign users to the role
3. Create permission with the temporary role
4. When task is complete, disable the role or permission using `isEnabled: false`

### Use Case 5: Multi-Application Access

**Scenario:** A user needs different levels of access to multiple applications.

**Solution:**
Create separate permissions for each application:
```json
{
  "permissions": [
    {
      "name": "user-app1-admin",
      "users": ["org/john"],
      "resources": ["org/app1"],
      "actions": ["Read", "Write", "Admin"],
      "effect": "Allow"
    },
    {
      "name": "user-app2-read",
      "users": ["org/john"],
      "resources": ["org/app2"],
      "actions": ["Read"],
      "effect": "Allow"
    }
  ]
}
```

## Troubleshooting

### Permission Not Working

**Check the following:**

1. **Is the permission enabled?**
   - Verify `isEnabled: true` for both role and permission

2. **Is the model correct?**
   - Ensure the model ID exists (e.g., `built-in/user-model-built-in`)
   - Verify the model's matcher supports your use case

3. **Are resource IDs correct?**
   - Resource IDs must match exactly (format: `owner/name`)
   - Check for typos in resource names

4. **Is the user assigned correctly?**
   - Verify user ID format: `owner/username`
   - Check if user is in the specified group or role

5. **Are there conflicting permissions?**
   - Check if a "Deny" effect permission overrides your "Allow" permission

### Role Hierarchy Not Working

**Check the following:**

1. **Role ID format:**
   - Ensure role IDs in the `roles` field use format: `owner/rolename`

2. **Circular dependencies:**
   - Avoid circular role inheritance (Role A → Role B → Role A)

3. **Model support:**
   - Verify the Casbin model has `[role_definition]` section

### Users Not Inheriting Group Permissions

**Check the following:**

1. **Group membership:**
   - Verify users are actually members of the group
   - Check group ID format in role: `owner/groupname`

2. **Role assignment:**
   - Ensure the group is assigned to the role
   - Verify the role is assigned to the permission

### Permission Denied Errors

**Debug steps:**

1. Check Casdoor logs for policy evaluation details
2. Verify the enforcer is using the correct model and adapter
3. Test with a wildcard permission to confirm the system is working
4. Review the Casbin matcher in your model

## Advanced Topics

### Custom Models

For complex scenarios, you can create custom Casbin models:

1. Navigate to **Models** in Casdoor
2. Create a new model with custom request, policy, and matcher definitions
3. Use the custom model in your permissions

**Example: Resource-Type-Based Model**
```
[request_definition]
r = sub, obj, act, res_type

[policy_definition]
p = sub, obj, act, res_type

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act && r.res_type == p.res_type
```

### Domain-Based Access Control

Use domains to implement multi-tenancy:

```json
{
  "roles": [
    {
      "name": "tenant-admin",
      "domains": ["tenant1", "tenant2"],
      "users": ["org/user1"]
    }
  ]
}
```

### Integration with External Systems

Casdoor supports syncing roles and permissions from external systems:
- LDAP/Active Directory groups can be mapped to Casdoor groups
- SCIM protocol for user and group provisioning
- Custom sync adapters for proprietary systems

## API Reference

### Get User Permissions
```bash
GET /api/get-permissions?owner=<org>&user=<username>
```

### Get Role Permissions
```bash
GET /api/get-permissions?owner=<org>&role=<rolename>
```

### Check Permission
```bash
POST /api/enforce
{
  "permissionId": "owner/permission-name",
  "user": "owner/username",
  "resource": "owner/resource-name",
  "action": "Read"
}
```

## Additional Resources

- [Casdoor Documentation](https://casdoor.org)
- [Casbin Documentation](https://casbin.org)
- [Casdoor API Documentation](https://door.casdoor.com/swagger)
- [Online Demo](https://demo.casdoor.com)

## Support

For additional help:
- [GitHub Discussions](https://github.com/casdoor/casdoor/discussions)
- [Discord Community](https://discord.gg/5rPsrAzK7S)
- [GitHub Issues](https://github.com/casdoor/casdoor/issues)
