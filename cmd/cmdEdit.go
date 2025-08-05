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
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	recordIndex int
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var editCmd = &cobra.Command{
	Use:   "edit [record]",
	Short: chalk.Yellow.Color("Edit an existing record in your CSV."),
	Long: chalk.Green.Color(chalk.Bold.TextStyle("zenith edit")) + `
Load a record by its row number, modify fields via TUI (default) or CLI, and save changes back to the CSV.`,
	Example: `
  # Interactive edit using config.toml
  zenith edit 3 --config=config.toml

  # Non-interactive edit, overriding config
  zenith edit 5 --no-tui --csv-path=data.csv --headers="Name,Age,City" Alice 30 Oslo
`,

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Args: cobra.MinimumNArgs(1),

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Run: func(cmd *cobra.Command, args []string) {
		// 1. Parse record index
		idx, err := strconv.Atoi(args[0])
		cobra.CheckErr(err)
		recordIndex = idx

		// 2. Load config file if provided
		if config != "" {
			viper.SetConfigFile(config)
			cobra.CheckErr(viper.ReadInConfig())
		}

		// 3. Bind override flags
		cobra.CheckErr(viper.BindPFlag("csv-path", cmd.Flags().Lookup("csv-path")))
		cobra.CheckErr(viper.BindPFlag("headers", cmd.Flags().Lookup("headers")))

		// 4. Resolve effective config values
		path := viper.GetString("csv-path")
		hdrs := viper.GetStringSlice("headers")
		if len(headers) > 0 {
			hdrs = headers
		}

		if path == "" {
			cobra.CheckErr(fmt.Errorf("csv-path must be specified"))
		}
		if len(hdrs) == 0 {
			cobra.CheckErr(fmt.Errorf("headers must be defined"))
		}

		// 5. Read entire CSV into memory
		f, err := os.Open(path)
		cobra.CheckErr(err)
		defer f.Close()

		reader := csv.NewReader(f)
		allRows, err := reader.ReadAll()
		cobra.CheckErr(err)

		if recordIndex < 1 || recordIndex >= len(allRows) {
			cobra.CheckErr(fmt.Errorf("record must be between 1 and %d", len(allRows)-1))
		}

		// 6. Extract the target record for editing
		orig := allRows[recordIndex]

		// 7. Collect updated values
		var updated []string
		if noTUI {
			// Expect CLI args after the record index
			newVals := args[1:]
			if len(newVals) != len(hdrs) {
				cobra.CheckErr(fmt.Errorf(
					"expected %d fields, got %d", len(hdrs), len(newVals),
				))
			}
			updated = newVals
		} else {
			model := newEditModel(hdrs, orig)
			p := tea.NewProgram(model)
			out, err := p.Run()
			cobra.CheckErr(err)
			updated = out.(editModel).values
		}

		// 8. Replace the row in memory
		allRows[recordIndex] = updated

		// 9. Write out to a temp file and atomically replace
		tmpPath := path + ".tmp"
		tmpDir := filepath.Dir(path)
		tmpFile := filepath.Join(tmpDir, filepath.Base(tmpPath))
		wf, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		cobra.CheckErr(err)
		defer wf.Close()

		writer := csv.NewWriter(wf)
		cobra.CheckErr(writer.WriteAll(allRows))
		writer.Flush()
		cobra.CheckErr(writer.Error())

		cobra.CheckErr(os.Rename(tmpFile, path))

		fmt.Println(chalk.Green.Color("✅ Record edited successfully!"))
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVarP(&config, "config", "c", "", "TOML config file with csv-path & headers")
	editCmd.Flags().BoolVar(&noTUI, "no-tui", false, "Disable TUI; use CLI args for field values")
	editCmd.Flags().StringVarP(&csvPath, "csv-path", "f", "", "Path to your CSV file")
	editCmd.Flags().StringSliceVarP(&headers, "headers", "H", []string{}, "Comma-separated list of CSV headers")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// bubbletea model for editing an existing record
type editModel struct {
	inputs  []textinput.Model
	idx     int
	values  []string
	headers []string
}

func newEditModel(headers, orig []string) editModel {
	m := editModel{
		headers: headers,
		values:  make([]string, len(headers)),
	}
	copy(m.values, orig)

	for i, h := range headers {
		ti := textinput.New()
		ti.Placeholder = h
		ti.SetValue(orig[i])
		ti.CharLimit = 256
		ti.Width = 40
		if i == 0 {
			ti.Focus()
		} else {
			ti.Blur()
		}
		m.inputs = append(m.inputs, ti)
	}
	return m
}

func (m editModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m editModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m editModel) View() string {
	var b strings.Builder
	b.WriteString("Editing record — use Enter to advance, Esc to cancel:\n\n")
	for _, ti := range m.inputs {
		b.WriteString(ti.View() + "\n")
	}
	return b.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////
