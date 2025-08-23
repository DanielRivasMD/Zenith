----------------------------------------------------------------------------------------------------
PRAGMA foreign_keys = ON;

----------------------------------------------------------------------------------------------------
CREATE TABLE organizations (
	id integer PRIMARY KEY AUTOINCREMENT,
	name text NOT NULL UNIQUE,
	location text,
	created DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
	updated DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

----------------------------------------------------------------------------------------------------
CREATE TABLE contacts (
	id integer PRIMARY KEY AUTOINCREMENT,
	organization integer NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
	name text NOT NULL,
	role TEXT,
	email text UNIQUE,
	linkedin text,
	created DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
	updated DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

----------------------------------------------------------------------------------------------------
CREATE TABLE interactions (
	id integer PRIMARY KEY AUTOINCREMENT,
	contact integer NOT NULL REFERENCES contacts (id) ON DELETE CASCADE,
	occurred DATETIME NOT NULL,
	mode text,
	priority integer,
	context text,
	description text,
	action text,
	comment text,
	created DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
	updated DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

----------------------------------------------------------------------------------------------------
CREATE TABLE tasks (
	id integer PRIMARY KEY AUTOINCREMENT,
	interaction integer REFERENCES interactions (id) ON DELETE SET NULL,
	assigned integer,
	title text NOT NULL,
	duedate date,
	status text DEFAULT 'pending',
	notes text,
	created DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
	updated DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

----------------------------------------------------------------------------------------------------
