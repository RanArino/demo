# Software Demo Requirements: Hierarchical Text Visualization

## 1. Introduction

This document outlines the requirements for building a demonstration version of the software. The core concept is to structure text documents hierarchically based on granularity (detail level) and visualize this structure in an interactive 2D/3D space using Next.js App router for the frontend and Go/Python for the backend. This demo aims to showcase the feasibility and user experience of navigating information from a high-level overview down to specific details.

## 2. Goals

*   Demonstrate the core concept of hierarchical text structuring and visualization.
*   Showcase the interactive navigation between different levels of information granularity.
*   Provide a basic search functionality within the visualized structure.
*   Validate the chosen technology stack for core functionalities (visualization, basic backend processing).

## 3. Scope

### 3.1 In Scope

*   **Data Ingestion:** Ability to upload a small number of local text files (e.g., `.txt`, `.md`). The application will also include a default set of text data in a `data/` folder for immediate demonstration.
*   **Basic Processing:**
    *   Text segmentation into sentences with approximately 100-200 tokens per segment.
    *   Generating vector embeddings for document summaries and text chunks using Gemini API.
    *   Generating a basic summary for each uploaded document using Gemini API.
    *   Storing and retrieving embeddings using **Qdrant**.
*   **Hierarchical Structure (Simplified):**
    *   A two-level hierarchy: Document Summaries -> Text Chunks.
    *   Applying dimensionality reduction (e.g., UMAP with n_neighbors=15) to embeddings for 2D/3D positioning.
    *   Applying soft clustering (e.g., Gaussian Mixture Model or HDBSCAN) to group related summaries and chunks visually.
*   **Interactive Visualization:**
    *   Displaying the hierarchical structure in a 2D or 3D force-directed graph using `react-force-graph-3d` within a Next.js (App Router) application.
    *   Nodes representing document summaries and text chunks.
    *   Basic interactions: Zooming, Panning.
    *   Selecting a node (summary or chunk) displays its content (summary text or original chunk text).
    *   Hovering over a node shows basic information (e.g., keywords - *optional stretch goal*).
*   **Basic Search:**
    *   A simple search bar for keyword input in the frontend.
    *   Performing vector similarity search using **Qdrant** based on the query against chunk embeddings (via backend API call).
    *   Highlighting the relevant nodes (chunks and their parent summary) in the visualization.

### 3.2 Out of Scope (for Demo)

*   Cloud storage integration (Google Drive, OneDrive).
*   Cloud-based Vector Databases (e.g., Pinecone, Weaviate, Vertex AI Vector Search) - *Note: Qdrant Cloud is an option, but for the demo, we'll assume a local/containerized instance.*
*   Cloud vendor services integration (AWS, GCP, Azure).
*   Advanced LLM features (fine-tuning, complex summarization).
*   Multi-level deep hierarchies (beyond Summary -> Chunks).
*   Complex inter-layer relationship calculations and display (beyond parent-child).
*   Real-time updates of visualization based on reading progress.
*   Advanced query systems (Natural Language Query understanding beyond simple keywords, query support features).
*   Embedding model fine-tuning or custom model training.
*   Advertising features.
*   User accounts, authentication, collaboration features.
*   Mobile application.
*   Advanced non-functional requirements (high scalability, high availability, robust error handling).
*   Character design / Agent features.

## 4. Functional Requirements

### 4.1 User Stories

*   **UC1: Upload Documents:** As a user, I want to upload one or more local text files (`.txt`, `.md`) so that they can be processed and visualized.
*   **UC2: Process Documents:** As a system, upon file upload, I need to segment the text, generate embeddings for chunks, generate a summary for the document, and calculate positions for visualization.
*   **UC3: View Visualization:** As a user, I want to see the processed documents represented as an interactive 2D/3D graph, with distinct representations for summaries and text chunks.
*   **UC4: Interact with Graph:** As a user, I want to be able to zoom and pan the graph to explore the structure.
*   **UC5: View Content:** As a user, I want to click on a node (summary or chunk) in the graph to view its corresponding text content in a separate panel.
*   **UC6: Search Content:** As a user, I want to enter keywords into a search bar and see the nodes relevant to my search highlighted in the graph.

