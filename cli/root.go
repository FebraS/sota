package cli

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
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
	interactive     bool
	search          string
	importMigration []string
	outputFile      string
	generateYkman   string
	exportFile      string
)

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

		// Filter by search
		if search != "" {
			var filtered []otp.Account
			s := strings.ToLower(search)
			for _, acc := range accounts {
				if strings.Contains(strings.ToLower(acc.Name), s) {
					filtered = append(filtered, acc)
				}
			}
			accounts = filtered
			if len(accounts) == 0 {
				fmt.Printf("No accounts found matching '%s'.\n", search)
				return
			}
		}

		fmt.Printf("%d accounts loaded from '%s'.\n\n", len(accounts), readFile)
		maxNameLen := 0
		for _, acc := range accounts {
			if len(acc.Name) > maxNameLen {
				maxNameLen = len(acc.Name)
			}
		}

		// 5. Interactive Mode
		if interactive {
			fmt.Println("Interactive Mode Enabled. Choose an account to copy its OTP.\n")
			for i, acc := range accounts {
				paddedName := fmt.Sprintf("[%d] %-*s", i+1, maxNameLen, acc.Name)
				fmt.Printf("%s: %s\n", paddedName, acc.TotpObj.Now())
			}

			reader := bufio.NewReader(os.Stdin)
			for {
				fmt.Print("\nEnter account number to copy OTP (or 0 to exit): ")
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)

				choice, err := strconv.Atoi(input)
				if err != nil {
					fmt.Println("Invalid input. Please enter a number.")
					continue
				}

				if choice == 0 {
					break
				}
				if choice >= 1 && choice <= len(accounts) {
					selectedOtp := accounts[choice-1].TotpObj.Now()
					clipboard.WriteAll(selectedOtp)
					fmt.Printf("\nOTP for '%s' (%s) copied to clipboard!\n", accounts[choice-1].Name, selectedOtp)
				} else {
					fmt.Println("Invalid choice.")
				}
			}
			return
		}

		// 6. Live Mode (Default)
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			<-c
			color.Red("\n\nProgram stopped.")
			os.Exit(0)
		}()

		previousOtps := make(map[string]string)
		for {
			now := time.Now().Unix()
			remainingSeconds := 30 - (now % 30)

			currentOtps := make(map[string]string)
			otpChanged := false

			for _, acc := range accounts {
				token := acc.TotpObj.Now()
				currentOtps[acc.Name] = token
				if prev, exists := previousOtps[acc.Name]; !exists || prev != token {
					otpChanged = true
				}
			}

			if otpChanged {
				terminal.Clear()
				terminal.Banner()
				fmt.Printf("%d accounts loaded from '%s'.\n\n", len(accounts), readFile)

				for _, acc := range accounts {
					paddedName := fmt.Sprintf("[%s]", acc.Name)
					fmt.Printf("%-*s OTP: %s\n", maxNameLen+2, paddedName, color.HiGreenString(currentOtps[acc.Name]))
				}
				fmt.Println()

				if len(accounts) > 0 {
					clipboard.WriteAll(accounts[0].TotpObj.Now())
				}
				previousOtps = currentOtps
			}

			fmt.Printf("\rRemaining Time: %ds          ", remainingSeconds)
			time.Sleep(1 * time.Second)
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
	rootCmd.Flags().BoolVarP(&interactive, "interactive", "t", false, "Enable interactive mode to select which OTP to copy.")
	rootCmd.Flags().StringVarP(&search, "search", "s", "", "Search for accounts by name.")
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
