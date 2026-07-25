package cmd

import "os"

const (
	parallelTriggerThresholdBytes = 4 * 1024 * 1024
	defaultAutoChunkSizeBytes     = 64 * 1024
)

func resolveChunkSize(file *os.File, requested int) int {
	if requested > 0 {
		return requested
	}

	info, err := file.Stat()
	if err != nil {
		return 0
	}

	if info.Size() > parallelTriggerThresholdBytes {
		return defaultAutoChunkSizeBytes
	}

	return 0
}

func prepareChunkSize(file *os.File, requested int) (int, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return 0, err
	}

	return resolveChunkSize(file, requested), nil
}
