package cli

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/febras/sota/otp"
	"github.com/febras/sota/qrcode"
	"github.com/febras/sota/storage"
	"github.com/febras/sota/terminal"
	"github.com/febras/sota/ykman"
)

var (
	readFile        string
	importMigration []string
	outputFile      string
	generateYkman   string
	exportFile      string
)

var (
	baseStyle     = lipgloss.NewStyle().Margin(1, 2)
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	itemStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).PaddingLeft(2).Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(lipgloss.Color("42"))
	otpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	msgStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("204")).MarginTop(1)
	progressStyle = lipgloss.NewStyle().MarginTop(1)
)

type tickMsg time.Time

type model struct {
	accounts    []otp.Account
	filtered    []otp.Account // Accounts displayed after filtering
	cursor      int
	message     string
	progress    progress.Model
	maxNameLen  int
	textInput   textinput.Model // Search input component
	isSearching bool            // Indicator for typing mode
}

func initialModel(accs []otp.Account) model {
	maxLen := 0
	for _, a := range accs {
		if len(a.Name) > maxLen {
			maxLen = len(a.Name)
		}
	}

	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)

	ti := textinput.New()
	ti.Placeholder = "Find account name..."
	ti.CharLimit = 50
	ti.Width = 30

	return model{
		accounts:    accs,
		filtered:    accs, // Initially display all accounts
		progress:    prog,
		maxNameLen:  maxLen,
		textInput:   ti,
		isSearching: false,
	}
}

func (m model) Init() tea.Cmd {
	// Run tick and cursor blink simultaneously
	return tea.Batch(tickCmd(), textinput.Blink)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// IF IN SEARCH MODE
		if m.isSearching {
			switch msg.String() {
			case "esc", "enter": // Exit search mode
				m.isSearching = false
				m.textInput.Blur()
				return m, nil
			}

			// Insert typing input into textInput
			m.textInput, cmd = m.textInput.Update(msg)
			cmds = append(cmds, cmd)

			// Process filtering in real-time
			term := strings.ToLower(m.textInput.Value())
			m.filtered = nil
			for _, acc := range m.accounts {
				if term == "" || strings.Contains(strings.ToLower(acc.Name), term) {
					m.filtered = append(m.filtered, acc)
				}
			}
			m.cursor = 0 // Reset cursor to the top after searching
			return m, tea.Batch(cmds...)
		}

		// IF IN NORMAL NAVIGATION MODE
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if len(m.filtered) > 0 {
				m.cursor = (m.cursor - 1 + len(m.filtered)) % len(m.filtered)
			}
			m.message = ""
		case "down", "j":
			if len(m.filtered) > 0 {
				m.cursor = (m.cursor + 1) % len(m.filtered)
			}
			m.message = ""
		case "/": // Press slash (/) to start searching
			m.isSearching = true
			m.textInput.Focus()
			m.message = ""
			return m, textinput.Blink
		case "enter":
			if len(m.filtered) > 0 {
				selectedOtp := m.filtered[m.cursor].TotpObj.Now()
				clipboard.WriteAll(selectedOtp)
				m.message = fmt.Sprintf("Copied OTP for %s!", m.filtered[m.cursor].Name)
			}
		}

	case tickMsg:
		cmds = append(cmds, tickCmd())
	}

	// Catch all non-key events (like cursor Blink) to textInput
	if m.isSearching {
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	s := titleStyle.Render("SOTA - Simple One Time Authenticator") + "\n"

	// Dynamic Help Text
	helpText := "Use arrows to navigate, Enter to copy, / to search, q to quit."
	if m.isSearching {
		helpText = "Type to search, Enter or Esc to finish."
	}
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(helpText) + "\n\n"

	// Display search bar if it is active or has a value
	if m.isSearching || m.textInput.Value() != "" {
		s += m.textInput.View() + "\n\n"
	}

	//
	if len(m.filtered) == 0 {
		s += "No accounts found.\n"
	} else {
		for i, acc := range m.filtered {
			currentOtp := acc.TotpObj.Now()
			namePadded := fmt.Sprintf("%-*s", m.maxNameLen, acc.Name)

			// List numbering (i+1) with 2-digit padding
			row := fmt.Sprintf("%2d. [%s] OTP: %s", i+1, namePadded, otpStyle.Render(currentOtp))

			if m.cursor == i {
				s += selectedStyle.Render(row) + "\n"
			} else {
				s += itemStyle.Render(row) + "\n"
			}
		}
	}

	// Render time progress
	now := time.Now().Unix()
	remainingSeconds := 30 - (now % 30)
	remainingRatio := float64(remainingSeconds) / 30.0

	barStr := m.progress.ViewAs(remainingRatio)
	s += "\n" + progressStyle.Render(fmt.Sprintf("%s %2ds", barStr, remainingSeconds)) + "\n"

	// Render Notification message
	if m.message != "" {
		s += msgStyle.Render(m.message) + "\n"
	}

	return baseStyle.Render(s)
}

