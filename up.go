package migrate

import (
	"log"

	"github.com/johngibb/migrate/db"
	"github.com/johngibb/migrate/source"
	"github.com/pkg/errors"
)

// Up applies all pending migrations from src to the db.
func Up(src *source.Source, db *db.Client) error {
	migrations, err := src.FindMigrations()
	if err != nil {
		return errors.Wrap(err, "error reading migration files")
	}
	applied, err := db.GetMigrations()
	if err != nil {
		return errors.Wrap(err, "error fetching migrations")
	}
	isApplied := func(name string) bool {
		for _, a := range applied {
			if a.Name == name {
				return true
			}
		}
		return false
	}

	var pending []*source.Migration
	for _, m := range migrations {
		if !isApplied(m.Name) {
			pending = append(pending, m)
		}
	}

	if len(pending) == 0 {
		log.Println("nothing to do")
		return nil
	}

	for _, m := range pending {
		log.Printf("Running %s\n", m.Name)
		stmts, err := m.ReadStatements()
		if err != nil {
			return errors.Wrapf(err, "error reading migration: %v", m)
		}
		if err := db.ApplyMigration(m.Name, stmts); err != nil {
			return errors.Wrapf(err, "error applying migration: %v", m)
		}
	}
	return nil
}
