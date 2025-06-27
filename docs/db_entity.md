# Database Schema for Knowledge Space - Version 2

## Core Design Principles

1. **Unified Node Hierarchy**: All entities (content sources, content chunks, clusters) are represented as nodes with flexible parent-child relationships
2. **Separation of Display vs Retrieval**: Clear distinction between text for UI display and text for vector search
3. **Scalable Clustering**: No fixed hierarchy levels - clusters can contain clusters to any depth
4. **Global vs Space Context**: Knowledge exists globally but can be organized differently in each space

---

## Knowledge Service
This service is responsible for managing the core entities of the platform: knowledge spaces and the content sources within them. It handles creation, metadata, storage quotas, and the lifecycle of all ingested content.

### `SPACES`: Knowledge Space Entity
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
| `document_count`      | Integer         | Number of content sources uploaded/contained within the space.    |
| `total_size_bytes`    | BigInt          | Total storage space consumed by the content sources in the space.  |
| `storage_quota_bytes` | BigInt          | Maximum storage allowed for this space (default: 1GB).      |
| `access_level`        | String          | Defines the default access level (private, shared, public). |
| `guest_access_enabled` | Boolean         | Indicates if temporary guest access is allowed.             |
| `guest_access_expiry` | Timestamp       | Default expiry duration for guest links.                     |
| `status`              | String          | Current lifecycle state of the space (active, archived).    |
| `processing_status`   | String          | Indicates the state of document processing in background.    |
| `deleted_at`          | Timestamp       | Soft delete timestamp (NULL if not deleted).                |

### `CONTENT_SOURCES`: Source Content Entity

| Field Name            | Data Type      | Description                                                               |
|-----------------------|----------------|---------------------------------------------------------------------------|
| `id`                  | UUID            | Unique identifier for the content source.                                |
| `owner_id`            | UUID            | User ID of the person who uploaded/added the content.                    |
| `title`               | String          | The primary name of the content.                                         |
| `media_type`          | String          | Type of media content (document, audio, video, web, text). |
| `source`              | String          | The source type (upload, gdrive, onedrive, notion, web, paste).          |
| `original_blob_hash`  | CHAR(64)        | FK to `KNOWLEDGE_CONTENT_BLOBS.blob_hash`. The hash of the original uploaded file. |
| `processed_blob_hash` | CHAR(64)        | FK to `KNOWLEDGE_CONTENT_BLOBS.blob_hash`. The hash of the processed markdown content. |
| `content_summary`          | Text            | AI-generated summary of the content.                                      |
| `keywords`            | Array of Strings| Tags/keywords for the content.                                            |
| `status`              | String          | Current processing state (processing, processed, failed).               |
| `num_chunks`          | Integer         | Number of text chunks generated from this content source.                |
| `total_size_bytes`    | BigInt          | Original file size in bytes. Derived from the blob for convenience.      |
| `created_at`          | Timestamp       | Timestamp of when the content was created.                               |
| `updated_at`          | Timestamp       | Timestamp of the last modification to the content metadata.              |
| `deleted_at`          | Timestamp       | Soft delete timestamp (NULL if not deleted).                              |

### `KNOWLEDGE_CONTENT_BLOBS`: Content Source Blob Registry
Acts as a content-addressable storage registry for all large file objects related to knwoledge content sources. This table is owned and managed exclusively by the Knowledge Service.

| Field Name | Data Type | Description |
| :--- | :--- | :--- |
| `blob_hash` | CHAR(64) | **Primary Key**. The SHA-256 hash of the S3 object content. |
| `s3_bucket` | VARCHAR(255) | The name of the S3 bucket where the content is stored. |
| `s3_key` | VARCHAR(1024) | The path/key of the object within the S3 bucket. |
| `size_bytes` | BIGINT | The size of the content file in bytes. Useful for analytics. |
| `created_at` | Timestamp | When this content was first ingested and stored. |

---

## Canvas Service

### `NODES`: Node Views (DynamoDB)
- `position_2d` and `position_3d` can be editable by drag and drop on the canvas by users.
- `display_props` can be editable by clicking on the node (not for user-driven actions, but for style and layout info).
- `action_data` is a flexible, data-centric payload to support frontend interactions.
- `engagement_score` is computed by the Activity Service.
- `context_metadata` is user-added context for this space (should be scalable).
- `media_type` enables dynamic UI rendering based on content type (e.g., document icons, audio players, video thumbnails).

