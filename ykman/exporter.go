package ykman

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/febras/sota/otp"
)

func GenerateCommands(migrationUri string) []string {
	var commands []string
	otpUris, err := otp.ParseMigrationURI(migrationUri)
	if err != nil {
		fmt.Printf("Error parsing migration URI: %v\n", err)
		return commands
	}

	for _, uri := range otpUris {
		parsedUri, err := url.Parse(uri)
		if err != nil {
			continue
		}

		query := parsedUri.Query()
		ykargs := []string{"oath", "add"}

		otpType := strings.ToLower(parsedUri.Host)
		if otpType == "totp" {
			ykargs = append(ykargs, "-o", "TOTP", "-p", "30")
		} else if otpType == "hotp" {
			ykargs = append(ykargs, "-o", "HOTP")
		} else {
			continue
		}

		digits := query.Get("digits")
		if digits == "" {
			digits = "6"
		}
		ykargs = append(ykargs, "-d", digits)

		algorithm := strings.ToUpper(query.Get("algorithm"))
		if algorithm == "" {
			algorithm = "SHA1"
		}
		ykargs = append(ykargs, "-a", algorithm)

		if otpType == "hotp" {
			counter := query.Get("counter")
			if counter == "" {
				counter = "0"
			}
			ykargs = append(ykargs, "-c", counter)
		}

		issuer := query.Get("issuer")
		if issuer != "" {
			ykargs = append(ykargs, "-i", fmt.Sprintf(`"%s"`, issuer))
		}

		accountPath := strings.TrimPrefix(parsedUri.Path, "/")
		accountName, err := url.PathUnescape(accountPath)
		if err != nil {
			accountName = accountPath
		}

		secret := query.Get("secret")

		if issuer == "" && strings.Contains(accountName, ":") {
			parts := strings.SplitN(accountName, ":", 2)
			issuer = strings.TrimSpace(parts[0])
			accountName = strings.TrimSpace(parts[1])
			ykargs = append(ykargs, "-i", fmt.Sprintf(`"%s"`, issuer))
		}

		ykargs = append(ykargs, fmt.Sprintf(`"%s"`, accountName))
		ykargs = append(ykargs, secret)

		fullCmd := "ykman " + strings.Join(ykargs, " ")
		commands = append(commands, fullCmd)
	}

	return commands
}