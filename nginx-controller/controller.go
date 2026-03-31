package nginxcontroller  // ← must NOT be "package main"

import (
    "fmt"
    "os"
    "os/exec"
)

func Run(args []string) {
	switch args[0] {
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
	case "proxy":
		runProxy(args[1:]) // 交給 proxy.go 處理
	case "help":
		helpPrint()
	case "config":
    		if len(args) < 2 {
        		ConfigShow()
        		return
    		}
    		if len(args) < 3 {
        		fmt.Println("Usage: team18 nginx config set <key> <value>")
        		return
    		}
    		ConfigSet(args[1], args[2])
	default:
		fmt.Println("unknown action:", args[0])
		fmt.Println("Usage: team18 nginx [start/stop/restart/reload/status/proxy/help]")
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