var rootCmd = &cobra.Command{
	Use:   "sota",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		terminal.Clear()
		terminal.Banner()

		targetOutput := outputFile
		if targetOutput == "" {
			targetOutput = readFile
		}

		// 1. Mode: Import Migration
		if len(importMigration) > 0 {
			for _, uriOrPath := range importMigration {
				var targetUri string

				if strings.HasPrefix(uriOrPath, "otpauth-migration://") || strings.HasPrefix(uriOrPath, "otpauth://") {
					targetUri = uriOrPath
				} else {
					targetUri = qrcode.ReadURI(uriOrPath)
				}

				if targetUri != "" {
					if strings.HasPrefix(targetUri, "otpauth-migration://") {
						otpUris, _ := otp.ParseMigrationURI(targetUri)
						storage.SaveURIs(targetOutput, otpUris)
					} else if strings.HasPrefix(targetUri, "otpauth://") {
						storage.SaveURIs(targetOutput, []string{targetUri})
					}
				} else {
					color.Red("Invalid import argument or failed to parse: %s", uriOrPath)
				}
			}
			return
		}

		// 2. Mode: Generate Ykman
		if generateYkman != "" {
			targetUri := generateYkman
			if !strings.HasPrefix(targetUri, "otpauth-migration://") {
				targetUri = qrcode.ReadURI(generateYkman)
			}

			if targetUri != "" {
				fmt.Println("YubiKey Manager (ykman) commands:")
				fmt.Println(strings.Repeat("-", 50))
				cmds := ykman.GenerateCommands(targetUri)
				for _, c := range cmds {
					fmt.Println(c)
				}
				fmt.Println(strings.Repeat("-", 50))
			}
			return
		}

		// 3. Mode: Export to QR
		if exportFile != "" {
			fmt.Printf("Exporting accounts from '%s' to QR codes.\n", exportFile)
			os.MkdirAll("export", os.ModePerm)

			accounts := storage.LoadAccounts(exportFile)
			exportedCount := 0
			reg, _ := regexp.Compile("[^a-zA-Z0-9]+")

			for _, acc := range accounts {
				parsedUri, _ := url.Parse(acc.RawURI)
				label := strings.Trim(parsedUri.Path, "/")
				issuer := parsedUri.Query().Get("issuer")
				if issuer == "" {
					issuer = label
				}

				safeIssuer := reg.ReplaceAllString(issuer, "_")
				safeLabel := reg.ReplaceAllString(label, "_")
				outPath := filepath.Join("export", fmt.Sprintf("qrcode_%s_%s.png", safeIssuer, safeLabel))

				if qrcode.Generate(acc.RawURI, outPath) {
					exportedCount++
				}
			}

			if exportedCount > 0 {
				color.Green("\nSuccessfully exported %d QR codes to the 'export' directory.", exportedCount)
			} else {
				color.Red("No valid URIs found to export.")
			}
			return
		}

		// 4. Mode: Load Accounts
		accounts := storage.LoadAccounts(readFile)
		if len(accounts) == 0 {
			fmt.Printf("No accounts loaded from '%s'.\n", readFile)
			return
		}

		// 5. Run UI Bubble Tea (SUDAH DIUBAH: Menghapus tea.WithAltScreen() agar Banner terlihat)
		p := tea.NewProgram(initialModel(accounts))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&readFile, "read", "r", "accounts.txt", "Specify the file to read account URIs from.")
	rootCmd.Flags().StringSliceVarP(&importMigration, "import-migration", "i", nil, "Import accounts from a QR code, migration URI, or OTP URI.")
	rootCmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Specify the output file to write the decoded URIs.")
	rootCmd.Flags().StringVarP(&generateYkman, "generate-ykman", "g", "", "Generate ykman commands directly from a migration QR code or URI.")
	rootCmd.Flags().StringVarP(&exportFile, "export", "e", "", "Export accounts to QR codes.")
	rootCmd.Flags().Lookup("export").NoOptDefVal = "accounts.txt"

	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		terminal.Banner()
		defaultHelp(cmd, args)
	})
}
