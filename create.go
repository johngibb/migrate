package migrate

import (
	"log"

	"github.com/johngibb/migrate/source"
)

// Create generates an empty migration file.
func Create(src *source.Source, name string) error {
	path, err := src.Create(name)
	if err != nil {
		return err
	}
	log.Printf("Created %s\n", path)
	return nil
}
