package settings

import (
	"os"
	"path/filepath"
	"sync"
)

var settingsMu sync.Mutex

func writeAtomically(path string, write func(*os.File) error) (err error) {
	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()

	if err = write(f); err != nil {
		_ = f.Close()
		return err
	}

	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}

	if err = f.Close(); err != nil {
		return err
	}

	if err = os.Rename(tmpPath, path); err != nil {
		return err
	}

	return nil
}
