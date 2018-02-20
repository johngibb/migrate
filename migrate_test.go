package migrate

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx"
)

var (
	connectionString string // set in setup
	waitOnce         sync.Once
)

func TestMigrateUp(t *testing.T) {
	setup(t)
	createMigration("1_add_users_table.sql", `
		begin;
		create table users(id int);
		commit;
		create index concurrently on users(id);
	`)
	createMigration("2_add_orders_table.sql", `create table orders(id int);`)
	mustRun("migrate --src ./migrations --conn %s up", connectionString)
}

func TestMigrateCreate(t *testing.T) {
	setup(t)
	// Create a new migration.
	mustRun("migrate --src ./migrations --conn %s create add_users_table", connectionString)

	// Confirm the new file exists.
	files, err := filepath.Glob("./migrations/*add_users_table.sql")
	must(err, "error globbing migrations")
	if len(files) == 0 {
		t.Error("file not found: *add_users_table.sql")
	}
}

func TestMigrateStatus(t *testing.T) {
	setup(t)
	// Create a new migration.
	createMigration("1_add_users_table.sql", `create table users(id int);`)

	// Run `migrate status`, ensuring the above migration is displayed
	// as pending.
	out := mustRun("migrate --src ./migrations --conn %s status", connectionString)
	if want := "1_add_users_table pending"; !strings.Contains(out, want) {
		t.Errorf("output missing: %q", want)
	}
}

func TestMigrateStatusAndUp(t *testing.T) {
	setup(t)
	// Create two migrations.
	createMigration("1_add_users_table.sql", "create table users(id int);")
	createMigration("2_add_users_table.sql", "create table orders(id int);")

	// Confirm they are listed as pending.
	out := mustRun("migrate --src ./migrations --conn %s status", connectionString)
	if want := "1_add_users_table pending"; !strings.Contains(out, want) {
		t.Errorf("output missing: %q", want)
	}
	if want := "2_add_users_table pending"; !strings.Contains(out, want) {
		t.Errorf("output missing: %q", want)
	}

	// Apply them using "migrate up".
	mustRun("migrate --src ./migrations --conn %s up", connectionString)

	// Confirm they are listed as applied.
	out = mustRun("migrate --src ./migrations --conn %s status", connectionString)
	if want := "1_add_users_table applied"; !strings.Contains(out, want) {
		t.Errorf("output missing: %q", want)
	}
	if want := "2_add_users_table applied"; !strings.Contains(out, want) {
		t.Errorf("output missing: %q", want)
	}
}

// run runs the command, returning the output and an error.
func run(cmd string, args ...interface{}) (string, error) {
	parts := strings.Split(fmt.Sprintf(cmd, args...), " ")
	out, err := exec.Command(parts[0], parts[1:]...).CombinedOutput()
	return string(out), err
}

// mustRun runs the command, calling log.Fatal if any error occurs.
func mustRun(cmd string, args ...interface{}) string {
	out, err := run(cmd, args...)
	if err != nil {
		log.Fatalf("%s:\n%v\n\n%s", err, "migrate", out)
	}
	return out
}

// must calls log.Fatal if the err is not nil.
func must(err error, msg string) {
	if err != nil {
		log.Fatalf(msg+": %v", err)
	}
}

// clearMigrations recreates an empty migrations folder for testing.
func clearMigrations() {
	must(os.RemoveAll("./migrations"), "error deleting migrations directory")
	must(os.MkdirAll("./migrations", os.ModeDir), "error creating migrations directory")
}

// resetDB drops all tables in the database.
func resetDB() {
	cfg, err := pgx.ParseURI(connectionString)
	must(err, "error parsing connection uri")
	conn, err := pgx.Connect(cfg)
	must(err, "error connecting to database")
	defer conn.Close()
	rows, err := conn.Query(`select table_name from information_schema.tables where table_schema = current_schema();`)
	must(err, "error fetching tables")
	var tables []string
	for rows.Next() {
		var name string
		must(rows.Scan(&name), "error scanning row")
		tables = append(tables, name)
	}
	for _, table := range tables {
		_, err := conn.Exec("drop table " + table)
		must(err, "error dropping table")
	}
}

// createMigration creates a new migration with the given name in the
// migrations folder.
func createMigration(name, source string) {
	f, err := os.Create("./migrations/" + name)
	must(err, "error writing migration")
	defer f.Close()
	fmt.Fprintln(f, source)
}

// setup prepares the environment for testing:
// - waits for the db to accept connections
// - clears the migrations directory
// - resets the database
func setup(t *testing.T) {
	if os.Getenv("RUN_MIGRATIONS") != "YES" {
		t.SkipNow()
	}
	uri := os.Getenv("DATABASE_URL")
	if uri == "" {
		t.Fatal("DATABASE_URL env not set")
	}

	// Wait for the DB to become available, since the `depends_on`
	// configuration in the docker-compose file waits for the postgres
	// container to finish booting, but not for the actual service to be
	// ready to accept connections.
	waitOnce.Do(func() {
		cfg, err := pgx.ParseURI(uri)
		must(err, "error parsing uri")
		for i := 0; i < 20; i++ {
			conn, connErr := pgx.Connect(cfg)
			if connErr == nil {
				conn.Close()
				return
			}
			err = connErr
			time.Sleep(200 * time.Millisecond)
		}
		must(err, "error connecting to database")
	})
	connectionString = uri
	clearMigrations()
	resetDB()
}
