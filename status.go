package migrate

import (
	"context"
	"log"
	"strconv"

	"github.com/johngibb/migrate/db"
	"github.com/johngibb/migrate/source"
)

// Status displays every migration, and whether it's been applied yet.
func Status(ctx context.Context, src *source.Source, db *db.Client) error {
	migrations, err := src.FindMigrations()
	if err != nil {
		return err
	}
	applied, err := db.GetMigrations(ctx)
	if err != nil {
		return err
	}
	isApplied := func(name string) bool {
		for _, a := range applied {
			if a.Name == name {
				return true
			}
		}
		return false
	}

	w := maxNameWidth(migrations)
	for _, m := range migrations {
		status := "pending"
		if isApplied(m.Name) {
			status = "applied"
		}

		log.Printf("%-"+strconv.Itoa(w)+"s %s\n", m.Name, status)
	}
	return nil
}

func maxNameWidth(mm []*source.Migration) int {
	w := 0
	for _, m := range mm {
		if n := len(m.Name); n > w {
			w = n
		}
	}
	return w
}
