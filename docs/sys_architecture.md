# System Architecture: Hierarchical Text Visualization Demo

This document outlines the system architecture for the Hierarchical Text Visualization demo, based on the defined requirements.

## 1. Overview

The system consists of three main components:
1.  **Frontend:** A Next.js application responsible for user interaction, data visualization, and communication with the backend.
2.  **Backend (Go):** A Go service acting as the main API gateway, orchestrating file processing, managing metadata, interacting with the vector store (Qdrant), and communicating with the Python ML service.
3.  **ML Service (Python):** A Python service (using gRPC or Flask/FastAPI) responsible for computationally intensive ML tasks: text embedding, summarization, dimensionality reduction, and clustering.

```mermaid
graph LR
    subgraph Browser
        Frontend[Next.js Frontend]
    end

    subgraph "Server / Docker Compose"
        Kong[Kong API Gateway]
        Backend[Go Backend API]
        MLService[Python ML Service]
        VectorDB[(Qdrant Vector Database)]
        Storage[(Persistent Storage / Volume)]
    end

    subgraph "External Services"
        Gemini[Gemini API]
    end

    subgraph "Conceptual / Libraries"
        style MLAlgorithms fill:#ff4aff,stroke:#333,stroke-width:2px
        MLAlgorithms["ML Algorithms<br/>(Reduction, Clustering)"]
    end

    Frontend -- REST API --> Kong
    Kong -- Routes Requests --> Backend

    Backend -- REST API or gRPC --> MLService
    Backend -- Qdrant Go Client --> VectorDB
    Backend -- Manages/Reads --> Storage
    Backend -- Direct API Call --> Gemini
    MLService -- Uses --> MLAlgorithms

    style Kong fill:#158415,stroke:#333,stroke-width:2px
    style VectorDB fill:#7a7afa,stroke:#333,stroke-width:2px
    style Storage fill:#7a7afa,stroke:#333,stroke-width:2px
    style Gemini fill:#ff4aff,stroke:#333,stroke-width:2px
```

## 2. Data Storage
With the introduction of Qdrant and Gemini API, the data storage strategy is updated as follows:
1. **Uploaded/Default Text Files:** 
   - Original .txt and .md files will be stored in the frontend's `data/` directory for default files.
   - Uploaded files are temporarily stored in memory and not persisted between sessions.
   
2. **Vector Embeddings & Search:** 
   - All vector embeddings (text chunks, potentially summaries) will be stored, indexed, and searched using a dedicated Qdrant service instance. 
   - The Go backend will interact with Qdrant via its official Go client.
   - Vector embeddings will be generated using Gemini API, called directly from the Go backend.
   
3. **Metadata:** 
   - Information linking documents, chunks, summaries, coordinates, and cluster assignments will be stored within the payload of the vectors in Qdrant. 
   - Document summaries and overall structure will be stored in Qdrant with appropriate metadata.


**Example Metadata Structure (Conceptual):**

**(A) Document-Level Information (Potentially stored in a separate JSON file or Qdrant collection):**

```
// Example: Stored in documents.json or a dedicated Qdrant collection
[
  {
    "documentId": "doc_1", // Unique ID for the document
    "fileName": "example1.txt",
    "summaryText": "This is a summary of example1.",
    // Summary embedding could be stored here if needed, or as a point in Qdrant
    "summaryPosition": [x_doc, y_doc, z_doc], // Position for the summary node in visualization
    "summaryClusterId": "cluster_A", // Cluster assignment for the summary node
    "chunkIds": ["chunk_1_1", "chunk_1_2", ...] // List of associated chunk IDs
  }
  // ... more documents
]
```

**(B) Chunk-Level Information (Stored as points in Qdrant):**

**(Conceptual Qdrant Point Structure)**

```
// Stored within Qdrant chunk collection
{
  "id": "vector_id_chunk_1_1", // Qdrant point ID (e.g., maps to chunkId or is chunkId)
  "vector": [0.3, 0.4, ...],   // Embedding of the chunk text
  "payload": {
    "documentId": "doc_1",      // Link back to the parent document
    "chunkId": "chunk_1_1",     // Internal ID for the chunk
    "text": "This is the first paragraph.",
    "position": [x1, y1, z1],   // Position for the chunk node (calculated by ML service)
    "clusterId": "cluster_A"   // Cluster assignment for the chunk node
  }
}
```
*(Note: This shows separate handling. Document metadata (A) provides overall info and links to chunks. Chunk data (B) is stored in Qdrant with embeddings and includes a link back to the parent document. How document data is persisted (JSON vs. separate Qdrant collection) is an implementation detail.)*

