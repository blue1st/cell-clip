package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered setting names",
	Run: func(cmd *cobra.Command, args []string) {
		usr, err := user.Current()
		if err != nil {
			log.Fatalf("Unable to get current user: %v", err)
		}

		configPath := filepath.Join(usr.HomeDir, ".cell-clip", "config.yml")
		configData, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No settings found.")
				return
			}
			log.Fatalf("Unable to read config file: %v", err)
		}

		var configs map[string]Config
		err = yaml.Unmarshal(configData, &configs)
		if err != nil {
			log.Fatalf("Unable to parse config file: %v", err)
		}

		fmt.Println("Registered setting names:")
		for name := range configs {
			fmt.Println("- ", name)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
