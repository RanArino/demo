## Knowledge Service: The Brain & Librarian

This service is the **stateful core** of the application. It knows everything about the *what*, *who*, and *where* of the data, but not *how* to transform it. It is the single source of truth for all metadata and relationships.

### **Core Business Logic:**

* **1. Workspace and Space Management:**
    * CRUD operations for `Spaces` (the top-level containers for content; each space is a distinct group of contents).
    * Manages the relationship between users and spaces (e.g., which users are members of a space).
    * Handles settings and metadata specific to a space (e.g., space name, icon, public/private status).

* **2. Content Metadata Management:**
    * Manages the lifecycle of all content records. This includes creating and storing metadata for every uploaded file, note, or page.
    * Stores owner_id`: The `user_id` of the end-user who created the content. This is mandatory for enforcing user-level security.
    * Stores pointers to the file objects in object storate (like Cloudflare R2 for demo, but we will use S3 in production); both the `original_blob_url` and the `processed_blob_url` will be linked to the content record.
    * Maintains the `processing_status` (`PROCESSING`, `COMPLETED`, `FAILED`) for each content item.
    * Handles file versioning logic. When a file is updated, it creates a new version record.

* **3. Authorization and Access Control:**
    * Enforces business rules based on user identity. After receiving the validated `user_id` and `role` from the Auth Service, this service determines if that user can perform an action.
    * Checks `owner_id` for all CRUD operations on spaces and content.
    * Manages permissions at a more granular level in the future (e.g., read, write, comment permissions per-page).

4.  **Knowledge Graph & Content Relationships:**
    * **Maintains the Graph of Links:** When a user creates a link (e.g., `[[Link to another page]]`) in a Markdown document, the service parses this and stores a direct relationship (`document_A -> links_to -> document_B`) in its database. This graph enables critical features like backlinks.
    * **Manages Rich Mentions:** Handles `@` mentions of other pages, users, or dates, storing these relationships in the database to be rendered as rich links.
    * **Manages Taxonomies:** Handles the creation and association of tags and other categories to content.

* **5. Annotation and Highlight Management (Future):**
    * Stores the data for annotations and highlights. When a user highlights text on a document, the Knowledge Service stores the highlighted content, its position, the associated user ID, and any comments, linking it all to the content's unique ID.


### Core Workflow
Upload Request: A user sends a request to the Knowledge Service to upload a document to a specific space.

Initial Record Creation: The Knowledge Service creates a CONTENT_SOURCES record in its PostgreSQL database. It sets the status field to PROCESSING. It then uploads the original, unconverted file to Cloudflare R2.

Delegation via gRPC: The Knowledge Service then acts as a client. It makes a gRPC call to the Document Process Service, sending it a job containing a pre-signed URL for the original file on R2.

Stateless Processing: The Document Process Service receives the gRPC call. It fetches the file from R2, performs the conversion to Markdown, and uploads the new, processed Markdown file back to a different location in R2.

Completion Notification via gRPC: Upon successful conversion, the Document Process Service makes a gRPC call back to the Knowledge Service to report that the job is complete, returning the processed_blob_hash or the new URL.

Finalizing State: The Knowledge Service receives this completion call, updates the CONTENT_SOURCES record with the new processed_blob_hash, and changes the status to PROCESSED.

