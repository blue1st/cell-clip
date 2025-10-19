# cell-clip

A command-line tool for copying cell values from Google Sheets to the clipboard.

## Features

- **OAuth 2.0 Authentication**: Secure authentication with Google Sheets using PKCE
- **Setting Management**: Save and manage multiple spreadsheet configurations
- **Clipboard Integration**: Automatically copy cell values to clipboard
- **Interactive Setup**: Easy-to-use commands for configuration

## Installation

```bash
go build -o cell-clip
```

## Setup

### 1. Google Cloud Console Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create a new project or select an existing one
3. Enable the Google Sheets API
4. Create OAuth 2.0 Client IDs for **Desktop application**
5. You will get a **Client ID** and **Client Secret**.

### 2. Configure Authentication

Run the `auth setup` command and paste your Client ID and Client Secret when prompted:

```bash
./cell-clip auth setup
```

This will create a `credentials.json` file in `~/.cell-clip/`.

### 3. Authenticate

Run the `auth login` command to authenticate with Google Sheets:

```bash
./cell-clip auth login
```

This will open a browser window for you to grant access. After you approve, the tool will be authenticated.

## Usage

### Commands

- `cell-clip auth`: Manage authentication with subcommands:
    - `login`: Authenticate with Google Sheets.
    - `logout`: Remove the stored authentication token.
    - `setup`: Interactively set up your Google OAuth credentials.
- `cell-clip new`: Add a new setting interactively.
- `cell-clip list`: List all registered settings.
- `cell-clip edit <setting_name>`: Edit an existing setting.
- `cell-clip get <setting_name>`: Get a cell value and copy it to the clipboard.

### Example Workflow

1. **Setup Credentials**:
   ```bash
   ./cell-clip auth setup
   ```

2. **Authenticate**:
   ```bash
   ./cell-clip auth login
   ```

3. **Add a setting**:
   ```bash
   ./cell-clip new
   # Follow the prompts to enter:
   # - Setting name: my-sheet
   # - Spreadsheet URL or ID: https://docs.google.com/spreadsheets/d/SPREADSHEET_ID/edit
   # - Sheet name: Sheet1
   # - Column: A
   # - Row: 1
   ```

4. **List settings**:
   ```bash
   ./cell-clip list
   ```

5. **Get cell value**:
   ```bash
   ./cell-clip get my-sheet
   # The value from cell A1 will be copied to your clipboard
   ```

## Security Features

- **PKCE (Proof Key for Code Exchange)**: Enhanced security for the OAuth 2.0 flow.
- **Secure Token Storage**: Tokens are stored with restricted file permissions (`0600`).
- **Automatic Token Refresh**: Access tokens are automatically refreshed when they expire.

## Configuration

Settings are stored in `~/.cell-clip/config.yml` in YAML format:

```yaml
my-sheet:
  spreadsheet: "https://docs.google.com/spreadsheets/d/SPREADSHEET_ID/edit"
  sheet: "Sheet1"
  x_axis: "A"
  y_axis: 1
```

## Troubleshooting

### Authentication Issues
- Ensure your OAuth credentials are correctly configured by running `cell-clip auth setup` again.
- Check that the Google Sheets API is enabled in your Google Cloud project.
- Try running `cell-clip auth logout` followed by `cell-clip auth login`.

### Permission Issues
- Make sure the spreadsheet is accessible to the Google account you authenticated with.
- Verify that the sheet name and cell coordinates are correct when creating a setting.