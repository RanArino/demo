package python

import (
	"context"
	privpb "demo/ms_canvas/go_app/api/proto/private/v1"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPythonServiceIntegration tests real communication with the Python service
// This test requires the Python service to be available and running
func TestPythonServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if we should run integration tests
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=1 to run")
	}

	// Start Python service or connect to existing one
	pythonProcess, cleanup := startPythonService(t)
	defer cleanup()

	// Wait for service to be ready
	waitForPythonService(t, 30*time.Second)

	// Create gateway factory and get gateway
	config := Config{
		Host:        "localhost",
		Port:        50054,
		MaxMsgBytes: 64 * 1024 * 1024,
	}

	factory := &GatewayFactory{
		gateways: make(map[string]*Gateway),
	}

	gateway, err := factory.newGateway(config)
	require.NoError(t, err, "Failed to create gateway")
	defer gateway.Close()

	ctx := context.Background()

	// Test all gateway methods with real Python service
	t.Run("Healthz", func(t *testing.T) {
		testHealthzIntegration(t, ctx, gateway)
	})

	t.Run("EmbedQuery", func(t *testing.T) {
		testEmbedQueryIntegration(t, ctx, gateway)
	})

	t.Run("ChunkEmbed", func(t *testing.T) {
		testChunkEmbedIntegration(t, ctx, gateway)
	})

	// Cleanup Python process
	if pythonProcess != nil {
		pythonProcess.Process.Signal(syscall.SIGTERM)
		pythonProcess.Wait()
	}
}

func testHealthzIntegration(t *testing.T, ctx context.Context, gateway *Gateway) {
	response, err := gateway.Healthz(ctx)
	require.NoError(t, err, "Healthz should not return an error")
	require.NotNil(t, response, "Healthz response should not be nil")

	assert.Equal(t, "OK", response.Status, "Health status should be OK")
	assert.Contains(t, response.Components, "chunking", "Should have chunking component")
	assert.Contains(t, response.Components, "embedding", "Should have embedding component")
	assert.Equal(t, "healthy", response.Components["chunking"], "Chunking should be healthy")
	assert.Equal(t, "healthy", response.Components["embedding"], "Embedding should be healthy")
}

func testEmbedQueryIntegration(t *testing.T, ctx context.Context, gateway *Gateway) {
	text := "This is a test query for embedding"
	config := &privpb.EmbeddingConfig{
		Provider:     "huggingface",
		ModelId:      "all-MiniLM-L6-v2",
		ModelVersion: "",
	}

	response, err := gateway.EmbedQuery(ctx, text, config)
	require.NoError(t, err, "EmbedQuery should not return an error")
	require.NotNil(t, response, "EmbedQuery response should not be nil")

	assert.Greater(t, len(response.Vector), 0, "Vector should not be empty")
	assert.Greater(t, response.Dims, int32(0), "Dims should be positive")
	assert.Equal(t, "all-MiniLM-L6-v2", response.ModelId, "Model ID should match")
	assert.Equal(t, len(response.Vector), int(response.Dims), "Vector length should match dims")

	// Test that we get consistent results for the same input
	response2, err := gateway.EmbedQuery(ctx, text, config)
	require.NoError(t, err, "Second EmbedQuery should not return an error")
	assert.Equal(t, len(response.Vector), len(response2.Vector), "Vector lengths should be consistent")
}

