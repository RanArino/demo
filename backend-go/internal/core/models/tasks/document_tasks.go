package tasks

// DocumentParseTaskPayload represents the data needed for parsing a document
type DocumentParseTaskPayload struct {
	DocumentID    string `json:"document_id"`
	S3Path        string `json:"s3_path"` // Path to the raw document in S3
	SourceSpaceID string `json:"source_space_id,omitempty"`
	FileName      string `json:"file_name"`
	ContentType   string `json:"content_type"`
}

// DocumentIndexTaskPayload represents the data needed for indexing a parsed document
type DocumentIndexTaskPayload struct {
	DocumentID          string `json:"document_id"`
	ParsedContentS3Path string `json:"parsed_content_s3_path"` // Path to the parsed content in S3
	SourceSpaceID       string `json:"source_space_id,omitempty"`
}

// DocumentProcessTaskPayload represents the data needed for complete document processing
type DocumentProcessTaskPayload struct {
	DocumentID    string   `json:"document_id"`
	S3Path        string   `json:"s3_path"`
	SourceSpaceID string   `json:"source_space_id,omitempty"`
	FileName      string   `json:"file_name"`
	ContentType   string   `json:"content_type"`
	ProcessSteps  []string `json:"process_steps"` // e.g., ["parse", "index", "vectorize"]
}

// Task type constants
const (
	DocumentParseTaskType   = "DocumentParseTask"
	DocumentIndexTaskType   = "DocumentIndexTask"
	DocumentProcessTaskType = "DocumentProcessTask"
)
