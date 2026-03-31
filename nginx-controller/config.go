package nginxcontroller

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const globalConfigFile = "/etc/team18/config.json"

type GlobalConfig struct {
	DefaultServerName string `json:"default_server_name"`
}

func loadGlobalConfig() (GlobalConfig, error) {
	data, err := os.ReadFile(globalConfigFile)
	if os.IsNotExist(err) {
		return GlobalConfig{}, nil
	}
	if err != nil {
		return GlobalConfig{}, err
	}
	var cfg GlobalConfig
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

func saveGlobalConfig(cfg GlobalConfig) error {
	os.MkdirAll("/etc/team18", 0755)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(globalConfigFile, data, 0644)
}

// 取得預設 server name，若未設定則詢問使用者並儲存
func getDefaultServerName() string {
	cfg, err := loadGlobalConfig()
	if err != nil {
		fmt.Println("Warning: could not load config:", err)
	}

	if cfg.DefaultServerName != "" {
		return cfg.DefaultServerName
	}

	// 第一次使用，詢問並儲存
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("No default server name set. Enter your domain (e.g. example.com): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		fmt.Println("Error: server name cannot be empty")
		os.Exit(1)
	}

	cfg.DefaultServerName = input
	if err := saveGlobalConfig(cfg); err != nil {
		fmt.Println("Warning: could not save config:", err)
	}
	fmt.Printf("Default server name set to: %s\n", input)
	return input
}

// 讓使用者可以直接更新預設 server name
func ConfigSet(key, value string) {
	cfg, err := loadGlobalConfig()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	switch key {
	case "default_server_name":
		cfg.DefaultServerName = value
		if err := saveGlobalConfig(cfg); err != nil {
			fmt.Println("Error saving:", err)
			return
		}
		fmt.Printf("Set default_server_name = %s\n", value)
	default:
		fmt.Println("unknown config key:", key)
	}
}

func ConfigShow() {
	cfg, err := loadGlobalConfig()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("default_server_name = %s\n", cfg.DefaultServerName)
}
