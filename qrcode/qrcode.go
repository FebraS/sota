package qrcode

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	go_qrcode "github.com/skip2/go-qrcode"
)

func ReadURI(imagePath string) string {
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		color.Red("Error: File '%s' not found.", imagePath)
		return ""
	}

	scriptPath := "scripts/extract.py"
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		color.Red("Error: Helper script '%s' not found.", scriptPath)
		return ""
	}

	cmd := exec.Command("python", scriptPath, imagePath)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		color.Red("Failed to decode QR: %s", strings.TrimSpace(stderr.String()))
		return ""
	}

	uri := strings.TrimSpace(out.String())
	if strings.HasPrefix(uri, "otpauth-migration://") || strings.HasPrefix(uri, "otpauth://") {
		color.Green("Successfully read QR Code!")
		return uri
	}

	color.Yellow("Warning: QR code found, but invalid URI format.")
	return ""
}

func Generate(uri, outputPath string) bool {
	err := go_qrcode.WriteFile(uri, go_qrcode.High, 512, outputPath)
	if err != nil {
		color.Red("Error generating QR: %v", err)
		return false
	}
	magentaPrinter := color.New(color.FgHiMagenta).SprintFunc()
	fmt.Printf("QR Code generated and saved to: %s\n", magentaPrinter(outputPath))
	return true
}
