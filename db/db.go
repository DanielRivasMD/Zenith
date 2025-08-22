/*
Copyright © 2025 Daniel Rivas <danielrivasmd@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package db

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	sqlitem "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"

	"github.com/aarondl/sqlboiler/v4/boil"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var Conn *sql.DB

////////////////////////////////////////////////////////////////////////////////////////////////////

// InitDB opens the file, applies migrations, and hooks up SQLBoiler.
func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	driver, err := sqlitem.WithInstance(db, &sqlitem.Config{})
	if err != nil {
		return nil, fmt.Errorf("initializing migrations: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("initializing migrations: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("applying migrations: %w", err)
	}

	boil.SetDB(db)
	Conn = db
	return db, nil
}

// Ctx returns a base context for all DB operations.
func Ctx() context.Context {
	return context.Background()
}

////////////////////////////////////////////////////////////////////////////////////////////////////
