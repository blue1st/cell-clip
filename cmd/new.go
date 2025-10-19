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

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Add a new setting interactively",
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Setting name: ")
		settingName, _ := reader.ReadString('\n')
		settingName = strings.TrimSpace(settingName)

		fmt.Print("Spreadsheet URL or ID: ")
		spreadsheet, _ := reader.ReadString('\n')
		spreadsheet = strings.TrimSpace(spreadsheet)

		fmt.Print("Sheet name: ")
		sheet, _ := reader.ReadString('\n')
		sheet = strings.TrimSpace(sheet)

		fmt.Print("Column (X-axis): ")
		xAxis, _ := reader.ReadString('\n')
		xAxis = strings.TrimSpace(xAxis)

		fmt.Print("Row (Y-axis): ")
		yAxisStr, _ := reader.ReadString('\n')
		yAxis, err := strconv.Atoi(strings.TrimSpace(yAxisStr))
		if err != nil {
			log.Fatalf("Invalid input for Row (Y-axis): %v", err)
		}

		newConfig := Config{
			Spreadsheet: spreadsheet,
			Sheet:       sheet,
			XAxis:       xAxis,
			YAxis:       yAxis,
		}

		usr, err := user.Current()
		if err != nil {
			log.Fatalf("Unable to get current user: %v", err)
		}

		configPath := filepath.Join(usr.HomeDir, ".cell-clip", "config.yml")
		configData, err := os.ReadFile(configPath)
		if err != nil && !os.IsNotExist(err) {
			log.Fatalf("Unable to read config file: %v", err)
		}

		var configs map[string]Config
		if len(configData) > 0 {
			err = yaml.Unmarshal(configData, &configs)
			if err != nil {
				log.Fatalf("Unable to parse config file: %v", err)
			}
		} else {
			configs = make(map[string]Config)
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

		fmt.Printf("Successfully added setting '%s' to %s\n", settingName, configPath)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
