package db

import "hash/fnv"

func generateAdvisoryLockID(database string) int {
	h := fnv.New32a()
	h.Write([]byte(database))
	h.Write([]byte("migrate"))
	return int(h.Sum32())
}

// TryLock attempts to acquire an exclusive lock for running migrations
// on this database.
func (c *Client) TryLock() (bool, error) {
	id := generateAdvisoryLockID(c.databaseName)
	var success bool
	err := c.conn.QueryRow(`select pg_try_advisory_lock($1);`, id).Scan(&success)
	if err != nil {
		return false, err
	}
	if success {
		c.locked = true
	}
	return success, nil
}

// Unlock unlocks the exclusive migration lock.
func (c *Client) Unlock() (bool, error) {
	id := generateAdvisoryLockID(c.databaseName)
	var success bool
	err := c.conn.QueryRow(`select pg_advisory_unlock($1);`, id).Scan(&success)
	if err != nil {
		return false, err
	}
	if success {
		c.locked = false
	}
	return success, nil
}