## 3. Folder Structures

### 3.1 Frontend (Next.js - App Router)
```text
frontend/
├── src/                  # Main source code directory
│   ├── app/              # Next.js App Router directory
│   │   ├── page.tsx        # Homepage (/) component (Displays space overview/gallery)
│   │   ├── layout.tsx    # Root layout
│   │   ├── globals.css   # Global styles
│   │   ├── spaces/       # # Routes related to viewing specific spaces
│   │   │   └── [space_id]/ # # Dynamic route using space ID (e.g., /spaces/xyz789)
│   │   │       └── page.tsx  # # Page component to display the visualization for a specific space
│   │   └── ...  # new pages come in     
│   ├── components/       # Reusable React components
│   │   ├── ui/           # General UI elements (buttons, inputs etc. - shadcn/ui potentially)
│   │   ├── visualization/ # Visualization specific components
│   │   │   ├── Canvas3D.tsx # Main 3D visualization component (handling layers, nodes)
│   │   │   ├── NodeDetailsPanel.tsx # Panel to show node content/document preview
│   │   │   ├── SearchBar.tsx      # Search input component
│   │   │   ├── FlowView.tsx       # Git-style chat flow visualization
│   │   │   ├── MiniMap.tsx        # Mini-map for Tree/Canvas view (Homepage)
│   │   │   └── ...
│   │   ├── home/           # Components specific to the homepage/spaces overview
│   │   │   ├── SpaceCard.tsx      # Card component for gallery/tree view
│   │   │   ├── SpaceGalleryView.tsx
│   │   │   ├── SpaceListView.tsx
│   │   │   ├── SpaceTreeView.tsx  # Carousel/Tree view component
│   │   │   └── ...
│   │   ├── shared/         # Components shared across different pages/features
│   │   │   ├── DocumentUploadModal.tsx # Modal for uploading/adding documents
│   │   │   ├── WebSearchModal.tsx      # Modal for web search functionality
│   │   │   ├── ChatInterface.tsx       # Chat input/output component
│   │   │   └── ...
│   │   └── SettingsButton.tsx # Reusable settings access component
│   ├── hooks/            # Custom React hooks (e.g., useGraphData, useApi, useChat)
│   ├── lib/              # Utility functions, API client setup
│   │   ├── api/          # # Centralized Go API client wrappers
│   │   │   ├── fetcher.ts    # # Generic secure fetcher (handles credentials, tokens)
│   │   │   ├── spaceClient.ts# # Client for space-related Go API endpoints
│   │   │   ├── chatClient.ts # # Client for chat-related Go API endpoints
│   │   │   └── ...         # # More clients as needed
│   │   └── utils.ts      # General utility functions
│   └── styles/           # Styling files (if not using CSS-in-JS or Tailwind extensively)
├── public/               # Static assets (images, etc.)
├── data/                 # Default sample text data for demo (Consider moving to backend or dedicated storage)
├── .env.local            # # Contains NEXT_PUBLIC_API_URL for Go backend endpoint
├── next.config.ts        # Next.js configuration
├── tsconfig.json         # TypeScript configuration
├── package.json          # Project dependencies
├── package-lock.json     # Lockfile
├── postcss.config.mjs    # PostCSS configuration
├── eslint.config.mjs     # ESLint configuration
├── .gitignore            # Git ignore file
├── README.md             # Project README
└── Dockerfile            # Dockerfile for building the frontend image
```

