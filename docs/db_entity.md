# Database Schema for Knowledge Space

## `SPACE`: Knowledge Space Entity

| Field Name            | Data Type      | Description                                                  |
|-----------------------|----------------|--------------------------------------------------------------|
| `id`                  | UUID            | Unique identifier for the knowledge space.                  |
| `title`               | String          | The primary name of the space.                               |
| `description`         | Text            | A brief summary of the space's purpose or content.          |
| `icon`                | String          | A visual identifier for the space (URL or emoji).           |
| `cover_image`         | String          | URL of the background image for visual appeal.              |
| `keywords`            | Array of Strings| Tags to help categorize and search for the space.           |
| `owner_id`            | UUID            | User ID of the person who created and owns the space.       |
| `created_at`          | Timestamp       | Timestamp of when the space was created.                    |
| `created_by`          | UUID            | User ID of the creator (initially the same as `owner_id`).  |
| `last_updated_at`     | Timestamp       | Timestamp of the last modification to the space's metadata.  |
| `last_updated_by`     | UUID            | User ID of the person who last made an update.              |
| `document_count`      | Integer         | Number of documents uploaded/contained within the space.    |
| `total_size_bytes`    | Integer         | Total storage space consumed by the documents in the space.  |
| `visibility`          | String          | Defines the default access level (e.g., private, shared, public).   |
| `shared_with`         | Array of Objects | List of users/groups and their roles for this space.        |
| `guest_access_enabled` | Boolean         | Indicates if temporary guest access is allowed.             |
| `guest_access_expiry` | Timestamp       | Default expiry duration for guest links.                     |
| `status`              | String          | Current lifecycle state of the space (e.g., active, archived). |
| `processing_status`   | String          | Indicates the state of document processing in background; processing, completed, failed. |


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