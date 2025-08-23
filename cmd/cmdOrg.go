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

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Zenith/db"
	"github.com/DanielRivasMD/Zenith/models"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var orgCmd = &cobra.Command{
	Use:               "org",
	Short:             "Manage orgs",
	PersistentPreRun:  persistentPreRun,
	PersistentPostRun: persistentPostRun,
}

var orgAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive TUI to add a new org",
	Run:   runOrgAdd,
}

var orgEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Interactive TUI to edit an existing org",
	Args:  cobra.ExactArgs(1),
	Run:   runOrgEdit,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(orgCmd)

	RegisterCrudSubcommands(orgCmd, "", CrudModel[*models.Org]{
		Singular: "org",
		ListFn: func(ctx context.Context, conn *sql.DB) ([]*models.Org, error) {
			return models.Orgs(qm.OrderBy("id ASC")).All(ctx, conn)
		},
		Format: func(o *models.Org) (int64, string) {
			return o.ID.Int64, fmt.Sprintf("%s (%s)", o.Name, o.Location.String)
		},
		RemoveFn: func(ctx context.Context, conn *sql.DB, id int64) error {
			org, err := models.FindOrg(ctx, conn, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = org.Delete(ctx, conn)
			return err
		},
	})

	// Add the interactive add/edit commands
	orgCmd.AddCommand(orgAddCmd, orgEditCmd)
}

func runOrgAdd(cmd *cobra.Command, args []string) {
	org := &models.Org{}

	fields := []Field{
		{
			Label:   "Org Name",
			Initial: "",
			Parse: func(s string) (any, error) {
				if strings.TrimSpace(s) == "" {
					return nil, fmt.Errorf("name cannot be blank")
				}
				return s, nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().FieldByName("Name").
					SetString(v.(string))
			},
		},
		{
			Label:   "Location (optional)",
			Initial: "",
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().FieldByName("Location").
					Set(reflect.ValueOf(v))
			},
		},
	}

	// Launch the Bubble Tea form wizard
	RunFormWizard(fields, org)

	// Persist new org
	if err := org.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert org: %v", err)
	}
	fmt.Printf("✅ Created org %d\n", org.ID.Int64)
}

func runOrgEdit(cmd *cobra.Command, args []string) {
	// Parse the ID
	idNum, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		log.Fatalf("invalid org ID %q: %v", args[0], err)
	}

	// Load existing record
	org, err := models.FindOrg(context.Background(), db.Conn, null.Int64From(idNum))
	if err != nil {
		log.Fatalf("find org: %v", err)
	}

	fields := []Field{
		{
			Label:   "Org Name",
			Initial: org.Name,
			Parse: func(s string) (any, error) {
				if strings.TrimSpace(s) == "" {
					return nil, fmt.Errorf("name cannot be blank")
				}
				return s, nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().FieldByName("Name").
					SetString(v.(string))
			},
		},
		{
			Label:   "Location (optional)",
			Initial: org.Location.String,
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().FieldByName("Location").
					Set(reflect.ValueOf(v))
			},
		},
	}

	RunFormWizard(fields, org)

	// Persist updates
	if _, err := org.Update(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("update org: %v", err)
	}
	fmt.Printf("✅ Updated org %d\n", org.ID.Int64)
}
