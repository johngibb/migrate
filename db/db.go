// Package db implements a Postgres client for applying, querying, and
// recording migrations.
package db

import (
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
)

// Migration is a migration that's been applied to the database.
type Migration struct {
	Name string
}

// Client is a migration database connection.
type Client struct {
	conn    *pgx.Conn
	ensured bool
}

// Connect connects to the Postgres database at the given uri.
func Connect(uri string) (*Client, error) {
	cfg, err := pgx.ParseURI(uri)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse uri: %s", uri)
	}
	conn, err := pgx.Connect(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to database")
	}
	return &Client{conn: conn}, nil
}

// Close closes the underlying database connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// ensureMigrationsTable ensures that the migrations table exists.
func (c *Client) ensureMigrationsTable() error {
	if c.ensured { // only need to run the full check once
		return nil
	}
	_, err := c.conn.Exec(`
        create table if not exists migrations (
            name text
        );
    `)
	if err == nil {
		c.ensured = true
	}
	return err
}

// ApplyMigration executes the given statements against the database,
// and then records that the migration has been applied.
func (c *Client) ApplyMigration(name string, stmts []string) error {
	if err := c.ensureMigrationsTable(); err != nil {
		return err
	}
	for _, s := range stmts {
		if _, err := c.conn.Exec(s); err != nil {
			return errors.Wrapf(err, "error running migration: %s", name)
		}
	}
	_, err := c.conn.Exec(`insert into migrations values ($1);`, name)
	return errors.Wrap(err, "error logging migration")
}

// GetMigrations returns all migrations that have been applied to the
// database.
func (c *Client) GetMigrations() ([]*Migration, error) {
	if err := c.ensureMigrationsTable(); err != nil {
		return nil, err
	}
	rows, err := c.conn.Query(`select name from migrations;`)
	if err != nil {
		return nil, errors.Wrap(err, "could not query migrations")
	}
	var result []*Migration
	for rows.Next() {
		var m Migration
		if err := rows.Scan(&m.Name); err != nil {
			return nil, errors.Wrap(err, "error scanning migration")
		}
		result = append(result, &m)
	}
	return result, nil
}
