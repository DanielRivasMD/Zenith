////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// formatHelp produces the “help” header + description.
//
//	author: name, e.g. "Daniel Rivas"
//	email:  email, e.g. "danielrivasmd@gmail.com"
//	desc:   the multi‐line description, "\n"-separated.
func formatHelp(author, email, desc string) string {
	header := chalk.Bold.TextStyle(
		chalk.Green.Color(author+" "),
	) +
		chalk.Dim.TextStyle(
			chalk.Italic.TextStyle("<"+email+">"),
		)

	// prefix two newlines to your desc, chalk it cyan + dim it
	body := "\n\n" + desc
	return header + chalk.Dim.TextStyle(chalk.Cyan.Color(body))
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpRoot = formatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"",
)

var helpMigrate = formatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Apply any pending up or down migration scripts against the configured SQLite database",
)

var helpOrg = formatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Apply any pending up or down migration scripts against the configured SQLite database",
)

var helpContact = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"",
)

var helpEvent = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"",
)

var helpTask = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"",
)

////////////////////////////////////////////////////////////////////////////////////////////////////
