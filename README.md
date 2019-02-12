[![Build Status](https://travis-ci.org/johngibb/migrate.svg?branch=master)](https://travis-ci.org/johngibb/migrate)
[![GoDoc](https://godoc.org/github.com/johngibb/migrate?status.svg)](https://godoc.org/github.com/johngibb/migrate)

# migrate

Migrate is an exceedingly simple Postgres database migration tool.

The simplicity is manifested in the following principles:

* There are no "down" migrations. Because migrations necessarily mutate
  state on a live server, [you can't have a safe rollback button](https://blog.skyliner.io/you-cant-have-a-rollback-button-83e914f420d9).
* Migrations are *not* automatically executed within a transaction.
  Transactions are expensive, often unnecessary, and prevent certain
  operations (e.g. `create index concurrently`). If the migration
  warrants transactional semantics, simply include `begin; ... ;
  commit;` within the source of your migration.
* Migrations are not skipped if they are added out of order with regard
  to their version number, unlike [some alternative tools](https://github.com/mattes/migrate/issues/237).
  This handles the case where a migration is added in a feature branch,
  and no longer has the highest version number when merged into master.

## Install

```
$ go get -u github.com/johngibb/migrate/cmd/migrate
```

This will install `migrate` to your $GOPATH/bin directory.

## Usage

Create a migration:

```
$ migrate create -src <folder> <migration name>:
    Creates a new migration file.
  -src string
      directory containing migration files (default ".")
```

View pending and applied migrations:

```
$ migrate status:
    Display a list of pending and applied migrations.
  -conn string
      postgres connection string
  -src string
      directory containing migration files (default ".")
```

Apply pending migrations:

```
$ migrate up -src <migrations folder> -conn <connection string> [-quiet]:
    Apply all pending migrations.
  -conn string
      postgres connection string
  -quiet
      only print errors
  -src string
      directory containing migration files (default ".")
```

## Migrations

Migrations are written as plain SQL scripts. All statements should be
terminated with a semicolon, as `migrate` will execute the script one
statement at a time.

A simple migration to add a users table might look like:

```sql
create table users (id int, name text);
```

A more complicated migration that uses a transaction to create multiple
tables and build an index might look like:

```sql
begin;
create table users (id int, name text);
create table groups (id int, name text);
create table users_groups (user_id int, group_id int);
commit;
create index concurrently on users (id);
```

# Development

To run the full integration tests, you'll need to have
[Docker for Mac](https://www.docker.com/docker-mac) installed.

```
make install  // compiles the entire project
make test     // runs the unit tests on your host machine
make test-all // compiles and runs all tests (including integration
              // tests against a Postgres instance) using Docker
```
