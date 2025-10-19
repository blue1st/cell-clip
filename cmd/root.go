package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
)

// Config represents a single setting for a spreadsheet.

type Config struct {
	Spreadsheet string `yaml:"spreadsheet"`
	Sheet       string `yaml:"sheet"`
	XAxis       string `yaml:"x_axis"`
	YAxis       int    `yaml:"y_axis"`
}

var rootCmd = &cobra.Command{
	Use:   "cell-clip",
	Short: "A CLI tool to get cell values from Google Sheets",
	Long:  `A CLI tool to get cell values from Google Sheets.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

