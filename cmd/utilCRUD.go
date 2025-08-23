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
	"fmt"
	"log"
	"strconv"
	"strings"

	"database/sql"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Zenith/db"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

type CrudModel[T any] struct {
	Singular string
	ListFn   func(ctx context.Context, db *sql.DB) ([]T, error)
	Format   func(item T) (int64, string)
	RemoveFn func(ctx context.Context, db *sql.DB, id int64) error
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func RegisterCrudSubcommands[T any](
	parent *cobra.Command,
	dbPath string,
	desc CrudModel[T],
) {
	parent.PersistentPreRun = persistentPreRun
	parent.PersistentPostRun = persistentPostRun

	// list
	list := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List all %s", desc.Singular),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := db.Ctx()
			items, err := desc.ListFn(ctx, db.Conn)
			if err != nil {
				log.Fatalf("list %s: %v", desc.Singular, err)
			}
			for _, it := range items {
				id, human := desc.Format(it)
				fmt.Printf("%d\t%s\n", id, human)
			}
		},
	}
	parent.AddCommand(list)

	// rm
	rm := &cobra.Command{
		Use:   "rm [id]",
		Short: fmt.Sprintf("Remove a %s by ID", desc.Singular),
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				log.Fatalf("invalid id: %v", err)
			}
			ctx := db.Ctx()
			if err := desc.RemoveFn(ctx, db.Conn, raw); err != nil {
				log.Fatalf("rm %s: %v", desc.Singular, err)
			}
			fmt.Printf("Removed %s %d\n", desc.Singular, raw)
		},

		// optional: live completion of IDs
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			ctx := db.Ctx()
			items, err := desc.ListFn(ctx, db.Conn)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			var comps []string
			for _, it := range items {
				id, _ := desc.Format(it)
				s := strconv.FormatInt(id, 10)
				if toComplete == "" || strings.HasPrefix(s, toComplete) {
					comps = append(comps, s)
				}
			}
			return comps, cobra.ShellCompDirectiveNoFileComp
		},
	}
	parent.AddCommand(rm)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

type Field struct {
	Name    string                    // struct field name
	Label   string                    // what to show user
	Initial string                    // starting value input box
	Parse   func(string) (any, error) // raw string → typed value
	Assign  func(holder any, v any)   // setter write into model
	Input   textinput.Model           // the Bubble Tea textinput component
}

// FormModel drives the multi‐field wizard
type FormModel struct {
	fields []Field
	idx    int // which field is active
	holder any // model instance being modified
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// NewFormModel builds wizard over given fields & model holder
func NewFormModel(fields []Field, holder any) FormModel {
	for i := range fields {
		ti := textinput.New()
		ti.Placeholder = fields[i].Label
		ti.SetValue(fields[i].Initial)
		if i == 0 {
			ti.Focus()
		}
		fields[i].Input = ti
	}
	return FormModel{
		fields: fields,
		idx:    0,
		holder: holder,
	}
}

func (m FormModel) Init() tea.Cmd { return nil }

func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	f := &m.fields[m.idx]
	// Let the textinput handle keystrokes
	ti, cmd := f.Input.Update(msg)
	f.Input = ti

	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		raw := f.Input.Value()
		val, err := f.Parse(raw)
		if err != nil {
			// Here you might flash an error message instead of skipping
			return m, nil
		}

		// Assign the parsed value into the holder via reflect
		f.Assign(m.holder, val)

		// Advance or finish
		m.idx++
		if m.idx >= len(m.fields) {
			return m, tea.Quit
		}
		// Focus next field
		m.fields[m.idx].Input.Focus()
		return m, nil
	}

	return m, cmd
}

func (m FormModel) View() string {
	if m.idx >= len(m.fields) {
		return ""
	}
	f := m.fields[m.idx]
	header := fmt.Sprintf("[%d/%d] %s\n\n", m.idx+1, len(m.fields), f.Label)
	footer := "\n\n(enter to confirm, ctrl+c to cancel)"
	return header + f.Input.View() + footer
}

func RunFormWizard(fields []Field, holder any) {
	p := tea.NewProgram(NewFormModel(fields, holder))
	if _, err := p.StartReturningModel(); err != nil {
		log.Fatalf("form wizard failed: %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func RunFormWizardWithSubmit(
	fields []Field,
	holder any,
	onSubmit func(holder any) error,
) {
	p := tea.NewProgram(NewFormModel(fields, holder))
	if _, err := p.StartReturningModel(); err != nil {
		log.Fatalf("form wizard failed: %v", err)
	}
	// once the wizard quits, run your Insert or Update
	if err := onSubmit(holder); err != nil {
		log.Fatalf("submit failed: %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
