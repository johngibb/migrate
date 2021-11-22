package migrate

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Create generates an empty migration file.
func Create(path, name string) error {
	timestamp := time.Now().UTC().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s.sql", timestamp, name)
	path = filepath.Join(path, filename)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	log.Printf("Created %s\n", path)

	return nil
}