func testChunkEmbedIntegration(t *testing.T, ctx context.Context, gateway *Gateway) {
	testText := `This is the first sentence for testing chunking. This is the second sentence that should be in a different chunk. Here is a third sentence to test the chunking algorithm. And finally, this is the fourth sentence to ensure we have enough content for multiple chunks.`

	req := &privpb.ChunkEmbedRequest{
		SpaceId:         "test-space-123",
		ContentSourceId: "test-content-456",
		ContentNodeId:   "test-node-789",
		Source: &privpb.ChunkEmbedRequest_Text{
			Text: testText,
		},
		Chunking: &privpb.ChunkingConfig{
			Type:           "sentence",
			TargetTokens:   50, // Small target to ensure multiple chunks
			OverlapPercent: 10,
			Tokenizer:      "tiktoken:cl100k_base",
		},
		Embedding: &privpb.EmbeddingConfig{
			Provider:     "huggingface",
			ModelId:      "all-MiniLM-L6-v2",
			ModelVersion: "",
		},
		BatchSize: 10,
	}

	response, err := gateway.ChunkEmbed(ctx, req)
	require.NoError(t, err, "ChunkEmbed should not return an error")
	require.NotNil(t, response, "ChunkEmbed response should not be nil")

	// Validate response structure
	assert.Greater(t, len(response.Results), 0, "Should have at least one chunk result")
	assert.Greater(t, response.Dims, int32(0), "Dims should be positive")
	assert.Equal(t, "all-MiniLM-L6-v2", response.ModelId, "Model ID should match")

	// Validate each chunk result
	for i, result := range response.Results {
		t.Run(fmt.Sprintf("Chunk_%d", i), func(t *testing.T) {
			require.NotNil(t, result.Chunk, "Chunk should not be nil")
			require.NotNil(t, result.Vector, "Vector should not be nil")

			// Validate chunk metadata
			assert.Equal(t, int32(i), result.Chunk.SequenceIndex, "Sequence index should match position")
			assert.Greater(t, len(result.Chunk.Content), 0, "Chunk content should not be empty")
			assert.GreaterOrEqual(t, result.Chunk.StartPosition, int64(0), "Start position should be non-negative")
			assert.Greater(t, result.Chunk.EndPosition, result.Chunk.StartPosition, "End position should be greater than start")

			// Validate embedding vector
			assert.Equal(t, len(result.Vector), int(response.Dims), "Vector length should match dims")
			assert.Greater(t, len(result.Vector), 0, "Vector should not be empty")

			// Check that vector contains actual values (not all zeros)
			hasNonZero := false
			for _, val := range result.Vector {
				if val != 0.0 {
					hasNonZero = true
					break
				}
			}
			assert.True(t, hasNonZero, "Vector should contain non-zero values")
		})
	}

	// Validate that chunks cover the original text
	if len(response.Results) > 1 {
		// Check that chunks are in order
		for i := 1; i < len(response.Results); i++ {
			prevChunk := response.Results[i-1].Chunk
			currChunk := response.Results[i].Chunk
			assert.LessOrEqual(t, prevChunk.EndPosition, currChunk.StartPosition+10, // Allow small overlap
				"Chunks should be roughly sequential")
		}
	}
}

// startPythonService starts the Python service for testing
func startPythonService(t *testing.T) (*exec.Cmd, func()) {
	// Check if Python service is already running
	if isPythonServiceRunning() {
		t.Log("Python service already running, using existing instance")
		return nil, func() {} // No cleanup needed
	}

	// Find Python app directory
	pythonAppDir := findPythonAppDir(t)
	if pythonAppDir == "" {
		t.Skip("Python app directory not found, skipping integration test")
	}

	t.Logf("Starting Python service from: %s", pythonAppDir)

	// Start Python service
	cmd := exec.Command("python", "-m", "app.server")
	cmd.Dir = pythonAppDir
	cmd.Env = append(os.Environ(),
		"CANVAS_PY_HOST=localhost",
		"CANVAS_PY_PORT=50054",
		"CANVAS_CHUNK_TARGET_TOKENS=300",
		"CANVAS_CHUNK_OVERLAP_PERCENT=10",
		"CANVAS_TOKENIZER=tiktoken:cl100k_base",
		"CANVAS_BATCH_SIZE=50",
	)

	// Capture output for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	require.NoError(t, err, "Failed to start Python service")

	cleanup := func() {
		if cmd.Process != nil {
			cmd.Process.Signal(syscall.SIGTERM)
			cmd.Wait()
		}
	}

	return cmd, cleanup
}

// findPythonAppDir finds the Python app directory relative to the test
func findPythonAppDir(t *testing.T) string {
	// Start from current directory and walk up to find python_app
	currentDir, err := os.Getwd()
	require.NoError(t, err, "Failed to get current directory")

	// Walk up directories looking for python_app
	for dir := currentDir; dir != "/" && dir != "."; dir = filepath.Dir(dir) {
		pythonAppPath := filepath.Join(dir, "..", "..", "python_app")
		if _, err := os.Stat(pythonAppPath); err == nil {
			absPath, err := filepath.Abs(pythonAppPath)
			require.NoError(t, err, "Failed to get absolute path")
			return absPath
		}
	}

	return ""
}

// isPythonServiceRunning checks if the Python service is already running
func isPythonServiceRunning() bool {
	config := Config{
		Host:        "localhost",
		Port:        50054,
		MaxMsgBytes: 64 * 1024 * 1024,
	}

	factory := &GatewayFactory{
		gateways: make(map[string]*Gateway),
	}

	gateway, err := factory.newGateway(config)
	if err != nil {
		return false
	}
	defer gateway.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = gateway.Healthz(ctx)
	return err == nil
}

// waitForPythonService waits for the Python service to be ready
func waitForPythonService(t *testing.T, timeout time.Duration) {
	t.Log("Waiting for Python service to be ready...")

	config := Config{
		Host:        "localhost",
		Port:        50054,
		MaxMsgBytes: 64 * 1024 * 1024,
	}

	start := time.Now()
	for time.Since(start) < timeout {
		factory := &GatewayFactory{
			gateways: make(map[string]*Gateway),
		}

		gateway, err := factory.newGateway(config)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err = gateway.Healthz(ctx)
		cancel()
		gateway.Close()

		if err == nil {
			t.Log("Python service is ready!")
			return
		}

		time.Sleep(500 * time.Millisecond)
	}

	t.Fatalf("Python service did not become ready within %v", timeout)
}
