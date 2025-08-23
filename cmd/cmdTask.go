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

// taskCmd is the parent for all "zenith task ..." subcommands.
var taskCmd = &cobra.Command{
	Use:               "task",
	Short:             "Manage tasks",
	Long:              helpTask,
	Example:           exampleTask,
	PersistentPreRun:  persistentPreRun,
	PersistentPostRun: persistentPostRun,
}

var taskAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive TUI to add a new task",
	Run:   runTaskAdd,
}

var taskEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Interactive TUI to edit an existing task",
	Args:  cobra.ExactArgs(1),
	Run:   runTaskEdit,
}

func init() {
	rootCmd.AddCommand(taskCmd)

	RegisterCrudSubcommands(taskCmd, "", CrudModel[*models.Task]{
		Singular: "task",
		ListFn: func(ctx context.Context, conn *sql.DB) ([]*models.Task, error) {
			return models.Tasks(qm.OrderBy("id ASC")).All(ctx, conn)
		},
		Format: func(t *models.Task) (int64, string) {
			return t.ID.Int64, fmt.Sprintf("%s (status=%s)", t.Title, t.Status.String)
		},
		RemoveFn: func(ctx context.Context, conn *sql.DB, id int64) error {
			tk, err := models.FindTask(ctx, conn, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = tk.Delete(ctx, conn)
			return err
		},
	})

	taskCmd.AddCommand(taskAddCmd, taskEditCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runTaskAdd(cmd *cobra.Command, args []string) {
	tk := &models.Task{}

	fields := []Field{
		{
			Label:   "Interaction ID (optional)",
			Initial: "",
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
					FieldByName("Interaction").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Assigned (optional)",
			Initial: "",
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
					FieldByName("Assigned").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Title",
			Initial: "",
			Parse: func(s string) (any, error) {
				if strings.TrimSpace(s) == "" {
					return nil, fmt.Errorf("title cannot be blank")
				}
				return s, nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Title").
					SetString(v.(string))
			},
		},
		{
			Label:   "Due Date (YYYY-MM-DD)",
			Initial: time.Now().Format("2006-01-02"),
			Parse: func(s string) (any, error) {
				t, err := time.Parse("2006-01-02", s)
				if err != nil {
					return nil, err
				}
				return null.TimeFrom(t), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Duedate").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Status",
			Initial: "pending",
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Status").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Notes (optional)",
			Initial: "",
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Notes").
					Set(reflect.ValueOf(v))
			},
		},
	}

	RunFormWizard(fields, tk)

	if err := tk.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert task: %v", err)
	}
	fmt.Printf("Created task %d\n", tk.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runTaskEdit(cmd *cobra.Command, args []string) {
	idNum, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		log.Fatalf("invalid task ID %q: %v", args[0], err)
	}

	tk, err := models.FindTask(context.Background(), db.Conn, null.Int64From(idNum))
	if err != nil {
		log.Fatalf("find task: %v", err)
	}

	fields := []Field{
		{
			Label:   "Interaction ID (optional)",
			Initial: strconv.FormatInt(tk.Interaction.Int64, 10),
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
					FieldByName("Interaction").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Assigned (optional)",
			Initial: strconv.FormatInt(tk.Assigned.Int64, 10),
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
					FieldByName("Assigned").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Title",
			Initial: tk.Title,
			Parse: func(s string) (any, error) {
				if strings.TrimSpace(s) == "" {
					return nil, fmt.Errorf("title cannot be blank")
				}
				return s, nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Title").
					SetString(v.(string))
			},
		},
		{
			Label:   "Due Date (YYYY-MM-DD)",
			Initial: tk.Duedate.Time.Format("2006-01-02"),
			Parse: func(s string) (any, error) {
				t, err := time.Parse("2006-01-02", s)
				if err != nil {
					return nil, err
				}
				return null.TimeFrom(t), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Duedate").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Status",
			Initial: tk.Status.String,
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Status").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Notes (optional)",
			Initial: tk.Notes.String,
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Notes").
					Set(reflect.ValueOf(v))
			},
		},
	}

	RunFormWizard(fields, tk)

	if _, err := tk.Update(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("update task: %v", err)
	}
	fmt.Printf("Updated task %d\n", tk.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
