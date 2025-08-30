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
package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Zenith/db"
	"github.com/DanielRivasMD/Zenith/models"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO: format cmd
// TODO: add completions for tables
var (
	exportAll bool

	exportCmd = &cobra.Command{
		Use:   "export [tables...]",
		Short: "Export one or more tables to CSV files",
		Long: `Export specified tables from the database into CSV files in the
current working directory. Supported table names:

  orgs, contacts, events, tasks

Use --all to export every supported table.`,
		Example: `  zenith export orgs
  zenith export contacts events
  zenith export --all`,
		PersistentPreRun:  persistentPreRun,
		PersistentPostRun: persistentPostRun,
		Args:              cobra.ArbitraryArgs,
		Run:               runExport,
	}
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().BoolVar(&exportAll, "all", false, "Export all supported tables")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runExport(cmd *cobra.Command, args []string) {
	// Determine which tables to export
	if exportAll {
		args = []string{"orgs", "contacts", "events", "tasks"}
	}
	if len(args) == 0 {
		// return fmt.Errorf("must specify at least one table or use --all")
	}

	// Export each requested table
	for _, table := range args {
		switch table {
		case "orgs", "organizations":
			if err := exportOrgs(cmd.Context(), db.Conn); err != nil {
				// return err
			}
		case "contacts":
			if err := exportContacts(cmd.Context(), db.Conn); err != nil {
				// return err
			}
		case "events":
			if err := exportEvents(cmd.Context(), db.Conn); err != nil {
				// return err
			}
		case "tasks":
			if err := exportTasks(cmd.Context(), db.Conn); err != nil {
				// return err
			}
		default:
			// return fmt.Errorf("unknown table %q", table)
		}
	}

}

////////////////////////////////////////////////////////////////////////////////////////////////////

func exportOrgs(ctx context.Context, conn *sql.DB) error {
	rows, err := models.Orgs(qm.OrderBy("id ASC")).All(ctx, conn)
	if err != nil {
		return fmt.Errorf("query organizations: %w", err)
	}

	file, err := os.Create("organizations.csv")
	if err != nil {
		return fmt.Errorf("create organizations.csv: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	// header
	if err := w.Write([]string{"id", "name", "location", "created", "updated"}); err != nil {
		return err
	}

	// rows
	for _, o := range rows {
		record := []string{
			strconv.FormatInt(o.ID.Int64, 10),
			o.Name,
			o.Location.String,
			o.Created.Format(time.RFC3339),
			o.Updated.Format(time.RFC3339),
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}

	fmt.Println("exported organizations.csv")
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func exportContacts(ctx context.Context, conn *sql.DB) error {
	rows, err := models.Contacts(qm.OrderBy("id ASC")).All(ctx, conn)
	if err != nil {
		return fmt.Errorf("query contacts: %w", err)
	}

	file, err := os.Create("contacts.csv")
	if err != nil {
		return fmt.Errorf("create contacts.csv: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	if err := w.Write([]string{"id", "org", "name", "role", "email", "linkedin", "created", "updated"}); err != nil {
		return err
	}

	for _, c := range rows {
		record := []string{
			strconv.FormatInt(c.ID.Int64, 10),
			strconv.FormatInt(c.Org, 10),
			c.Name,
			c.Role.String,
			c.Email.String,
			c.Linkedin.String,
			c.Created.Format(time.RFC3339),
			c.Updated.Format(time.RFC3339),
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}

	fmt.Println("exported contacts.csv")
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func exportEvents(ctx context.Context, conn *sql.DB) error {
	rows, err := models.Events(qm.OrderBy("id ASC")).All(ctx, conn)
	if err != nil {
		return fmt.Errorf("query events: %w", err)
	}

	file, err := os.Create("events.csv")
	if err != nil {
		return fmt.Errorf("create events.csv: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	header := []string{
		"id", "contact", "occurred", "mode", "priority",
		"context", "description", "action", "comment", "created", "updated",
	}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, i := range rows {
		record := []string{
			strconv.FormatInt(i.ID.Int64, 10),
			strconv.FormatInt(i.Contact, 10),
			i.Occurred.Format(time.RFC3339),
			i.Mode.String,
			strconv.FormatInt(i.Priority.Int64, 10),
			i.Context.String,
			i.Description.String,
			i.Action.String,
			i.Comment.String,
			i.Created.Format(time.RFC3339),
			i.Updated.Format(time.RFC3339),
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}

	fmt.Println("exported events.csv")
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func exportTasks(ctx context.Context, conn *sql.DB) error {
	rows, err := models.Tasks(qm.OrderBy("id ASC")).All(ctx, conn)
	if err != nil {
		return fmt.Errorf("query tasks: %w", err)
	}

	file, err := os.Create("tasks.csv")
	if err != nil {
		return fmt.Errorf("create tasks.csv: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	if err := w.Write([]string{"id", "interaction", "assigned", "title", "duedate", "status", "notes", "created", "updated"}); err != nil {
		return err
	}

	for _, t := range rows {
		record := []string{
			strconv.FormatInt(t.ID.Int64, 10),
			strconv.FormatInt(t.Interaction.Int64, 10),
			strconv.FormatInt(t.Assigned.Int64, 10),
			t.Title,
			t.Duedate.Time.Format("2006-01-02"),
			t.Status.String,
			t.Notes.String,
			t.Created.Format(time.RFC3339),
			t.Updated.Format(time.RFC3339),
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}

	fmt.Println("exported tasks.csv")
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
