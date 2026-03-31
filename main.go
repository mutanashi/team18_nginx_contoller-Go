package main

import (
	"fmt"
	"os"
	nginxcontroller "team18/nginx-controller"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: team18 [system] -h")
		return
	}
	switch os.Args[1] {
	case "nginx":
		if len(os.Args) < 3 {
			fmt.Println("Usage: team18 nginx [start/stop/restart/reload/status/proxy/help]")
			return
		}
		nginxcontroller.Run(os.Args[2:]) // 傳入剩餘所有參數
	default:
		fmt.Println("unknown system:", os.Args[1])
	}
}
