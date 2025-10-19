package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var editCmd = &cobra.Command{
	Use:   "edit [setting_name]",
	Short: "Edit an existing setting interactively",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		settingName := args[0]

		usr, err := user.Current()
		if err != nil {
			log.Fatalf("Unable to get current user: %v", err)
		}

		configPath := filepath.Join(usr.HomeDir, ".cell-clip", "config.yml")
		configData, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatalf("Unable to read config file: %v", err)
		}

		var configs map[string]Config
		err = yaml.Unmarshal(configData, &configs)
		if err != nil {
			log.Fatalf("Unable to parse config file: %v", err)
		}

		config, ok := configs[settingName]
		if !ok {
			log.Fatalf("Setting '%s' not found in config file", settingName)
		}

		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("Spreadsheet URL or ID (current: %s): ", config.Spreadsheet)
		spreadsheet, _ := reader.ReadString('\n')
		spreadsheet = strings.TrimSpace(spreadsheet)
		if spreadsheet == "" {
			spreadsheet = config.Spreadsheet
		}

		fmt.Printf("Sheet name (current: %s): ", config.Sheet)
		sheet, _ := reader.ReadString('\n')
		sheet = strings.TrimSpace(sheet)
		if sheet == "" {
			sheet = config.Sheet
		}

		fmt.Printf("Column (X-axis) (current: %s): ", config.XAxis)
		xAxis, _ := reader.ReadString('\n')
		xAxis = strings.TrimSpace(xAxis)
		if xAxis == "" {
			xAxis = config.XAxis
		}

		fmt.Printf("Row (Y-axis) (current: %d): ", config.YAxis)
		yAxisStr, _ := reader.ReadString('\n')
		yAxisStr = strings.TrimSpace(yAxisStr)
		var yAxis int
		if yAxisStr == "" {
			yAxis = config.YAxis
		} else {
			yAxis, err = strconv.Atoi(yAxisStr)
			if err != nil {
				log.Fatalf("Invalid input for Row (Y-axis): %v", err)
			}
		}

		newConfig := Config{
			Spreadsheet: spreadsheet,
			Sheet:       sheet,
			XAxis:       xAxis,
			YAxis:       yAxis,
		}

		configs[settingName] = newConfig

		newData, err := yaml.Marshal(&configs)
		if err != nil {
			log.Fatalf("Unable to marshal new config: %v", err)
		}

		configDir := filepath.Dir(configPath)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0755); err != nil {
				log.Fatalf("Unable to create config directory: %v", err)
			}
		}

		err = os.WriteFile(configPath, newData, 0644)
		if err != nil {
			log.Fatalf("Unable to write to config file: %v", err)
		}

		fmt.Printf("Successfully updated setting '%s' in %s\n", settingName, configPath)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
