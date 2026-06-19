package clip

import (
	"fmt"
	"os/exec"
	"strings"
)

func Read() (string, error) {
	if out, err := exec.Command("xclip", "-o", "-selection", "clipboard").Output(); err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	if out, err := exec.Command("xsel", "-o", "-b").Output(); err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	return "", fmt.Errorf("clipboard requires xclip or xsel; install with: apt install xclip")
}
