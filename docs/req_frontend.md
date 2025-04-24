# Frontend Requirements

## Homepage (`/`) or Spaces (` spaces/ `)
- homepage after the user signed in and logged in 
- the collection of the knowledge spaces
- examples/references: homepage of the NotebookLM, Notion Database

### User Functinalities:
- Owner/Admin can create a new knowledge space, a collection of contents/documents. (Based on Owner's full control, Admin's management capabilities)
- Owner/Admin can delete a space from the setting button. (Only Owner has explicit permission to delete spaces)
- Owner/Admin can edit the title, image, description, and keywords of the space from the setting button. (Based on Owner/Admin's control) setting button.

### Visual Requirement:
1. Gallery view:
    - The size of each space card can be changed to small, medium, large.
    - The created space displays the `icon`, `title`, `cover_image`, `created_at`, `document_count`; on hover, showing the descriptions and keywords.

2. List/Table view:
    - The created space displays the following metadata; 
        - (individual) icon, title, keywords, description(?), num of document_count, created_at, updated_at, visibility
        - (team) icon, title, keywords, description(?), num of shared_with, document_count, created_at, updated_at
    - `icon`(upload), `title`(text), `description`(text), `keywords`(multi-select), `visibility`(select) can be directly editable by clicking each section.
    - if user click number of `shared_with`, the shared_with users list will be shown up; add new users from both spread sheet and manual way (idea).
    - if user click number of `document_count`, the document list will be shown up.
    - if user click the column names of `title`, num of `shared_with`, `document_count`, `created_at`, `updated_at`, do the sorting; make sure to show up the sign of ascending or descending.

3. Tree(Canvas) view:
    - The created spaces are displayed as its tree structure on the canvas.
    -  Showing `title`, `keywords`, `description`, `created_at` during hovering.


