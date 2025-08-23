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

var contactCmd = &cobra.Command{
	Use:     "contact",
	Short:   "Manage contacts",
	Long:    helpContact,
	Example: exampleContact,

	PersistentPreRun:  persistentPreRun,
	PersistentPostRun: persistentPostRun,
}

var contactAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive TUI to add a new contact",
	Run:   runContactAdd,
}

var contactEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Interactive TUI to edit an existing contact",
	Args:  cobra.ExactArgs(1),
	Run:   runContactEdit,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(contactCmd)

	RegisterCrudSubcommands(contactCmd, "", CrudModel[*models.Contact]{
		Singular: "contact",
		ListFn: func(ctx context.Context, conn *sql.DB) ([]*models.Contact, error) {
			return models.Contacts(qm.OrderBy("id ASC")).All(ctx, conn)
		},
		Format: func(c *models.Contact) (int64, string) {
			return c.ID.Int64, fmt.Sprintf("%s <%s> org=%d", c.Name, c.Email.String, c.Org)
		},
		RemoveFn: func(ctx context.Context, conn *sql.DB, id int64) error {
			c, err := models.FindContact(ctx, conn, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = c.Delete(ctx, conn)
			return err
		},
	})

	contactCmd.AddCommand(contactAddCmd, contactEditCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runContactAdd(cmd *cobra.Command, args []string) {
	c := &models.Contact{}

	fields := []Field{
		{
			Label:   "Organization ID",
			Initial: "",
			Parse: func(s string) (any, error) {
				i, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, err
				}
				return null.Int64From(i), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Org").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Name",
			Initial: "",
			Parse: func(s string) (any, error) {
				if strings.TrimSpace(s) == "" {
					return nil, fmt.Errorf("name cannot be blank")
				}
				return s, nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Name").
					SetString(v.(string))
			},
		},
		{
			Label:   "Role (optional)",
			Initial: "",
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Role").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Email (optional)",
			Initial: "",
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Email").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "LinkedIn (optional)",
			Initial: "",
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Linkedin").
					Set(reflect.ValueOf(v))
			},
		},
	}

	RunFormWizard(fields, c)

	if err := c.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert contact: %v", err)
	}
	fmt.Printf("Created contact %d\n", c.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runContactEdit(cmd *cobra.Command, args []string) {
	idNum, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		log.Fatalf("invalid contact ID %q: %v", args[0], err)
	}

	c, err := models.FindContact(context.Background(), db.Conn, null.Int64From(idNum))
	if err != nil {
		log.Fatalf("find contact: %v", err)
	}

	fields := []Field{
		{
			Label:   "Organization ID",
			Initial: strconv.FormatInt(c.Org, 10),
			Parse: func(s string) (any, error) {
				i, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, err
				}
				return null.Int64From(i), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Organization").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Name",
			Initial: c.Name,
			Parse: func(s string) (any, error) {
				if strings.TrimSpace(s) == "" {
					return nil, fmt.Errorf("name cannot be blank")
				}
				return s, nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Name").
					SetString(v.(string))
			},
		},
		{
			Label:   "Role (optional)",
			Initial: c.Role.String,
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Role").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Email (optional)",
			Initial: c.Email.String,
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Email").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "LinkedIn (optional)",
			Initial: c.Linkedin.String,
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().
					FieldByName("Linkedin").
					Set(reflect.ValueOf(v))
			},
		},
	}

	RunFormWizard(fields, c)

	if _, err := c.Update(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("update contact: %v", err)
	}
	fmt.Printf("Updated contact %d\n", c.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
