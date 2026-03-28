package nginxcontroller  // ← must NOT be "package main"

import (
    "fmt"
    "os"
    "os/exec"
)

func Run(action string) {
    switch action {
    case "start":
        execCommand("systemctl", "start", "nginx")
    case "stop":
        execCommand("systemctl", "stop", "nginx")
    case "restart":
        execCommand("systemctl", "restart", "nginx")
    case "reload":
        execCommand("systemctl", "reload", "nginx")
    case "status":
        execCommand("systemctl", "status", "nginx")
    case "help":
    	helpPrint()
    default:
        fmt.Println("unknown action:", action)
        fmt.Println("Usage: team18 nginx [start/stop/restart/reload/status/help]")
    }
}

func execCommand(name string, args ...string) {
    cmd := exec.Command(name, args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        fmt.Println("Error:", err)
    }
}

func helpPrint() {
	fmt.Println("Usage: team18 nginx [action]")
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("  start    Start the Nginx service")
	fmt.Println("  stop     Stop the Nginx service")
	fmt.Println("  restart  Stop and then start the Nginx service")
	fmt.Println("  reload   Reload Nginx config without dropping connections")
	fmt.Println("  status   Show the current status of the Nginx service")
	fmt.Println("  help     Show this help message")
}
