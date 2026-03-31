// proxy.go
package nginxcontroller

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

const proxyConfigFile = "/etc/team18/proxies.json"

type ProxyEntry struct {
	Project    string `json:"project"`
	ServerName string `json:"server_name"`
	Path       string `json:"path"`
	Target     string `json:"target"`
	Type       string `json:"type"` // "proxy" or "static"
}

func runProxy(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: team18 nginx proxy [list/add/add-static/remove]")
		return
	}
	switch args[0] {
	case "list":
		ProxyList()
	case "add":
		if len(args) < 4 {
			fmt.Println("Usage: team18 nginx proxy add <project> <url-path> <internal-ip>")
			return
		}
		ProxyAdd(args[1], args[2], args[3])
	case "add-static":
		if len(args) < 3 {
			fmt.Println("Usage: team18 nginx proxy add-static <project> <dist-path> [api-target]")
			return
		}
		apiTarget := ""
		if len(args) >= 4 {
			apiTarget = args[3]
		}
		ProxyAddStatic(args[1], args[2], apiTarget)
	case "remove":
		if len(args) < 2 {
			fmt.Println("Usage: team18 nginx proxy remove <project>")
			return
		}
		ProxyRemove(args[1])
	default:
		fmt.Println("unknown proxy action:", args[0])
		fmt.Println("Usage: team18 nginx proxy [list/add/add-static/remove]")
	}
}

func loadProxies() ([]ProxyEntry, error) {
	data, err := os.ReadFile(proxyConfigFile)
	if os.IsNotExist(err) {
		return []ProxyEntry{}, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []ProxyEntry
	err = json.Unmarshal(data, &entries)
	return entries, err
}

func saveProxies(entries []ProxyEntry) error {
	os.MkdirAll("/etc/team18", 0755)
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(proxyConfigFile, data, 0644)
}

func ProxyList() {
	entries, err := loadProxies()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if len(entries) == 0 {
		fmt.Println("No proxy entries found.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "PROJECT\tTYPE\tSERVER\tPATH\tTARGET")
	fmt.Fprintln(w, "-------\t----\t------\t----\t------")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", e.Project, e.Type, e.ServerName, e.Path, e.Target)
	}
	w.Flush()
}

func promptServerName() string {
	defaultServerName := getDefaultServerName()
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Server name [%s]: ", defaultServerName)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		return input
	}
	return defaultServerName
}

func ProxyAdd(project, path, target string) {
	serverName := promptServerName()
	entries, err := loadProxies()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, e := range entries {
		if e.Project == project {
			fmt.Printf("Error: project '%s' already exists\n", project)
			return
		}
	}
	entries = append(entries, ProxyEntry{
		Project:    project,
		ServerName: serverName,
		Path:       path,
		Target:     target,
		Type:       "proxy",
	})
	if err := saveProxies(entries); err != nil {
		fmt.Println("Error saving:", err)
		return
	}
	fmt.Printf("Added: %s -> %s%s -> %s\n", project, serverName, path, target)
	applyNginxConfig(entries)
}

// ProxyAddStatic registers a React static site.
// dist-path: filesystem path to the build/dist folder (e.g. /var/www/myapp/dist)
// apiTarget: optional upstream for /api/ proxy (e.g. 127.0.0.1:3000), leave "" if none
func ProxyAddStatic(project, distPath, apiTarget string) {
	serverName := promptServerName()
	entries, err := loadProxies()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, e := range entries {
		if e.Project == project {
			fmt.Printf("Error: project '%s' already exists\n", project)
			return
		}
	}
	entries = append(entries, ProxyEntry{
		Project:    project,
		ServerName: serverName,
		Path:       distPath,
		Target:     apiTarget,
		Type:       "static",
	})
	if err := saveProxies(entries); err != nil {
		fmt.Println("Error saving:", err)
		return
	}
	if apiTarget != "" {
		fmt.Printf("Added: %s -> %s (static: %s, api: %s)\n", project, serverName, distPath, apiTarget)
	} else {
		fmt.Printf("Added: %s -> %s (static: %s)\n", project, serverName, distPath)
	}
	applyNginxConfig(entries)
}

func ProxyRemove(project string) {
	entries, err := loadProxies()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	newEntries := []ProxyEntry{}
	found := false
	for _, e := range entries {
		if e.Project == project {
			found = true
		} else {
			newEntries = append(newEntries, e)
		}
	}
	if !found {
		fmt.Printf("Error: project '%s' not found\n", project)
		return
	}
	if err := saveProxies(newEntries); err != nil {
		fmt.Println("Error saving:", err)
		return
	}
	fmt.Printf("Removed: %s\n", project)
	applyNginxConfig(newEntries)
}

func applyNginxConfig(entries []ProxyEntry) {
	// Group entries by server_name
	groups := map[string][]ProxyEntry{}
	for _, e := range entries {
		groups[e.ServerName] = append(groups[e.ServerName], e)
	}

	config := ""
	for serverName, group := range groups {
		config += fmt.Sprintf("server {\n    listen 80;\n    server_name %s www.%s;\n\n", serverName, serverName)

		for _, e := range group {
			if e.Type == "static" {
				// React SPA: serve from dist folder with SPA fallback
				config += fmt.Sprintf("    root %s;\n", e.Path)
				config += "    index index.html;\n\n"

				// If an API target is provided, proxy /api/ to it
				if e.Target != "" {
					config += fmt.Sprintf(
						"    location /api/ {\n        proxy_pass http://%s/;\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n    }\n\n",
						e.Target,
					)
				}

				// Cache static assets
				config += "    location ~* \\.(?:js|css|png|jpg|svg|ico|woff2)$ {\n"
				config += "        expires 1y;\n"
				config += "        add_header Cache-Control \"public, immutable\";\n"
				config += "    }\n\n"

				// SPA fallback — must be last
				config += "    location / {\n"
				config += "        try_files $uri $uri/ /index.html;\n"
				config += "    }\n\n"

			} else {
				// Original proxy_pass behaviour
				config += fmt.Sprintf(
					"    location %s/ {\n        proxy_pass http://%s/;\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n    }\n\n",
					e.Path, e.Target,
				)
			}
		}
		config += "}\n\n"
	}

	if err := os.WriteFile("/etc/nginx/conf.d/team18.conf", []byte(config), 0644); err != nil {
		fmt.Println("Error writing nginx config:", err)
		return
	}
	execCommand("systemctl", "reload", "nginx")
	fmt.Println("Nginx reloaded.")
}
