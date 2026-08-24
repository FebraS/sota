package otp

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

type TOTP struct {
	Secret string
}

func (t *TOTP) Now() string {
	passcode, err := totp.GenerateCode(t.Secret, time.Now())
	if err != nil {
		return "ERROR"
	}
	return passcode
}

type Account struct {
	TotpObj *TOTP
	Name    string
	RawURI  string
}

func ParseURI(uri string) (*TOTP, string, error) {
	if !strings.HasPrefix(uri, "otpauth://") {
		return nil, "", errors.New("invalid URI protocol")
	}

	parsedUrl, err := url.Parse(uri)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse URI: %v", err)
	}

	query := parsedUrl.Query()
	secret := query.Get("secret")
	if secret == "" {
		return nil, "", errors.New("missing secret in URI")
	}

	issuer := query.Get("issuer")
	accountPath := strings.TrimPrefix(parsedUrl.Path, "/")
	
	if unquotedPath, err := url.PathUnescape(accountPath); err == nil {
		accountPath = unquotedPath
	}

	var account string
	if strings.Contains(accountPath, ":") {
		pathParts := strings.SplitN(accountPath, ":", 2)
		if issuer == "" {
			issuer = strings.TrimSpace(pathParts[0])
		}
		account = strings.TrimSpace(pathParts[1])
	} else {
		account = accountPath
	}

	if issuer == "" {
		issuer = "Unknown Issuer"
	}

	accountName := fmt.Sprintf("%s: %s", issuer, account)
	return &TOTP{Secret: secret}, accountName, nil
}