| Field Name                  | Data Type | Description                                                                    |
|-----------------------------|-----------|--------------------------------------------------------------------------------|
| `id`                        | UUID      | Primary Key. Unique ID for this node representation.                          |
| `space_id`                  | UUID      | FK to `SPACES.id`. The space this node representation belongs to.              |
| `content_source_id`         | UUID      | FK to `CONTENT_SOURCES.id`. The content source this node representation belongs to. |
| `parent_node_id`            | UUID      | FK to `NODES.id` (self-referential, scoped to `space_id`). Defines hierarchy. |
| `content_entity_type`       | String    | Type of content (content_chunk, chunk_cluster, content_source, source_cluster); determines which standardized actions to use. |
| `media_type`                | String    | Type of media content (document, audio, video, web, text); determines UI rendering and interaction patterns. |
| `depth_level`               | Integer   | Depth in the space-specific hierarchy tree.                                   |
| `position_2d`               | POINT     | Position within this specific space.                                           |
| `position_3d`               | POINTZ    | Position within this specific space.                                           |
| `is_position_locked`        | Boolean   | Whether user has manually locked this position.                               |
| `visibility`                | Boolean   | Whether this node is currently visible in this space's layout.                |
| `display_props`             | JSONB     | Space-specific visual properties (size, opacity, shape, color, etc.).                     |
| `clustering_model_version_id` | UUID    | FK to `ml_service.ML_MODEL_VERSIONS.id`. The model that determined this node's cluster parentage. NULL if manually assigned. |
| `dr_model_version_id`       | UUID    | FK to `ml_service.ML_MODEL_VERSIONS.id`. The model that determined this node's `position_3d`. NULL if manually assigned. |
| `engagement_score`          | Map       | Computed engagement scores from the Activity Service (e.g., `{"canvas_score": 0.85, "chat_score": 0.65, ... "overall_score": ...}`). |
| `action_data`               | JSONB     | Flexible, purpose-driven JSONB object containing data payloads to support frontend interactions. Its structure is determined by the `content_entity_type` and is interpreted by the application layer to enable actions. It contains the data *for* an action, not the action itself. See frontend requirements for examples. |
| `created_at`                | Timestamp | When this node was included in this space.                                     |
| `updated_at`                | Timestamp | Last update to space-specific properties.                                      |

### `EDGES`: Node Edges Entity (DynamoDB)
| Attribute Name | Data Type | Description |
| :--- | :--- | :--- |
| `id` | `String` | **(Partition Key)** A unique UUID for the edge. Essential for linking comments directly to an edge. |
| `start_node_id` | `String` | FK to `NODES.id`. The ID of the node where the edge originates. |
| `end_node_id` | `String` | FK to `NODES.id`. The ID of the node where the edge terminates. This provides directionality. |
| `description` | `String` | Optional short text describing the relationship (e.g., "Related To", "Supports"). |
| `style_metadata` | `Map` | An object containing visual styling properties, similar to Miro or Lucidchart. |
| `created_by` | `String` | ID of the user or system that created the edge. |
| `updated_by` | `String` | ID of the user or system that last updated the edge. |
| `created_at` | `String` | ISO 8601 timestamp of creation. |
| `updated_at` | `String` | ISO 8601 timestamp of the last update. |
| `deleted_at` | `String` | ISO 8601 timestamp of soft deletion (NULL if not deleted). |

#### `style_metadata` Object Example
```json
"style_metadata": {
  "line_type": "dashed",
  "line_weight": 2,
  "color": "#8A2BE2",
  "arrow_head_start": "none",
  "arrow_head_end": "filled_arrow"
}
```

### `COMMENTS`: Node Comments Entity
| Field Name            | Data Type      | Description                                                  |
|-----------------------|----------------|--------------------------------------------------------------|
| `id`                  | UUID           | Primary Key. Unique ID for the comment.                      |
| `node_id`             | UUID           | FK to `NODES.id`. The node this comment is attached to.      |
| `content_comment`     | Text           | The comment text content.                                    |
| `created_by`          | UUID           | FK to User ID. The user who created the comment.             |
| `updated_by`          | UUID           | FK to User ID. The user who last updated the comment.        |
| `created_at`          | Timestamp      | When this comment was created.                               |
| `updated_at`          | Timestamp      | Last update to this comment.                                 |
| `deleted_at`          | Timestamp      | Soft delete timestamp (NULL if not deleted).                  |

---

## Vector Service (Qdrant)

