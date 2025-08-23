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
	"log"

	"github.com/DanielRivasMD/horus"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Zenith/db"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var rootCmd = &cobra.Command{
	Use:     "zenith",
	Short:   "Customize with your actual tool description",
	Long:    helpRoot,
	Example: exampleRoot,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func Execute() {
	horus.CheckErr(rootCmd.Execute())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	verbose bool
	dbPath  string // populated by the --db flag
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose diagnostics")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "zenith.db", "path to sqlite database")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func persistentPreRun(cmd *cobra.Command, args []string) {
	if _, err := db.InitDB(dbPath); err != nil {
		log.Fatalf("init DB: %v", err)
	}
}

func persistentPostRun(cmd *cobra.Command, args []string) {
	if db.Conn != nil {
		_ = db.Conn.Close()
		db.Conn = nil
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
