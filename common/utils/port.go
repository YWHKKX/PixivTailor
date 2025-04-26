package utils

import (
	"os/exec"
	"strconv"
	"strings"
)

func KillPortWindows(port int) error {
	cmd := exec.Command("cmd", "/c", "netstat -ano | findstr :"+strconv.Itoa(port))
	output, _ := cmd.Output()
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "LISTENING") {
			parts := strings.Fields(line)
			pid := parts[len(parts)-1]
			exec.Command("taskkill", "/F", "/PID", pid).Run()
		}
	}
	return nil
}
