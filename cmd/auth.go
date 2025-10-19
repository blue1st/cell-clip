package cmd

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Credentials holds the client ID and secret.
type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// loadCredentials loads client credentials from the credentials file.
func loadCredentials() (*Credentials, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("unable to get current user: %w", err)
	}

	credPath := filepath.Join(usr.HomeDir, ".cell-clip", "credentials.json")
	file, err := os.Open(credPath)
	if err != nil {
		return nil, fmt.Errorf("could not open credentials file '%s'. Please run 'cell-clip auth setup' to configure credentials: %w", credPath, err)
	}
	defer file.Close()

	creds := &Credentials{}
	if err := json.NewDecoder(file).Decode(creds); err != nil {
		return nil, fmt.Errorf("could not decode credentials file: %w", err)
	}

	if creds.ClientID == "" || creds.ClientSecret == "" {
		return nil, fmt.Errorf("invalid credentials format in '%s': client_id and client_secret must be set. You can use 'cell-clip auth setup' to create the file", credPath)
	}

	return creds, nil
}

// OAuthManager handles OAuth 2.0 authentication flow
type OAuthManager struct {
	config *oauth2.Config
}

// NewOAuthManager creates a new OAuth manager
func NewOAuthManager() (*OAuthManager, error) {
	creds, err := loadCredentials()
	if err != nil {
		return nil, err
	}

	config := &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  google.Endpoint.AuthURL,
			TokenURL: google.Endpoint.TokenURL,
		},
		Scopes:      []string{"https://www.googleapis.com/auth/spreadsheets.readonly"},
		RedirectURL: "urn:ietf:wg:oauth:2.0:oob",
	}
	config.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	return &OAuthManager{config: config}, nil
}

// GetAuthenticatedClient returns an authenticated HTTP client
func (om *OAuthManager) GetAuthenticatedClient() (*http.Client, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("unable to get current user: %w", err)
	}

	tokFile := filepath.Join(usr.HomeDir, ".cell-clip", "token.json")
	tok, err := om.tokenFromFile(tokFile)
	if err != nil {
		fmt.Println("No valid token found. Starting OAuth flow...")
		tok, err = om.getTokenFromWeb()
		if err != nil {
			return nil, fmt.Errorf("failed to get token from web: %w", err)
		}
		om.saveToken(tokFile, tok)
	}

	return om.config.Client(context.Background(), tok), nil
}

// getTokenFromWeb implements the OAuth 2.0 PKCE flow.
func (om *OAuthManager) getTokenFromWeb() (*oauth2.Token, error) {
	// Use out-of-band redirect (no local server) because the sandbox
	// prevents listening on a TCP port. "urn:ietf:wg:oauth:2.0:oob" is a special
	// value that tells Google to display the authorization code directly to the
	// user.

	// Generate PKCE code verifier and challenge
	codeVerifier, err := om.generateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	codeChallenge := om.generateCodeChallenge(codeVerifier)

	authURL := om.config.AuthCodeURL(
		"state-token",
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	fmt.Printf("Opening browser for authentication...\n")
	fmt.Printf("If the browser doesn't open automatically, please visit:\n%s\n", authURL)

	// Open browser if possible
	_ = browser.OpenURL(authURL) // ignore error; user can copy URL manually

	// Prompt user to paste the code displayed by Google.
	fmt.Print("Enter the authorization code: ")
	var authCode string
	if _, err := fmt.Scanln(&authCode); err != nil {
		// In non‑interactive environments stdin may be closed, causing EOF.
		// Return a clear error so callers know manual input is required.
		return nil, fmt.Errorf("failed to read authorization code (no input provided): %w", err)
	}

	// Exchange authorization code for token

token, err := om.config.Exchange(
		context.Background(),
		authCode,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}
	return token, nil
}

// generateCodeVerifier generates a cryptographically random code verifier
func (om *OAuthManager) generateCodeVerifier() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes), nil
}

// generateCodeChallenge generates a code challenge from the verifier
func (om *OAuthManager) generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
}

// tokenFromFile retrieves a token from a local file
func (om *OAuthManager) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	if err != nil {
		return nil, err
	}

	// Check if token is expired and refresh if possible
	if !tok.Valid() {
		if tok.RefreshToken != "" {
			// Try to refresh the token
			ctx := context.Background()
			tokenSource := om.config.TokenSource(ctx, tok)
			newToken, err := tokenSource.Token()
			if err != nil {
				return nil, fmt.Errorf("token refresh failed: %w", err)
			}
			// After a refresh, the token in the file is outdated.
			// Save the new token to ensure the refresh token is not lost.
			om.saveToken(file, newToken)
			return newToken, nil
		}
		return nil, fmt.Errorf("token is expired and no refresh token available")
	}

	return tok, nil
}

// saveToken saves a token to a file
func (om *OAuthManager) saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving token to: %s\n", path)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		log.Printf("Warning: Could not create directory %s: %v", dir, err)
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("Warning: Could not save token: %v", err)
		return
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(token); err != nil {
		log.Printf("Warning: Could not encode token: %v", err)
	}
}
