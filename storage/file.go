package storage

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/febras/sota/otp"
)

func LoadAccounts(filename string) []otp.Account {
	if filename == "" {
		filename = "accounts.txt"
	}

	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Error: File '%s' not found.\n", filename)
		} else {
			fmt.Printf("An error occurred while reading the file: %v\n", err)
		}
		return nil
	}
	defer file.Close()

	var accounts []otp.Account
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		uri := strings.TrimSpace(line)

		if uri != "" {
			totpObj, name, err := otp.ParseURI(uri)
			if err == nil && totpObj != nil {
				accounts = append(accounts, otp.Account{
					TotpObj: totpObj,
					Name:    name,
					RawURI:  uri,
				})
			} else {
				fmt.Printf("Warning: Skipping invalid line -> %s\n", uri)
			}
		}
	}

	return accounts
}

func SaveURIs(fileName string, uris []string) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open file for writing: %v\n", err)
		return
	}
	defer file.Close()

	for _, uri := range uris {
		file.WriteString(uri + "\n")
	}
	
	fmt.Printf("Migration URIs successfully added to '%s'.\n", fileName)
}