### Collection Strategy
**Multi-tenant approach with shared public content:**
- **`USER_VECTORS_{user_id}`** - Per-user private collections for uploaded/personal content
- **`public_web_vectors`** - Shared collection for public web content accessible to all users

### `USER_VECTORS_{user_id}` Collections (Private Content)
**Purpose**: User-specific vector storage for private content (uploads, personal documents, etc.)

| Field Name | Data Type | Description |
|------------|-----------|-------------|
| `id` | String | Unique identifier for the vector point (matches entity UUID from PostgreSQL) |
| `vector` | Array[Float] | Embedding vector array (typically 1536 dimensions for OpenAI embeddings) |
| `payload` | Object | User-specific metadata object |

**Payload Structure:**
```json
{
  "id": "uuid",
  "vector": [0.123, 0.456, 0.789, ..., 0.999],
  "payload": {
    "content_entity_id": "content_entity_uuid",
    "content_entity_type": "content_source|content_chunk|source_cluster|chunk_cluster",
    "depth_level": 0,
    "title": "My Private Document",
    "keywords": ["personal", "research", "notes"],
    "contents": "Private content that was embedded for vector search...",
    "media_type": "document, audio, video, web, text, etc",
    "source": "upload|gdrive|onedrive|notion|paste",
    "chunk_type": "paragraph, section, table, code_block, list, etc",
    "chunk_index": 5,
    "token_count": 256,
    "parent_id": "parent_entity_uuid",
    "created_at": "2024-05-20T10:00:00Z",
    "updated_at": "2024-05-20T15:30:00Z",
    "created_by": "user_uuid",
    "updated_by": "user_uuid"
  }
}
```

### `WEB_VECTORS` Collection (Shared Web Content)
**Purpose**: Shared vector storage for public web content accessible to all users

| Field Name | Data Type | Description |
|------------|-----------|-------------|
| `id` | String | Unique identifier for the vector point (matches entity UUID from PostgreSQL) |
| `vector` | Array[Float] | Embedding vector array (typically 1536 dimensions for OpenAI embeddings) |
| `payload` | Object | Web content metadata with user access tracking |

**Payload Structure:**
```json
{
  "id": "web_content_entity_uuid",
  "vector": [0.123, 0.456, 0.789, ..., 0.999],
  "payload": {
    "content_entity_id": "web_content_entity_uuid",
    "content_entity_type": "content_source|content_chunk|source_cluster|chunk_cluster",
    "depth_level": 0,
    "title": "Public Web Article Title",
    "keywords": ["public", "web", "article", "shared"],
    "contents": "Public web content that was embedded for vector search...",
    "url": "https://example.com/article",
    "domain": "example.com",
    "content_hash": "sha256_hash_of_content",
    "media_type": "web",
    "chunk_type": "paragraph|section|table|code_block|list",
    "chunk_index": 5,
    "token_count": 256,
    "parent_id": "parent_web_entity_uuid",
    "status": "processed",
    "processing_status": "completed",
    "is_public": true,
    "access_count": 15,
    "created_at": "2024-05-20T10:00:00Z",
    "last_accessed_at": "2024-05-20T16:00:00Z"
  }
}
```


---

## ML Service
This service is responsible for creating, managing, versioning, and serving machine learning models for content organization and visualization (e.g., clustering, dimensionality reduction). It provides a central registry of models and their versions, which can be referenced by other microservices.

### `ML_MODELS`: Model Definitions
This table defines the high-level configuration and scope of a machine learning model. It acts as a container for all its versions.

| Field Name | Data Type | Description |
| :--- | :--- | :--- |
| `id` | UUID | **Primary Key**. Unique identifier for the model definition. |
| `model_name` | String | A human-readable name for the model (e.g., "Default UMAP", "Finance Docs Clustering"). |
| `model_scope` | String | The scope of the model: `GLOBAL`, `USER`, or `SPACE`. |
| `owner_id` | UUID | FK to `USERS.id`. The user who owns this model. NULL if `model_scope` is `GLOBAL`. |
| `space_id` | UUID | FK to `SPACES.id`. The space this model is specific to. NULL unless `model_scope` is `SPACE`. |
| `model_type` | String | The type of task this model performs (e.g., `clustering`, `dimensionality_reduction`). |
| `algorithm` | String | The specific algorithm used (e.g., `HDBSCAN`, `UMAP`, `PCA`). |
| `active_version_id` | UUID | FK to `ML_MODEL_VERSIONS.id`. Points to the version currently active for inference. |
| `creation_params` | JSONB | The initial configuration parameters for the model algorithm. |
| `is_active` | Boolean | Whether this model is actively being updated and used. |
| `created_at` | Timestamp | When the model definition was first created. |
| `updated_at` | Timestamp | When the model definition was last updated. |