### 3.2 Backend (Go)
```text
backend-go/
├── cmd/
│   └── server/           # Main application entry point
│       └── main.go
├── internal/             # Internal application code (not importable by others)
│   ├── api/              # API handlers (e.g., Gin, Echo)
│   │   ├── handlers/     # Request handler implementations (Consider splitting further if complex)
│   │   │   │── document_handler.go
│   │   │   │── search_handler.go
│   │   │   └── ...
│   │   ├── middleware/          # API middleware (logging, auth, CORS, etc.)
│   │   ├── internal_router.go   # Defines all internal API routes
│   │   ├── gateway_router/
│   │   │   ├── router.go  # Initializes all versions
│   │   │   ├── v1.go
│   │   │   ├── v2.go
│   │   │   └── ...
│   │   └── constants/      # API related constants
│   │       └── version.go  # Defines API version constants (e.g., APIVersionV2)
│   ├── config/           # Configuration loading (env vars, files)
│   │   └── config.go
│   ├── core/             # Core business logic & domain types
│   │   ├── models/       # Domain models (document, chunk, summary, space, etc.)
│   │   │   └── document.go
│   │   │   └── ...
│   │   ├── service/      # Business logic services (orchestration)
│   │   │   └── processing_service.go # Orchestrates embedding, ML calls, storage
│   │   │   └── search_service.go    # Handles search logic
│   │   │   └── ...
│   │   └── ports/        # Interfaces defining contracts for external systems
│   │       ├── ml_service.go      # Interface for ML service client
│   │       ├── storage_service.go # Interface for vector storage
│   │       └── llm_service.go     # Interface for Gemini/LLM client
│   ├── adapters/         # Implementations for external services (fits Dependency Inversion)
│   │   ├── geminiclient/ # Gemini API client implementation
│   │   │   └── client.go
│   │   ├── mlclient/     # Client for communicating with Python ML Service
│   │   │   └── client.go
│   │   ├── qdrant/       # Client logic for interacting with Qdrant
│   │   │   └── client.go
│   │   └── ...
│   └── utils/            # General utility functions specific to backend
│       └── utils.go
├── gateway/              # Configuration for the API Gateway by Kong
│   ├── v1/               # Version 1 specific gateway configuration
│   │   └── gateway-config.yaml
│   ├── v2/               # Version 2 specific gateway configuration
│   │   └── gateway-config.yaml
│   └── ...
├── data/                 # Persistent storage mount point (e.g., for uploaded files) - Managed by Docker Volume
├── scripts/              # Helper scripts (build, run, etc.)
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
└── Dockerfile            # Dockerfile for building the Go backend image
```

### 3.3 ML Service (Python - gRPC or Flask/FastAPI)
```
// if gRPC
backend-py/
├── app/ # Main application source code
│ ├── services/ # gRPC service implementations
│ │ ├── reduction_service.py # Dimensionality reduction (UMAP)
│ │ └── clustering_service.py # Clustering (GMM or HDBSCAN)
│ ├── generated/ # Generated Python code from .proto files
│ ├── utils/ # Utility functions
│ └── server.py # gRPC server initialization and entry point
├── proto/ # Protocol Buffer definitions (.proto files)
├── data/ # Potential location for downloaded models (or use cache dir)
├── requirements.txt # Python dependencies (including grpcio, grpcio-tools)
└── Dockerfile # Dockerfile for building the Python ML service image
```

```
// if Flask/FastAPI
backend-py/
├── app/                  # Main application source code
│   ├── api/              # API endpoint definitions
│   │   └── routes.py
│   ├── services/         # Core ML service logic
│   │   ├── reduction_service.py   # Dimensionality reduction (UMAP with n_neighbors=15)
│   │   └── clustering_service.py  # Clustering (GMM or HDBSCAN)
│   ├── utils/            # Utility functions
│   └── main.py           # FastAPI/Flask app initialization and entry point
├── data/                 # Potential location for downloaded models (or use cache dir)
├── requirements.txt      # Python dependencies
└── Dockerfile            # Dockerfile for building the Python ML service image
```

## 4. Inter-service Communication

*   **Frontend -> Go Backend:** REST API (HTTP requests). The frontend sends requests for data, initiates processing, and performs searches.
*   **Go Backend -> Python ML Service:** REST API (HTTP requests) for dimensionality reduction and clustering operations only.
*   **Go Backend -> Gemini API:** Direct API calls for embedding generation and document summarization.
*   **Go Backend -> Qdrant:** Direct interaction using Qdrant's Go client for vector storage and search.

## 5. Containerization

*   A `docker-compose.yml` file at the root of the project will define the four services (frontend, backend-go, backend-py, qdrant) and manage networking.
*  Docker volumes will be used to persist data for Qdrant and potentially the uploaded files directory used by the Go backend.
