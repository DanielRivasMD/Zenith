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

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

var (
	csvPath     string
	headers     []string
	interactive bool
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: chalk.Yellow.Color("Add a new record to your CSV."),
	Long: chalk.Green.Color(chalk.Bold.TextStyle("zenith add")) + `
Launch a flag-driven or interactive TUI workflow to append a row to your CSV.`,
	Example: `
  zenith add --csv-path=data.csv --headers=name,age,city
  zenith add --interactive
`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Load config & flags
		path := viper.GetString("csv-path")
		hdrs := viper.GetStringSlice("headers")
		if len(headers) > 0 {
			hdrs = headers
		}

		// 2. Validate inputs
		if path == "" {
			cobra.CheckErr(fmt.Errorf("csv-path must be specified (via flag or config)"))
		}
		if len(hdrs) == 0 {
			cobra.CheckErr(fmt.Errorf("no headers defined; use --headers or set in config"))
		}

		// 3. Collect record data
		var record []string
		if interactive {
			model := newAddModel(hdrs)
			p := tea.NewProgram(model)
			out, err := p.Run()
			cobra.CheckErr(err)
			record = out.(addModel).values
		} else {
			if len(args) != len(hdrs) {
				cobra.CheckErr(fmt.Errorf("expected %d values, got %d", len(hdrs), len(args)))
			}
			record = args
		}

		// 4. Open (or create) CSV and append
		f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		cobra.CheckErr(err)
		defer f.Close()

		writer := csv.NewWriter(f)
		defer writer.Flush()

		info, err := f.Stat()
		cobra.CheckErr(err)
		if info.Size() == 0 {
			cobra.CheckErr(writer.Write(hdrs))
		}

		cobra.CheckErr(writer.Write(record))
		writer.Flush()
		cobra.CheckErr(writer.Error())

		fmt.Println(chalk.Green.Color("✅ Record added successfully!"))
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&csvPath, "csv-path", "f", "", "Path to your CSV file")
	addCmd.Flags().StringSliceVarP(&headers, "headers", "H", []string{}, "Comma-separated list of CSV headers")
	addCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Use TUI to input fields")

	viper.BindPFlag("csv-path", addCmd.Flags().Lookup("csv-path"))
	viper.BindPFlag("headers", addCmd.Flags().Lookup("headers"))

	viper.SetDefault("csv-path", "data.csv")
}

// bubbletea model that steps through each header and collects its value
type addModel struct {
	inputs  []textinput.Model
	idx     int
	values  []string
	headers []string
}

func newAddModel(headers []string) addModel {
	m := addModel{
		headers: headers,
		values:  make([]string, len(headers)),
	}

	for i, h := range headers {
		ti := textinput.New()
		ti.Placeholder = h
		ti.CharLimit = 128
		ti.Width = 30
		if i == 0 {
			ti.Focus()
		} else {
			ti.Blur()
		}
		m.inputs = append(m.inputs, ti)
	}

	return m
}

func (m addModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.values[m.idx] = m.inputs[m.idx].Value()
			if m.idx == len(m.inputs)-1 {
				return m, tea.Quit
			}
			m.inputs[m.idx].Blur()
			m.idx++
			m.inputs[m.idx].Focus()
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	m.inputs[m.idx], cmd = m.inputs[m.idx].Update(msg)
	return m, cmd
}

func (m addModel) View() string {
	var b strings.Builder
	b.WriteString("Enter record values:\n\n")
	for _, ti := range m.inputs {
		b.WriteString(ti.View() + "\n")
	}
	b.WriteString("\nPress Enter to advance, Esc to cancel.\n")
	return b.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////
