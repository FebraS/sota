package otp

import (
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

/*
Ported and adapted from Authenticator extension
Ref:  https://github.com/Authenticator-Extension/Authenticator/blob/dev/src/models/migration.ts
*/

func ParseMigrationURI(migrationUri string) ([]string, error) {
	if !strings.HasPrefix(migrationUri, "otpauth-migration:") {
		return []string{}, nil
	}

	urlDecodedUri, err := url.QueryUnescape(migrationUri)
	if err != nil {
		return []string{}, err
	}

	parts := strings.Split(urlDecodedUri, "data=")
	if len(parts) < 2 {
		return []string{}, fmt.Errorf("invalid migration URI format")
	}
	base64Data := parts[1]

	if m := len(base64Data) % 4; m != 0 {
		base64Data += strings.Repeat("=", 4-m)
	}

	byteData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		byteData, err = base64.URLEncoding.DecodeString(base64Data)
		if err != nil {
			return []string{}, err
		}
	}

	var lines []string
	offset := 0

	defer func() {
		if r := recover(); r != nil {
		}
	}()

	for offset < len(byteData) {
		if byteData[offset] != 10 {
			break
		}

		lineLength := int(byteData[offset+1])
		secretLength := int(byteData[offset+3])
		secretBytes := byteData[offset+4 : offset+4+secretLength]

		encodedSecret := base32.StdEncoding.EncodeToString(secretBytes)
		secret := strings.ReplaceAll(encodedSecret, "=", "")

		accountLength := int(byteData[offset+4+secretLength+1])
		accountBytes := byteData[offset+4+secretLength+2 : offset+4+secretLength+2+accountLength]
		account := string(accountBytes)

		issuerLength := int(byteData[offset+4+secretLength+2+accountLength+1])
		issuerBytes := byteData[offset+4+secretLength+2+accountLength+2 : offset+4+secretLength+2+accountLength+2+issuerLength]
		issuer := string(issuerBytes)

		algorithmIndex := int(byteData[offset+4+secretLength+2+accountLength+2+issuerLength+1])
		algorithms := []string{"SHA1", "SHA1", "SHA256", "SHA512", "MD5"}
		algorithm := "SHA1"
		if algorithmIndex < len(algorithms) {
			algorithm = algorithms[algorithmIndex]
		}

		digitsIndex := int(byteData[offset+4+secretLength+2+accountLength+2+issuerLength+3])
		digitsList := []int{6, 6, 8}
		digits := 6
		if digitsIndex < len(digitsList) {
			digits = digitsList[digitsIndex]
		}

		typeIndex := int(byteData[offset+4+secretLength+2+accountLength+2+issuerLength+5])
		typeNames := []string{"totp", "hotp", "totp"}
		typeName := "totp"
		if typeIndex < len(typeNames) {
			typeName = typeNames[typeIndex]
		}

		line := fmt.Sprintf("otpauth://%s/%s?secret=%s&issuer=%s&algorithm=%s&digits=%d",
			typeName, account, secret, issuer, algorithm, digits)

		if typeName == "hotp" {
			counterOffset := offset + 4 + secretLength + 2 + accountLength + 2 + issuerLength + 7
			if counterOffset < offset+lineLength+2 {
				counter := int(byteData[counterOffset])
				line += fmt.Sprintf("&counter=%d", counter)
			}
		}

		lines = append(lines, line)
		offset += lineLength + 2
	}

	return lines, nil
}
