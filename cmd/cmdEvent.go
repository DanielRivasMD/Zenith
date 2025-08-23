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
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Zenith/db"
	"github.com/DanielRivasMD/Zenith/models"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// eventCmd is the parent for all "zenith event ..." subcommands.
var eventCmd = &cobra.Command{
	Use:               "event",
	Short:             "Manage events",
	Long:              helpEvent,
	Example:           exampleEvent,
	PersistentPreRun:  persistentPreRun,
	PersistentPostRun: persistentPostRun,
}

var eventAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive TUI to add a new event",
	Run:   runEventAdd,
}

var eventEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Interactive TUI to edit an existing event",
	Args:  cobra.ExactArgs(1),
	Run:   runEventEdit,
}

func init() {
	rootCmd.AddCommand(eventCmd)

	// list & rm are wired up generically
	RegisterCrudSubcommands(eventCmd, "", CrudModel[*models.Event]{
		Singular: "event",
		ListFn: func(ctx context.Context, conn *sql.DB) ([]*models.Event, error) {
			return models.Events(qm.OrderBy("id ASC")).All(ctx, conn)
		},
		Format: func(e *models.Event) (int64, string) {
			// ID is null.Int64, Occurred is time.Time, Mode is null.String
			return e.ID.Int64, fmt.Sprintf("%s at %s", e.Mode.String, e.Occurred.Format("2006-01-02 15:04"))
		},
		RemoveFn: func(ctx context.Context, conn *sql.DB, id int64) error {
			e, err := models.FindEvent(ctx, conn, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = e.Delete(ctx, conn)
			return err
		},
	})

	eventCmd.AddCommand(eventAddCmd, eventEditCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runEventAdd(cmd *cobra.Command, args []string) {
	e := &models.Event{}

	fields := []Field{
		{
			Label:   "Contact ID",
			Initial: "",
			Parse: func(s string) (any, error) {
				if strings.TrimSpace(s) == "" {
					return int64(0), nil
				}
				i, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, err
				}
				return i, nil
			},
			Assign: func(holder any, v any) {
				rv := reflect.ValueOf(holder).Elem()
				fv := rv.FieldByName("Contact")
				if !fv.IsValid() || !fv.CanSet() {
					log.Fatalf("cannot set Contact on %T", holder)
				}
				fv.SetInt(v.(int64))
			},
		},
		{
			Label:   "Occurred At (YYYY-MM-DD HH:MM)",
			Initial: time.Now().Format("2006-01-02 15:04"),
			Parse: func(s string) (any, error) {
				t, err := time.Parse("2006-01-02 15:04", s)
				if err != nil {
					return nil, err
				}
				return t, nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Occurred").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Mode",
			Initial: "",
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Mode").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Priority (integer)",
			Initial: "0",
			Parse: func(s string) (any, error) {
				if strings.TrimSpace(s) == "" {
					return null.Int64{}, nil
				}
				i, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, err
				}
				return null.Int64From(i), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Priority").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Context (optional)",
			Initial: "",
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Context").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Description (optional)",
			Initial: "",
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Description").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Action (optional)",
			Initial: "",
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Action").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Comment (optional)",
			Initial: "",
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Comment").
					Set(reflect.ValueOf(v))
			},
		},
	}

	RunFormWizard(fields, e)

	if err := e.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert event: %v", err)
	}
	fmt.Printf("Created event %d\n", e.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runEventEdit(cmd *cobra.Command, args []string) {
	idNum, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		log.Fatalf("invalid event ID %q: %v", args[0], err)
	}

	e, err := models.FindEvent(context.Background(), db.Conn, null.Int64From(idNum))
	if err != nil {
		log.Fatalf("find event: %v", err)
	}

	fields := []Field{
		{
			Label:   "Contact ID",
			Initial: strconv.FormatInt(e.Contact, 10),
			Parse: func(s string) (any, error) {
				if strings.TrimSpace(s) == "" {
					return int64(0), nil
				}
				i, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, err
				}
				return i, nil
			},
			Assign: func(holder any, v any) {
				rv := reflect.ValueOf(holder).Elem()
				fv := rv.FieldByName("Contact")
				if !fv.IsValid() || !fv.CanSet() {
					log.Fatalf("cannot set Contact on %T", holder)
				}
				fv.SetInt(v.(int64))
			},
		},
		{
			Label:   "Occurred At (YYYY-MM-DD HH:MM)",
			Initial: e.Occurred.Format("2006-01-02 15:04"),
			Parse: func(s string) (any, error) {
				t, err := time.Parse("2006-01-02 15:04", s)
				if err != nil {
					return nil, err
				}
				return t, nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Occurred").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Mode",
			Initial: e.Mode.String,
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Mode").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Priority (integer)",
			Initial: strconv.FormatInt(e.Priority.Int64, 10),
			Parse: func(s string) (any, error) {
				if strings.TrimSpace(s) == "" {
					return null.Int64{}, nil
				}
				i, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, err
				}
				return null.Int64From(i), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Priority").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Context (optional)",
			Initial: e.Context.String,
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Context").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Description (optional)",
			Initial: e.Description.String,
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Description").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Action (optional)",
			Initial: e.Action.String,
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Action").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Comment (optional)",
			Initial: e.Comment.String,
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Comment").
					Set(reflect.ValueOf(v))
			},
		},
	}

	RunFormWizard(fields, e)

	if _, err := e.Update(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("update event: %v", err)
	}
	fmt.Printf("Updated event %d\n", e.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