### `ML_MODEL_VERSIONS`: Incremental Model Snapshots
This table stores the immutable, versioned snapshots of a trained model. Each record represents a specific state of the model trained on a specific set of data. The model artifact itself is serialized using a library like `joblib` or `pickle` and stored in blob storage. This table provides the pointer to that stored object. **The `id` of this table is the critical foreign key that other services will use.**

| Field Name | Data Type | Description |
| :--- | :--- | :--- |
| `id` | UUID | **Primary Key**. Unique, stable identifier for this specific model version. |
| `model_id` | UUID | FK to `ML_MODELS.id`. The parent model definition this version belongs to. |
| `version_number` | Integer | A monotonically increasing version number scoped to the `model_id`. |
| `status` | String | The lifecycle status of this version (`training`, `available`, `archived`, `failed`). |
| `model_storage_path` | String | A URI pointing to the serialized model object (e.g., a `.pkl` or `.joblib` file) in a cloud blob storage service. |
| `training_metadata` | JSONB | Metadata about the training run: data sources used, vector count, training duration, etc. |
| `performance_metrics` | JSONB | Key performance metrics for this version (e.g., silhouette score, trust/continuity for UMAP). |
| `created_at` | Timestamp | When this version was created. |


## Chat Service
The Chat Service is architected to materialize a user's thought process as an interactive, explorable journey. It moves beyond a simple transcript to create a version-controlled map of inquiry, supporting non-linear exploration, layered information discovery, and proactive knowledge suggestion. For performance and scalability, large text payloads from AI agents and chat messages are offloaded to a content-addressed blob storage system.

### `AI_AGENTS`: AI Agent Definitions
Defines the available AI agents with their roles and capabilities for multi-agent conversations.

| Field Name | Data Type | Description |
| :--- | :--- | :--- |
| `id` | UUID | Primary Key. Unique identifier for the AI agent. |
| `name` | String | Unique name for the agent (e.g., "research_specialist", "summarizer"). |
| `agent_type` | String | Type category of the agent (research, analysis, creative, critique, etc.). |
| `role_description` | Text | Description of the agent's role and capabilities. |
| `system_prompt` | Text | The system prompt that defines the agent's behavior. |
| `model_config` | JSONB | Default model configuration (model, temperature, max_tokens, etc.). |
| `capabilities` | Array of Strings | List of capabilities (document_analysis, code_review, etc.). |
| `is_active` | Boolean | Whether this agent is currently available for use. Default: true. |
| `created_at` | Timestamp | When the agent was created. |
| `updated_at` | Timestamp | When the agent was last modified. |
| `created_by` | UUID | FK to `USERS.id`. User who created/configured this agent. |

### `CHAT_CONTENT_BLOBS`: Chat Service Blob Registry
Acts as a content-addressable storage registry for all large text objects generated within the Chat Service. This table is owned and managed exclusively by the Chat Service.

| Field Name | Data Type | Description |
| :--- | :--- | :--- |
| `blob_hash` | CHAR(64) | **Primary Key**. The SHA-256 hash of the S3 object content. |
| `s3_bucket` | VARCHAR(255) | The name of the S3 bucket where the content is stored. |
| `s3_key` | VARCHAR(1024) | The path/key of the object within the S3 bucket. |
| `size_bytes` | BIGINT | The size of the content file in bytes. Useful for analytics. |
| `created_at` | Timestamp | When this content was first ingested and stored. |

### `CHAT_SESSIONS`: Conversation Threads
Manages distinct conversation threads, acting as the top-level container for a user's exploration within a space.

| Field Name | Data Type | Description |
| :--- | :--- | :--- |
| `id` | UUID | Primary Key. Unique identifier for the chat session. |
| `space_id` | UUID | FK to `SPACES.id`. The space where the chat takes place. |
| `user_id` | UUID | FK to `USERS.id`. The user who initiated the session. |
| `title` | String | A user-editable or auto-generated title for the session. |
| `keywords` | Array of Strings | Extracted keywords from the conversation. NULL until summarized. |
| `content_summary` | Text | AI-generated summary of the conversation. NULL until summarized. |
| `summarized_at` | Timestamp | When the summarization was completed. NULL if not summarized. |
| `created_at` | Timestamp | When the session was created. |
| `updated_at` | Timestamp | When the session was last active. |
| `deleted_at` | Timestamp | Soft delete timestamp. |

