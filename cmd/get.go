package cmd

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "os/user"
    "path/filepath"
    "regexp"
    "sort"
    "strconv"
    "strings"

    "github.com/atotto/clipboard"
    "github.com/spf13/cobra"
    "google.golang.org/api/sheets/v4"
    "gopkg.in/yaml.v2"
)

var getCmd = &cobra.Command{
    Use:   "get [setting_name]",
    Short: "Get a cell value from Google Sheets and copy it to the clipboard",
    Args:  cobra.MaximumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        var settingName string
        // 設定名が指定されていない場合は対話的に選択
        if len(args) == 0 {
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

            // 名前をソートして表示
            var names []string
            for name := range configs {
                names = append(names, name)
            }
            sort.Strings(names)
            fmt.Println("Select a setting:")
            for i, n := range names {
                fmt.Printf("%d) %s\n", i+1, n)
            }
            fmt.Print("Enter number or name: ")
            reader := bufio.NewReader(os.Stdin)
            input, _ := reader.ReadString('\n')
            input = strings.TrimSpace(input)
            // 数字が入力された場合はインデックスに変換
            if idx, err := strconv.Atoi(input); err == nil && idx >= 1 && idx <= len(names) {
                settingName = names[idx-1]
            } else {
                settingName = input
            }
        } else {
            settingName = args[0]
        }

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

		oauthManager, err := NewOAuthManager()
		if err != nil {
			log.Fatalf("Unable to initialize OAuth manager: %v", err)
		}

		client, err := oauthManager.GetAuthenticatedClient()
		if err != nil {
			log.Fatalf("Unable to get authenticated client: %v", err)
		}

		srv, err := sheets.New(client)
		if err != nil {
			log.Fatalf("Unable to retrieve Sheets client: %v", err)
		}

		readRange := fmt.Sprintf("%s!%s%d", config.Sheet, config.XAxis, config.YAxis)
		// Extract spreadsheet ID if a full URL is provided.
		spreadsheetID := config.Spreadsheet
		if strings.Contains(spreadsheetID, "/d/") {
			re := regexp.MustCompile(`/d/([^/?#]+)`) // capture characters after /d/ up to '/', '?' or '#'
			matches := re.FindStringSubmatch(spreadsheetID)
			if len(matches) > 1 {
				spreadsheetID = matches[1]
			}
		}
		resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve data from sheet: %v", err)
		}

		if len(resp.Values) == 0 {
			fmt.Println("No data found.")
		} else {
			cellValue := fmt.Sprintf("%v", resp.Values[0][0])
			clipboard.WriteAll(cellValue)
			fmt.Printf("Copied to clipboard: %s\n", cellValue)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
