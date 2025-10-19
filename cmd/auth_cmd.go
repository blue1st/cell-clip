package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication with Google Sheets",
	Long: "Manage authentication with Google Sheets.\n\n" +
		"This command provides subcommands to authenticate, logout, and setup credentials.",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Google Sheets",
	Long: "Authenticate with Google Sheets to access your spreadsheets.\n\n" +
		"This command initiates the OAuth 2.0 flow to obtain an access token.\n\n" +
		"Before running this command, you must create a credentials file at:\n" +
		"~/.cell-clip/credentials.json\n\n" +
		"This file should contain your Google OAuth 2.0 client ID and secret in the following format:\n" +
		"{\n" +
		"  \"client_id\": \"YOUR_CLIENT_ID\",\n" +
		"  \"client_secret\": \"YOUR_CLIENT_SECRET\"\n" +
		"}\n\n" +
		"You can create this file manually or use the 'cell-clip auth setup' command.\n\n" +
		"After creating the file, run this command to start the authentication process.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting Google Sheets authentication...")

		oauthManager, err := NewOAuthManager()
		if err != nil {
			log.Fatalf("Unable to initialize OAuth manager: %v", err)
		}

		_, err = oauthManager.GetAuthenticatedClient()
		if err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}

		usr, err := user.Current()
		if err != nil {
			log.Fatalf("Unable to get current user: %v", err)
		}

		tokenPath := filepath.Join(usr.HomeDir, ".cell-clip", "token.json")
		fmt.Printf("✓ Authentication successful!\n")
		fmt.Printf("✓ Token saved to: %s\n", tokenPath)
		fmt.Printf("✓ You can now use 'cell-clip get <setting_name>' to access your sheets.\n")
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored authentication token",
	Long: "Remove the stored authentication token to sign out from Google Sheets.\n" +
		"This will require re-authentication for the next API call.",
	Run: func(cmd *cobra.Command, args []string) {
		usr, err := user.Current()
		if err != nil {
			log.Fatalf("Unable to get current user: %v", err)
		}

		tokenPath := filepath.Join(usr.HomeDir, ".cell-clip", "token.json")

		if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
			fmt.Println("No authentication token found. You are already logged out.")
			return
		}

		if err := os.Remove(tokenPath); err != nil {
			log.Fatalf("Unable to remove token file: %v", err)
		}

		fmt.Println("✓ Successfully logged out.")
		fmt.Println("✓ Authentication token removed.")
		fmt.Println("✓ You will need to run 'cell-clip auth login' before using the tool again.")
	},
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactively setup Google OAuth credentials",
	Long: "This command will prompt you for your Google OAuth 2.0 client ID and secret,\n" +
		"and create the 'credentials.json' file in '~/.cell-clip/'.",
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter your Google OAuth 2.0 Client ID: ")
		clientID, _ := reader.ReadString('\n')
		clientID = strings.TrimSpace(clientID)

		fmt.Print("Enter your Google OAuth 2.0 Client Secret: ")
		clientSecret, _ := reader.ReadString('\n')
		clientSecret = strings.TrimSpace(clientSecret)

		creds := &Credentials{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}

		usr, err := user.Current()
		if err != nil {
			log.Fatalf("Unable to get current user: %v", err)
		}

		credPath := filepath.Join(usr.HomeDir, ".cell-clip", "credentials.json")
		if err := os.MkdirAll(filepath.Dir(credPath), 0700); err != nil {
			log.Fatalf("Unable to create directory %s: %v", filepath.Dir(credPath), err)
		}

		file, err := os.Create(credPath)
		if err != nil {
			log.Fatalf("Unable to create credentials file: %v", err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(creds); err != nil {
			log.Fatalf("Unable to write to credentials file: %v", err)
		}

		fmt.Printf("\n✓ Credentials saved to: %s\n", credPath)
		fmt.Println("✓ You can now run 'cell-clip auth login' to authenticate with Google Sheets.")
	},
}

func init() {
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(authCmd)
}
