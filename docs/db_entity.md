# Database Schema for Knowledge Space

## `USER`: User Entity

| Field Name            | Data Type      | Description                                                  |
|-----------------------|----------------|--------------------------------------------------------------|
| `id`                  | UUID            | Unique identifier for the user.                             |
| `username`            | String          | Unique username for the user, used for login.              |
| `email`               | String          | User's email address, used for notifications and login.    |
| `password_hash`       | String          | Hashed password for secure authentication.                  |
| `role`                | String          | User's primary system role (e.g., Owner, Admin, Editor, etc., influencing default permissions). Note: Specific space roles are managed separately. |
| `created_at`          | Timestamp       | Timestamp of when the user account was created.            |
| `last_login_at`       | Timestamp       | Timestamp of the last time the user logged in.             |
| `status`              | String          | Current status of the user account (e.g., active, suspended, pending_verification, deleted). |
| `profile_picture_url` | String          | URL of the user's profile picture.                          |
| `full_name`           | String          | User's full name.                                           |
| `preferences`         | JSON/Text       | User-specific settings (e.g., theme, notification preferences). |
| `last_updated_at`     | Timestamp       | Timestamp of the last modification to the user's profile.   |