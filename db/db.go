// Package db implements a Postgres client for applying, querying, and
// recording migrations.
package db

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

// Migration is a migration that's been applied to the database.
type Migration struct {
	Name string
}

// Client is a migration database connection.
type Client struct {
	conn         *pgx.Conn
	databaseName string
	locked       bool
	ensured      bool
}

// Connect connects to the Postgres database at the given uri.
func Connect(ctx context.Context, uri string) (*Client, error) {
	cfg, err := pgx.ParseConfig(uri)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse uri: %s", uri)
	}
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to database")
	}
	c := &Client{
		conn:         conn,
		databaseName: cfg.Database,
	}
	return c, nil
}

// Close closes the underlying database connection.
func (c *Client) Close(ctx context.Context) error {
	return c.conn.Close(ctx)
}

// ensureMigrationsTable ensures that the migrations table exists.
func (c *Client) ensureMigrationsTable(ctx context.Context) error {
	if c.ensured { // only need to run the full check once
		return nil
	}
	_, err := c.conn.Exec(ctx, `
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
func (c *Client) Exec(ctx context.Context, sql string) error {
	if err := c.ensureMigrationsTable(ctx); err != nil {
		return err
	}
	_, err := c.conn.Exec(ctx, sql)
	return err
}

func (c *Client) LogCompletedMigration(ctx context.Context, name string) error {
	_, err := c.conn.Exec(ctx, `insert into migrations values ($1);`, name)
	return err
}

// GetMigrations returns all migrations that have been applied to the
// database.
func (c *Client) GetMigrations(ctx context.Context) ([]*Migration, error) {
	if err := c.ensureMigrationsTable(ctx); err != nil {
		return nil, err
	}
	rows, err := c.conn.Query(ctx, `select name from migrations;`)
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
