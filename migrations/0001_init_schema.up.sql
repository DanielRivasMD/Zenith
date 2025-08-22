PRAGMA foreign_keys = ON;

CREATE TABLE organizations (
	id integer PRIMARY KEY AUTOINCREMENT,
	name text NOT NULL UNIQUE,
	location text,
	created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
	updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE contacts (
	id integer PRIMARY KEY AUTOINCREMENT,
	organization_id integer NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
	name text NOT NULL,
	role TEXT,
	email text UNIQUE,
	linkedin text,
	created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
	updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE interactions (
	id integer PRIMARY KEY AUTOINCREMENT,
	contact_id integer NOT NULL REFERENCES contacts (id) ON DELETE CASCADE,
	occurred_at DATETIME NOT NULL,
	mode text,
	priority integer,
	context text,
	description text,
	follow_up text,
	comments text,
	created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
	updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE tasks (
	id integer PRIMARY KEY AUTOINCREMENT,
	interaction_id integer REFERENCES interactions (id) ON DELETE SET NULL,
	assigned_to integer,
	title text NOT NULL,
	due_date date,
	status text DEFAULT 'pending',
	notes text,
	created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
	updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

