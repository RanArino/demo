# User Roles Document

## Overview

| Role      | Space Functionality                                                |
| :-------- | :--------------------------------------------------------------- |
| Owner     | Full control: manage users, roles, billing, delete spaces        |
| Admin     | Like Owner but no billing access; can audit activity logs        |
| Editor    | Create/edit/delete content, invite users, configure space settings |
| Commenter | Read + comment only (ideal for stakeholders / QA)                |
| Viewer    | Read-only                                                        |
| Guest     | Temporary access (expires), limited to specific docs or pages    |
| Auditor   | Read-only + access to audit/tracking logs                        |

## Details

### Owner
- **Permissions:** Full control over the spaces, including user management (inviting, removing, changing roles), role definition, billing information, and the ability to delete the entire space.
- **Use Case:** Typically the creator of the space or the primary account holder.

### Admin
- **Permissions:** All permissions of an Owner *except* managing billing and deleting the space. Can access and review audit logs.
- **Use Case:** Trusted individuals responsible for day-to-day management and security monitoring without needing financial access.

### Editor
- **Permissions:** Can create, modify, and delete content within the space. Can invite new users (typically as Viewer, Commenter, or Editor). Can configure certain space-level settings (e.g., notifications, integrations).
- **Use Case:** Core team members actively working on the content within the space.

### Commenter
- **Permissions:** Can view all content within the space and add comments. Cannot make edits or change settings.
- **Use Case:** Stakeholders, reviewers, or QA personnel who need to provide feedback but not edit directly.

### Viewer
- **Permissions:** Can view all content within the space. Cannot comment, edit, or change settings.
- **Use Case:** Users who need read-only access for information purposes.

### Guest
- **Permissions:** Temporary access, typically limited to specific documents or pages designated by an Owner, Admin, or Editor. Access automatically expires after a set period. Permissions within the designated content are usually Read-only or Commenter.
- **Use Case:** External collaborators, clients, or temporary team members needing access to specific information for a limited time.

### Auditor
- **Permissions:** Read-only access to all content within the space. Additionally, has specific access to view audit and activity tracking logs. Cannot make any changes.
- **Use Case:** Compliance officers, security personnel, or external auditors needing to review activity and content without modification rights.