### 4.2 System Requirements

*   **SR1: File Handling:** The system must accept `.txt` and `.md` file uploads.
*   **SR2: Text Segmentation:** The system must segment document text into smaller chunks (e.g., by paragraph or a defined token limit).
*   **SR3: Embedding Generation:** The system must use a pre-trained sentence transformer or similar model (via Python service) to generate vector embeddings for each text chunk.
*   **SR4: Summarization:** The system must use a pre-trained LLM API (e.g., OpenAI, Gemini, local model via Python service) to generate a concise summary for each uploaded document.
*   **SR5: Dimensionality Reduction:** The system must apply a dimensionality reduction algorithm (e.g., UMAP) to the embeddings to generate 2D or 3D coordinates for visualization.
*   **SR6: Clustering:** The system must apply a clustering algorithm (e.g., K-Means) to group related nodes based on their embeddings. Cluster information should be usable for visual distinction (e.g., color).
*   **SR7: Visualization Rendering:** The system must render the summaries and chunks as nodes in a force-directed graph using `react-force-graph-3d`. Edges should potentially link chunks to their parent summary.
*   **SR8: Node Selection:** The system must detect user clicks on nodes and display the associated text content.
*   **SR9: Vector Search:** The system must embed the user's search query using the same embedding model and perform a cosine similarity search against the text chunk embeddings using **Qdrant**.
*   **SR10: Result Highlighting:** The system must visually highlight the top 10 nodes corresponding to the search results in the graph.

## 5. Non-Functional Requirements (Simplified for Demo)

*   **NFR1: Performance:** Processing and visualization generation for a small set of documents (e.g., <10 documents, < 50 pages total) should complete within a reasonable time frame (e.g., < 60 seconds). Search should be near real-time (< 3 seconds).
*   **NFR2: Usability:** Basic graph interactions (zoom, pan, click) should be intuitive. Text display should be clear.
*   **NFR3: Technology Stack:**
    *   **Frontend:** Next.js (App Router), React, `react-force-graph-3d`.
    *   **Backend (General):** Go (using standard libraries, potentially Gin/Echo framework). Responsible for API handling, file processing orchestration, and direct communication with Gemini API and Qdrant.
    *   **Backend (ML/DS):** Python (using Flask/FastAPI). Only used for operations that cannot be efficiently implemented in Go, such as complex dimensionality reduction and clustering algorithms.
    *   **Vector Store:** **Qdrant** instance (can be run locally via Docker or potentially use a free cloud tier for demo purposes).
    *   **Inter-service Communication:** REST API between Go backend and Python ML service.
*   **NFR4: Deployment:** The entire application (Frontend, Go Backend, Python Service, **Qdrant**) must be containerized using Docker and launchable via Docker Compose for local development and demonstration. Code structure should facilitate future deployment to cloud container platforms, avoiding hardcoded local paths where possible (use environment variables).

## 6. Data

*   Input: `.txt` or `.md` files containing plain text, either uploaded by the user or read from a default `data/` directory in the frontend. Uploaded files are not persisted between sessions.
*   Intermediate: Vector embeddings (stored in **Qdrant**), 2D/3D coordinates, cluster assignments, document summaries, text chunks.
*   Output: Interactive visualization, displayed text content.

## 7. Assumptions

*   Pre-trained embedding models and LLMs provide sufficient quality for the demo.
*   The number and size of documents used for the demo will be limited.
*   Basic text segmentation (e.g., by paragraph) is adequate.
*   Standard dimensionality reduction and clustering algorithms will produce meaningful visual structures.
*   **Qdrant** provides sufficient performance for vector storage and search for the demo scope.
*   A containerized **Qdrant** instance is adequate for vector storage for the demo.
*   A local development environment using Docker Compose is sufficient; cloud deployment setup is not required for the demo itself, but the architecture should allow for it.
