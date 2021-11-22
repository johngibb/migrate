package migrate

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	// Create two migrations./
	createMigration("1_add_users_table.sql", `
		begin;
		create table users(id int);
		commit;
		create index concurrently on users(id);
	`)
	createMigration("2_add_orders_table.sql", `create table orders(id int);`)

	// Run the migrations.
	out := mustRun("migrate up --src ./migrations --conn %s", connectionString)

	// Verify the output looks correct.
	want := regexp.MustCompile(`Running 1_add_users_table:
> begin;
=> OK \(.*\)
> create table users\(id int\);
=> OK \(.*\)
> commit;
=> OK \(.*\)
> create index concurrently on users\(id\);
=> OK \(.*\)
Running 2_add_orders_table:
> create table orders\(id int\);
=> OK \(.*\)
`)
	if !want.MatchString(out) {
		t.Errorf("output: want:\n%v\n\ngot:\n%s", want, out)
	}
}

func TestMigrateUpQuietNoError(t *testing.T) {
	setup(t)
	// Create a migration.
	createMigration("1_add_users_table.sql", "create table users(id int);")

	// Run the migration.
	out := mustRun("migrate up --src ./migrations --conn %s --quiet", connectionString)

	// Verify the output was quiet.
	if out != "" {
		t.Errorf("output: want blank, got:\n%s", out)
	}
}

func TestMigrateUpQuietError(t *testing.T) {
	setup(t)
	// Create a broken migration.
	createMigration("1_add_users_table.sql", `invalid sql statement;`)

	// Run the migrations.
	out, err := run("migrate up --src ./migrations --conn %s --quiet", connectionString)
	if err == nil {
		t.Fatal("error was nil")
	}

	// Verify the output was printed even though quiet was specified.
	want := regexp.MustCompile(`Running 1_add_users_table:
> invalid sql statement;
=> FAIL \(.*\)
migrate: ERROR: syntax error at or near "invalid" \(SQLSTATE 42601\)
`)
	if !want.MatchString(out) {
		t.Errorf("output: want:\n%v\n\ngot:\n%s", want, out)
	}
}

func TestMigrateCreate(t *testing.T) {
	setup(t)
	// Create a new migration.
	mustRun("migrate create --src ./migrations add_users_table")

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
	out := mustRun("migrate status --src ./migrations --conn %s", connectionString)
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
	out := mustRun("migrate status --src ./migrations --conn %s", connectionString)
	if want := "1_add_users_table pending"; !strings.Contains(out, want) {
		t.Errorf("output missing: %q", want)
	}
	if want := "2_add_users_table pending"; !strings.Contains(out, want) {
		t.Errorf("output missing: %q", want)
	}

	// Apply them using "migrate up".
	mustRun("migrate up --src ./migrations --conn %s", connectionString)

	// Confirm they are listed as applied.
	out = mustRun("migrate status --src ./migrations --conn %s", connectionString)
	if want := "1_add_users_table applied"; !strings.Contains(out, want) {
		t.Errorf("output missing: %q", want)
	}
	if want := "2_add_users_table applied"; !strings.Contains(out, want) {
		t.Errorf("output missing: %q", want)
	}
}

func TestDSN(t *testing.T) {
	setup(t)
	createMigration("1_add_users_table.sql", "create table users(id int);")

	// Parse the connection URI.
	cfg, err := pgx.ParseURI(connectionString)
	must(err, "parsing pg uri")

	// Convert the URI to a DSN.
	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	// Confirm the tool still works correctly with a DSN.
	out := mustRun("migrate --src ./migrations --conn '%s' status", dsn)
	if want := "1_add_users_table pending"; !strings.Contains(out, want) {
		t.Errorf("output missing: %q", want)
	}
}

func TestLocking(t *testing.T) {
	setup(t)
	createMigration("1_add_users_table.sql", "select pg_sleep(1);")

	// Run the migration twice in parallel.
	out := make([]string, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func(i int) {
			out[i], _ = run("migrate --src ./migrations --conn %s up", connectionString)
			wg.Done()
		}(i)
	}
	wg.Wait()

	// Look for one success and one failure.
	var foundSuccess, foundFailure bool
	for _, s := range out {
		if strings.Contains(s, "=> OK") {
			foundSuccess = true
		}
		if strings.Contains(s, "could not acquire lock") {
			foundFailure = true
		}
	}
	if !foundSuccess {
		t.Error("neither command succeeded")
	}
	if !foundFailure {
		t.Error("neither command failed")
	}
}

func TestLegacyCommandLineArgs(t *testing.T) {
	setup(t)
	mustRun("migrate -src ./migrations -conn %s create add_table", connectionString)
	mustRun("migrate -src ./migrations -conn %s status", connectionString)
	mustRun("migrate -src ./migrations -conn %s up", connectionString)

	mustRun("migrate -src=./migrations -conn=%s create add_table2", connectionString)
	mustRun("migrate -src=./migrations -conn=%s status", connectionString)
	mustRun("migrate -src=./migrations -conn=%s up", connectionString)
}

// run runs the command, returning the output and an error.
func run(cmd string, args ...interface{}) (string, error) {
	parts := splitCMD(fmt.Sprintf(cmd, args...))
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

// splitCMD splits apart a command by spaces, preserving strings within
// single quotes.
//
// For example: "migrate 'user=migrate password=migrate'"
//   => ["migrate", "user=migrate password=migrate"]
//
func splitCMD(s string) []string {
	r := regexp.MustCompile("'.+'|\".+\"|\\S+")
	result := r.FindAllString(s, -1)
	for i, s := range result {
		result[i] = strings.Trim(s, "'") // remove surrounding single quotes
	}
	return result
}

// clearMigrations recreates an empty migrations folder for testing.
func clearMigrations() {
	must(os.RemoveAll("./migrations"), "error deleting migrations directory")
	must(os.MkdirAll("./migrations", os.ModePerm), "error creating migrations directory")
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