### `CHAT_BRANCHES`: Pointers to Conversation Timelines
This table manages branches as named pointers to a specific message in the conversation graph. This allows for a non-duplicative, graph-based structure where branch membership is calculated by traversing backwards from a branch's "tip" message.

**Branching Logic**:
1.  **Branch Definition**: A branch is defined by its `tip_message_id`. Its content consists of the tip message and all its ancestors, found by recursively following `parent_message_id`.
2.  **Creating a Branch**: When a user forks from a source message (`M_source`), a new branch record is created. Its `tip_message_id` and `fork_message_id` are both set to `M_source.id`, and `parent_branch_id` is set to the branch `M_source` was on.
3.  **Adding a Message**: When a new message (`M_new`) is added to a branch, the system simply updates that branch's `tip_message_id` to point to `M_new`.

| Field Name | Data Type | Description |
| :--- | :--- | :--- |
| `id` | UUID | Primary Key. Unique identifier for the branch. |
| `session_id` | UUID | FK to `CHAT_SESSIONS.id`. The session this branch belongs to. |
| `tip_message_id` | UUID | FK to `CHAT_MESSAGES.id`. Points to the most recent message of this branch, defining its current state. |
| `fork_message_id` | UUID | FK to `CHAT_MESSAGES.id`. The specific message where this branch was forked. NULL for the initial `main` branch. |
| `parent_branch_id` | UUID | FK to `CHAT_BRANCHES.id` (self-referential). The branch from which this one was forked. NULL for the `main` branch. |
| `title` | String | A user-editable name for the branch. Defaults to `main`. |
| `status` | String | The current lifecycle state of the branch (e.g., `active`, `merged`, `archived`). |
| `created_by` | UUID | FK to `USERS.id`. The user who created the branch. |
| `created_at` | Timestamp | When the branch was created. |
| `updated_at` | Timestamp | When the branch was last modified. |
| `deleted_at` | Timestamp | Soft delete timestamp (NULL if not deleted). |

### `CHAT_MESSAGES`: Nodes in the Conversation Graph
Each record represents a single, immutable node in the conversation. A message only knows its direct parent, forming a Directed Acyclic Graph (DAG) of the entire session history. It has no concept of branch membership itself.

| Field Name | Data Type | Description |
| :--- | :--- | :--- |
| `id` | UUID | Primary Key. Unique ID for the message. |
| `session_id` | UUID | FK to `CHAT_SESSIONS.id`. The session this message belongs to. |
| `parent_message_id` | UUID | FK to `CHAT_MESSAGES.id` (self-referential). Defines the graph structure. NULL for the very first message in a session. |
| `role` | String | Role of the author (`user` or `assistant`). CHECK constraint applied. |
| `content_blob_hash` | CHAR(64) | FK to `CHAT_CONTENT_BLOBS.blob_hash`. The hash of the S3 object containing the message content. |
| `response_metadata` | JSONB | For AI responses: stores aggregated metadata like generation config, overall performance metrics, and retrieved context. See example below. |
| `model_version` | String | For AI responses: tracks the model version used for generation, ensuring reproducibility. |
| `user_feedback` | String | Simple feedback on AI responses (`good`, `bad`, NULL). NULL for user messages or no feedback given. |
| `is_branch_root` | Boolean | Denormalized flag for UI performance. True if this message is the starting point of one or more branches. Maintained by the application layer. |
| `created_at` | Timestamp | When the message was created. |

**Example `response_metadata` format (Aggregated):**
```json
{
  "generation_config": {
    "model": "gpt-4-turbo",
    "temperature": 0.5,
  },
  "performance_metrics": {
    "total_response_time_ms": 2800,
    "total_token_usage": {
      "prompt_tokens": 4096,
      "completion_tokens": 1024,
      "total_tokens": 5120
    },
    "total_cost_estimate_usd": 0.065
  },
  "retrieved_chunks": [
    {"chunk_id": "uuid-chunk-1", "citation_number": 1, "relevance_score": 0.95},
    {"chunk_id": "uuid-chunk-2", "citation_number": 2, "relevance_score": 0.87}
  ],
  "agents_involved": [
    "research-agent-uuid",
    "summary-agent-uuid"
  ]
}
```

### `AGENT_GENERATIONS`: Agent Thinking & Generation Steps
This table logs every single atomic action taken by any agent in the background. It is designed for detailed analytics, debugging, and tracing the exact thought process of the AI system.

