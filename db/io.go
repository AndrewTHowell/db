package db

import (
	"fmt"
	"math/rand"
	"os"
)

func Save(path string, data []byte) error {
	tmpFilePath := fmt.Sprintf("%s.tmp.%04d", path, rand.Intn(10000))
	tmpFile, err := os.OpenFile(tmpFilePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("opening tmp file: %w", err)
	}
	defer func() {
		// Close the file. This shouldn't error as it only errors if already closed.
		tmpFile.Close()

		if err != nil {
			os.Remove(tmpFilePath)
		}
	}()

	if _, err = tmpFile.Write(data); err != nil {
		return fmt.Errorf("writing tmp file: %w", err)
	}
	if err = tmpFile.Sync(); err != nil {
		return fmt.Errorf("syncing tmp file: %w", err)
	}
	// Save error to err var to ensure defer sees it.
	err = os.Rename(tmpFilePath, path)
	return err
}
