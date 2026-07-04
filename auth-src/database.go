package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDatabase(dbFile string) error {
	path := resolveDbPath(dbFile)
	if err := ensureDbFile(path); err != nil {
		return fmt.Errorf("database setup: %w", err)
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	conn.SetMaxOpenConns(1)

	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS mikrotik_users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		ip TEXT NOT NULL UNIQUE,
		netmask TEXT NOT NULL
	)`); err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	db = conn
	return nil
}

func resolveDbPath(dbFile string) string {
	isAbs := filepath.IsAbs(dbFile) || (len(dbFile) > 0 && dbFile[0] == '/')

	if isAbs {
		return dbFile
	}

	execPath, err := os.Executable()
	if err != nil {
		execPath = "."
	}

	return filepath.Join(filepath.Dir(execPath), dbFile)
}

func ensureDbFile(dbFile string) error {
	dir := filepath.Dir(dbFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create db directory %s: %w", dir, err)
	}

	f, err := os.OpenFile(dbFile, os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("create db file %s: %w", dbFile, err)
	}
	f.Close()
	return nil
}
