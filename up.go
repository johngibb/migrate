package migrate

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/johngibb/migrate/db"
	"github.com/johngibb/migrate/source"
	"github.com/pkg/errors"
)

var DefaultLogger = log.New(os.Stderr, "", 0)

// Up applies all pending migrations from src to the db.
func Up(src *source.Source, db *db.Client, quiet bool) (err error) {
	logger := DefaultLogger

	// If we're running in quiet mode, buffer all log messages, and only print
	// them if an error occurs.
	if quiet {
		var buf strings.Builder
		logger = log.New(&buf, "", 0)
		defer func() {
			if quiet && err != nil {
				DefaultLogger.Print(buf.String())
			}
		}()
	}

	migrations, err := src.FindMigrations()
	if err != nil {
		return errors.Wrap(err, "error reading migration files")
	}

	// Acquire an exclusive lock.
	locked, err := db.TryLock()
	if err != nil {
		return errors.Wrap(err, "error acquiring lock")
	}
	if !locked {
		return errors.New("could not acquire lock")
	}

	// Release the lock after running all migrations.
	defer func() {
		_, e := db.Unlock()
		if err != nil && e != nil {
			err = e
		}
	}()

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
		logger.Println("nothing to do")
		return nil
	}

	for _, m := range pending {
		logger.Printf("Running %s:", m.Name)
		stmts, err := m.ReadStatements()
		if err != nil {
			return errors.Wrap(err, "error reading migration")
		}
		for _, stmt := range stmts {
			logger.Println(prefixAll("> ", stmt))
			start := time.Now()
			err := db.Exec(stmt)
			elapsed := time.Since(start)
			if err != nil {
				logger.Printf("=> FAIL (%s)", elapsed)
				return err
			}
			logger.Printf("=> OK (%v)", elapsed)
		}
		if err := db.LogCompletedMigration(m.Name); err != nil {
			return errors.Wrap(err, "error completing migration")
		}
	}
	return nil
}

// prefixAll prefixes every line in the string.
func prefixAll(prefix, stmt string) string {
	ss := strings.Split(strings.TrimSpace(stmt), "\n")
	for i, s := range ss {
		ss[i] = prefix + s
	}
	return strings.Join(ss, "\n")
}