| Field Name | Data Type | Description |
| :--- | :--- | :--- |
| `id` | UUID | Primary Key. Unique ID for this specific generation step. |
| `final_message_id` | UUID | FK to `CHAT_MESSAGES.id`. Groups all steps that contribute to one final response. |
| `agent_id` | UUID | FK to `AI_AGENTS.id`. Tracks which agent performed this step. |
| `generation_type` | String | The type of action (`thinking`, `tool_call`, `retrieval`, `final_contribution`). |
| `content_blob_hash` | CHAR(64) | FK to `CHAT_CONTENT_BLOBS.blob_hash`. The hash of the S3 object containing the generation step's content. |
| `generation_metadata` | JSONB | Step-specific data (e.g., generation config, performance metrics, tool params). See example below. |
| `created_at` | Timestamp | Timestamp of this specific step. |

**Example `generation_metadata` format (Single Step):**
```json
{
  "generation_config": {
    "model": "gpt-4-turbo",
    "temperature": 0.5,
  },
  "performance_metrics": {
    "response_time_ms": 450,
    "token_usage": {
      "prompt_tokens": 512,
      "completion_tokens": 128,
      "total_tokens": 640
    },
    "cost_estimate_usd": 0.008
  },
  "step_specific_data": {
    "tool_name": "database_query",
    "relevance_score": 0.91 
  }
}
```

### `MESSAGE_CONTEXT_LINKS`: Traceable Context Lineage
Links a specific `NODE` (e.g., a document chunk) to the user query and its corresponding final AI response. This creates a traceable lineage of what information was used to answer what question. This table may be leveraged by the Canvas service in the future to visually connect chat interactions with knowledge graph nodes.

| Field Name | Data Type | Description |
| :--- | :--- | :--- |
| `id` | UUID | Primary Key. Unique ID for the context link. |
| `query_message_id` | UUID | FK to `CHAT_MESSAGES.id` (the user's query). |
| `response_message_id` | UUID | FK to `CHAT_MESSAGES.id` (the final AI response). |
| `node_id` | UUID | FK to `NODES.id` (the context source node). |
| `link_metadata` | JSONB | Stores metadata about the link, such as citation number and relevance score. See example below. |
| `created_at` | Timestamp | When this context link was established. |

**Example `link_metadata` format:**
```json
{
  "citation_number": 1,
  "relevance_score": 0.95
}
```

---

## User Service

### `USERS`: User Entity
| Field Name            | Data Type      | Description                                                  |
|-----------------------|----------------|--------------------------------------------------------------|
| `id`                  | UUID           | Unique identifier for the user (matches Clerk user ID).     |
| `clerk_user_id`       | String         | Clerk's unique user identifier for integration.             |
| `email`               | String         | User's email address (synced from Clerk).                   |
| `full_name`           | String         | User's full name (synced from Clerk).                       |
| `username`            | String         | Optional unique username for display.                       |
| `profile_picture_url` | String         | URL of the user's profile picture (synced from Clerk).      |
| `storage_used_bytes`  | BigInt         | Total storage consumed by user across all spaces.           |
| `storage_quota_bytes` | BigInt         | Maximum storage allowed for this user (default: 5GB).       |
| `status`              | String         | Account status (active, suspended, deleted).                |
| `created_at`          | Timestamp      | Timestamp of when the user was first created in our system. |
| `updated_at`          | Timestamp      | Timestamp of the last modification to the user's profile.   |
| `deleted_at`          | Timestamp      | Soft delete timestamp (NULL if not deleted).                |

### `USER_PREFERENCES`: User Settings and Preferences
| Field Name            | Data Type      | Description                                                  |
|-----------------------|----------------|--------------------------------------------------------------|
| `user_id`             | UUID           | Primary Key. Foreign Key referencing the `USERS.id`.       |
| `theme`               | String         | UI theme preference (light, dark, auto).                    |
| `language`            | String         | Preferred language code (e.g., 'en', 'ja').                 |
| `timezone`            | String         | User's timezone (e.g., 'America/New_York').                 |
| `canvas_settings`     | JSONB          | Canvas-specific preferences (grid, snap, etc.).             |
| `notification_settings` | JSONB        | Notification preferences (email, push, in-app).             |
| `accessibility_settings` | JSONB       | Accessibility preferences (font size, contrast, etc.).      |
| `created_at`          | Timestamp      | When preferences were first created.                        |
| `updated_at`          | Timestamp      | Last time preferences were updated.                         |

---

