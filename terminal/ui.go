package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/fatih/color"
)

func Clear() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func Banner() {
	asciiArt := `                                                         
  ▄▄▄▄▄                 
 ██▀▀▀▀█▄       █▄      
 ▀██▄  ▄▀      ▄██▄     
   ▀██▄▄  ▄███▄ ██ ▄▀▀█▄
 ▄   ▀██▄ ██ ██ ██ ▄█▀██
 ▀██████▀▄▀███▀▄██▄▀█▄██                                                                                                                                                                             
v2.0.0

`
	color.HiYellow(asciiArt)
	color.HiBlue("Simple One Time Authenticator")
	color.HiBlue("(c)2026, febras")
	fmt.Println()
}
