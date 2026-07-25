package chunk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveChunkSize(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		requested int
		want      int
	}{
		{
			name:      "Use requested chunk size when provided",
			size:      1,
			requested: 1024,
			want:      1024,
		},
		{
			name:      "Return auto chunk size above threshold",
			size:      parallelTriggerThresholdBytes + 1,
			requested: 0,
			want:      defaultAutoChunkSizeBytes,
		},
		{
			name:      "Return zero at threshold boundary",
			size:      parallelTriggerThresholdBytes,
			requested: 0,
			want:      0,
		},
		{
			name:      "Return zero below threshold",
			size:      128,
			requested: 0,
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := createSizedTempFile(t, tt.size)
			defer func() {
				_ = file.Close()
			}()

			got := ResolveChunkSize(file, tt.requested)
			if got != tt.want {
				t.Errorf("ResolveChunkSize() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPrepareChunkSize_SeeksToStartAndResolves(t *testing.T) {
	file := createSizedTempFile(t, 128)
	defer func() {
		_ = file.Close()
	}()

	if _, err := file.Seek(64, 0); err != nil {
		t.Fatalf("failed to seek file: %v", err)
	}

	got, err := PrepareChunkSize(file, 0)
	if err != nil {
		t.Fatalf("PrepareChunkSize() error = %v", err)
	}
	if got != 0 {
		t.Errorf("PrepareChunkSize() = %d, want %d", got, 0)
	}

	offset, err := file.Seek(0, 1)
	if err != nil {
		t.Fatalf("failed to read current offset: %v", err)
	}
	if offset != 0 {
		t.Errorf("file offset = %d, want 0", offset)
	}
}

func createSizedTempFile(t *testing.T, size int) *os.File {
	t.Helper()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "chunking.tmp")
	data := make([]byte, size)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open temp file: %v", err)
	}

	return file
}
