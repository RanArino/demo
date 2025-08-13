### **Document Process Service: The Specialized Worker**

This service is a **stateless, idempotent, and isolated worker**. Its sole purpose is to execute transformation tasks on raw file data. It has no knowledge of users, spaces, or business logic.

#### Core Responsibilities:

1.  **File Transformation:**
    * Accepts a job containing a reference to a file in object storage.
    * Converts various source formats (`.pdf`, `.docx`, `.html`, etc.) into a standardized, clean **Markdown** format.
    * Places the resulting Markdown file back into a designated location in object storage.

2.  **Content & Data Extraction:**
    * Extracts raw text from documents to be used by other services (e.g., search indexing).
    * Identifies, extracts, and uploads embedded media (like images) from within documents to object storage.
    * Parses and extracts structured data (e.g., tables) into formats like JSON or CSV.

3.  **Anti-Corruption Layer for External APIs:**
    * Acts as the sole client for any third-party processing services (e.g., a cloud OCR or AI service).
    * Manages all API keys, SDKs, and error handling logic specific to that external provider, abstracting it away from the rest of our system.
    * Notifies the Knowledge Service of job completion or failure.
