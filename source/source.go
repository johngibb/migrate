// Package source facilitates reading (and generating) migration source
// files.
package source

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Source is handle to a directory containing migration source files.
type Source struct {
	path string
}

// New creates a new Source, or returns an error if the path does not
// exist.
func New(path string) (*Source, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.Errorf("directory does not exist: %s", path)
	}
	return &Source{path: path}, nil
}

// Migration is a handle to a migration source file.
type Migration struct {
	// Path is the path of the file.
	Path string

	// Name is the name of the migration, derived from the file name.
	Name string

	// Version is the numeric version of the migration, derived from the
	// file name and used to sort the migrations.
	Version int
}

// parseMigration parses a path into a Migration.
func parseMigration(path string) (*Migration, error) {
	base := filepath.Base(path)
	sep := strings.Index(base, "_")
	if sep == -1 {
		return nil, errors.Errorf("invalid file name: %s", base)
	}
	version, err := strconv.Atoi(base[:sep])
	if err != nil {
		return nil, errors.Errorf("invalid file name: %s", base)
	}
	name := strings.TrimSuffix(base, ".sql")
	m := &Migration{
		Path:    path,
		Name:    name,
		Version: version,
	}
	return m, nil
}

// FindMigrations finds all migrations under the source path.
func (s *Source) FindMigrations() ([]*Migration, error) {
	paths, err := filepath.Glob(filepath.Join(s.path, "*.sql"))
	if err != nil {
		return nil, errors.Wrap(err, "could not glob path")
	}
	result := make([]*Migration, len(paths))
	for i, p := range paths {
		m, err := parseMigration(p)
		if err != nil {
			return nil, err
		}
		result[i] = m
	}
	sort.Sort(ByVersion(result))
	return result, nil
}

// Create generates a new migration source file under the source path.
func (s *Source) Create(name string) (string, error) {
	timestamp := time.Now().UTC().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s.sql", timestamp, name)
	path := filepath.Join(s.path, filename)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return path, nil
}

// ByVersion sorts migrations by their version numbers.
type ByVersion []*Migration

func (ms ByVersion) Len() int      { return len(ms) }
func (ms ByVersion) Swap(i, j int) { ms[i], ms[j] = ms[j], ms[i] }
func (ms ByVersion) Less(i, j int) bool {
	return ms[i].Version < ms[j].Version
}
