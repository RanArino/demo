# Requirements Document: Knowledge Service Fixes

## 1. Overview
This document defines the requirements for fixing multiple issues identified in the knowledge service and related frontend components. The main objective is to improve system stability and usability through bug fixes and UX improvements.

## 2. Requirements List

### Requirement 1: Modal Display for Document List
**User Story:**
> As a space management user, when I click on the document count in the Documents column of the list view on the `/spaces` page, I want to see a modal displaying the list of documents contained in that space. This is because I want to quickly understand the content without navigating to another page.

**Acceptance Criteria:**
```gherkin
GIVEN I have opened the list view page at `/spaces`
WHEN I click on the document count (number) in the "Documents" column for a specific space
THEN a modal dialog containing the document list for that space should be displayed
AND the modal should include loading state and error state displays
AND documents should be displayed in a compact list format similar to the existing `ContentSourceCard` component
```

### Requirement 1.1: Enhanced Document Modal UI/UX
**User Story:**
> As a space management user, when I view the document list modal, I want it to be large enough to comfortably view document details and have a table layout similar to the spaces list for consistency. This is because I want to efficiently review document information and access preview functionality.

**Acceptance Criteria:**
```gherkin
GIVEN I have opened the document list modal from the spaces list view
WHEN the modal is displayed
THEN the modal should occupy 80% of the screen width and height
AND the modal should be centered on the screen
AND the document list should be displayed in a table format similar to the spaces list view
AND the table should have the following columns: Title, Keywords, Content Summary, MIME Type (File Type), Size (Bytes), and Preview Button
AND each row should be non-editable (read-only)
AND the Preview Button should open the document in a new web tab/window
```

### Requirement 2: Setting `created_by` Field During Space Creation
**User Story:**
> As a developer/system administrator, when a new space is created, I want the `created_by` and `last_updated_by` fields to be correctly recorded in the database with the creator's ID. This is because I need to distinguish between owners and creators for future sharing features and auditing purposes.

**Acceptance Criteria:**
```gherkin
GIVEN a user sends a request to create a new space
WHEN `SpaceService.CreateSpace` is called
THEN a new record should be created in the `spaces` table in the database
AND the `created_by` and `last_updated_by` fields of that record should be set to the UUID of the user who sent the request
```

### Requirement 3: Resolving Conflict Between Inline Editing and Navigation in List View
**User Story:**
> As a space management user, when I click on a space title in the list view on the `/spaces` page, I want to be able to edit the title inline without page navigation. This is because I want to update multiple space information intuitively and quickly.

**Acceptance Criteria:**
```gherkin
GIVEN I have opened the list view page at `/spaces`
WHEN I click on a space title
THEN the title should switch to an editable input field
AND no page navigation should occur
WHEN I click on Description or Keywords text
THEN each field should become editable inline
AND the "Actions" column (pencil icon that opens edit modal) should be removed from the table
```

### Requirement 4: Database Update for Space Statistics (Document Count and Total Size)
**User Story:**
> As a developer/system administrator, when documents are added or deleted, I want the related space's statistics (`document_count`, `total_size_bytes`) to be automatically updated in the database. This is because I want to use accurate data for analysis and report generation.

**Acceptance Criteria:**
```gherkin
GIVEN a user uploads a document and a `ContentSource` is successfully created
WHEN `ContentService.CreateUploadURL` is executed
THEN the corresponding `spaces` table record's `document_count` should increase by 1 and `total_size_bytes` should increase by the uploaded file size

GIVEN a user deletes a document
WHEN `ContentService.DeleteContentSource` is executed
THEN the corresponding `spaces` table record's `document_count` should decrease by 1 and `total_size_bytes` should decrease by the deleted file size
AND a backfill process should be provided to recalculate and update statistics for existing spaces based on current `contentsource` data
```

### Requirement 5: Default Value Setting for Space `access_level` and UI Updates
**User Story:**
> As a space management user, I want newly created spaces to have a default access level of "private" and be able to easily change the access level from the UI. This is because I want to prevent unintended information disclosure and manage sharing settings through intuitive operations.

**Acceptance Criteria:**
```gherkin
GIVEN a user creates a new space
WHEN `SpaceService.CreateSpace` is executed
THEN the `access_level` field of the created record in the `spaces` table should be set to `'private'`

GIVEN I have opened the gallery view or list view at `/spaces`, or the `spaces/[spaceId]` page
WHEN I click on the badge or sharing icon indicating access level
THEN a common modal dialog for changing access level (e.g., private, public) should open
AND the new access level selected in the modal should be saved to the database
AND a backfill process should be provided to update existing records with empty or NULL `access_level` to `'private'`
```

### Requirement 6: UX Improvement for Left Sidebar
**User Story:**
> As a space browsing user, when the sidebar is closed on the `/spaces/[spaceId]` page, I want an icon to always be displayed for opening it. This is because I want to easily discover and use the functionality even if I don't know that the sidebar opens on hover.

**Acceptance Criteria:**
```gherkin
GIVEN I have opened the `/spaces/[spaceId]` page
WHEN the left sidebar is closed and not pinned
THEN an icon button (e.g., `PanelRightOpen`) for opening the sidebar should be displayed in the top-left corner of the screen
WHEN I click that icon button
THEN the left sidebar should be displayed
WHEN the sidebar is open or pinned
THEN that icon button should be hidden
```

### Requirement 7: Fix ContentSource Size Storage Issue
**User Story:**
> As a system administrator, when documents are uploaded, I want the file size information to be correctly stored and retrieved from the database. This is because I need accurate file size data for storage management, statistics, and the document size display in the UI.

**Acceptance Criteria:**
```gherkin
GIVEN a user uploads a document with a specific file size
WHEN the ContentSource is created in the database
THEN the size_bytes field should be populated with the correct file size value
AND when the ContentSource is retrieved from the database
THEN the size_bytes field should return the correct stored value
AND the document list modal should display accurate file sizes
```
