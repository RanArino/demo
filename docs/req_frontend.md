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


## Space Page (`spaces/{space_id}`)

### User Functinalities:
- Owner/Admin can delete the space from the setting button. (Only Owner has explicit permission to delete spaces)
- Owner/Admin can edit the title, image, description, and keywords of the space from the setting button. (Based on Owner/Admin's control) setting button.
- Editors can upload and delete documents, use of chat session.
- Commenters can comment on canvas nodes and paths
- Viewers can view the canvas and document preview, also use chat session.
- Guest can view the canvas and document preview, but not use chat session.

### Visual/Action Requirement:
- Option (1): The screen is split to 2 parts: left is the 2d/3d tree visualization canvas, the right-side component is based on two parts; document preview (upper) and chat session (lower).
- Option (2): The screen is split to 3 parts, left is the 2d/3d tree visualization canvas, middle is the document preview, the right-side component is based on two parts; chat session.

#### (1) 2d/3d tree visualization on canvas:
- **Default Structure:** Displays a 2D/3D representation of a hierarchical 3-layer graph (e.g., Documents, Clusters, Chunks).
- **Hierarchy Preservation (Default View):** The vertical order of layers (Documents highest, Chunks lowest) must be strictly maintained at all times and must not invert. Imagine that its hierarchical 3-layer graph is an object; and its object is only turning based on yaw axis. Users cannot rotate the object in pitch and roll. 
- **Layer Representation:** 
    - Each layer is visually contained within its own semi-transparent elliptical plane. 
    - These planes are parallel and maintain their relative vertical positions.
    - Hovering the mouse cursor over an elliptical plane visually highlights that specific plane (e.g., increased border opacity, subtle fill color change) to indicate interactivity. Simultaneously, the documents layer metadata will be shown up.
- **Default View:** Initializes with a moderate top-down perspective (e.g., ~ -15° to -20° pitch relative to the horizon) showing all layers clearly.
- **Node:** 
    - Placement: Nodes belonging to a layer are positioned on that layer's elliptical plane.
    - Selection (Click - Applies in Both Views):
        - Triggers a smooth animation (using appropriate rotation - primarily Yaw in default view - and Zoom/Dolly) to center the selected node in the viewport.
        - Animation in the default view must not introduce Pitch or Roll changes.
- **Layer Isolation View**:
    - **Activation**: Clicking directly on an elliptical plane transitions the view smoothly to display only the nodes and intra-layer connections belonging to that selected layer (e.g., clicking the "Documents" ellipse shows only Document nodes). The other layers and their ellipses are hidden.
    - **Isolated Layout (Selectable 2D/3D)**: The layout of the nodes within this isolated view can be presented (or potentially toggled by the user) in either a flat 2D arrangement or a 3D spatial arrangement. (The default state and the specific mechanism for toggling 2D/3D need to be defined).
    - **Navigation (Isolated View)**: Standard navigation controls like Yaw rotation (horizontal drag) and Zoom remain functional, adapted for the single-layer context.
    - **Exiting Isolation**: A clear and easily accessible mechanism must be provided to return to the default 3-layer view (e.g., a dedicated "Show All Layers" button appearing in this mode, clicking the empty canvas background, an Escape key press).
- **Navigation (Default 3-Layer View)**:
    - Rotation (User Drag): User click-and-drag horizontally only performs Yaw rotation (around the vertical Z-axis). Pitch and Roll via drag are disabled.
    - Angle Adjustment (User Scroll): Mouse scrolling (when pointer is over this panel) adjusts the apparent viewing angle of the layers (making them look flatter or more edge-on) within a limited range that does not invert the hierarchy.
    - Zoom: Standard zoom (dolly) functionality is enabled (e.g., via scroll wheel).
    - Vertical Pan: Disabled.