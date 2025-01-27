package main

import (
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

type PageData struct {
	EnvVars    map[string]string
	ConfigVars map[string]string
	Hostname   string
	IPAddress  string
}

func main() {
	log.Println("Starting the server on port 8080")
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil)) // Log fatal errors if server fails
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request from %s %s\n", r.RemoteAddr, r.Method)

	envVars := os.Environ()
	envMap := make(map[string]string)
	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	log.Println("Fetched environment variables")

	configVars := readConfigFiles("/etc/config")
	log.Printf("Fetched %d config variables\n", len(configVars))

	host, ip := getHostIP()
	log.Printf("Fetched hostname: %s, IP address: %s\n", host, ip)

	data := PageData{
		EnvVars:    envMap,
		ConfigVars: configVars,
		Hostname:   host,
		IPAddress:  ip,
	}

	renderTemplate(w, data)
}

func renderTemplate(w http.ResponseWriter, data PageData) {
	tmpl := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Environment and Config Variables</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				margin: 2em;
				background-color: #f4f4f9;
			}
			h1 {
				color: #333;
			}
			.container {
				background: #fff;
				padding: 20px;
				border-radius: 8px;
				box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
			}
			.value {
				color: #007BFF;
				font-weight: bold;
			}
			.env-table {
				width: 100%;
				border-collapse: collapse;
				margin-top: 1em;
			}
			.env-table th, .env-table td {
				border: 1px solid #ddd;
				padding: 8px;
			}
			.env-table th {
				background-color: #f4f4f9;
				text-align: left;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>Environment and Config Variable Viewer</h1>
			<p><strong>Hostname:</strong> <span class="value">{{.Hostname}}</span></p>
			<p><strong>IP Address:</strong> <span class="value">{{.IPAddress}}</span></p>

			<h2>Environment Variables</h2>
			<table class="env-table">
				<tr>
					<th>Key</th>
					<th>Value</th>
				</tr>
				{{range $key, $value := .EnvVars}}
				<tr>
					<td>{{$key}}</td>
					<td>{{$value}}</td>
				</tr>
				{{end}}
			</table>

			<h2>Config Variables (from ConfigMap Volume)</h2>
			<table class="env-table">
				<tr>
					<th>Key</th>
					<th>Value</th>
				</tr>
				{{range $key, $value := .ConfigVars}}
				<tr>
					<td>{{$key}}</td>
					<td>{{$value}}</td>
				</tr>
				{{end}}
			</table>
		</div>
	</body>
	</html>
	`
	t := template.Must(template.New("webpage").Parse(tmpl))
	err := t.Execute(w, data)
	if err != nil {
		log.Printf("Error rendering template: %v\n", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func readConfigFiles(configDir string) map[string]string {
	configVars := make(map[string]string)

	log.Printf("Reading config files from directory: %s\n", configDir)

	files, err := os.ReadDir(configDir)
	if err != nil {
		log.Printf("Error reading config directory: %v\n", err)
		return configVars // Return empty map if error occurs
	}

	for _, file := range files {
		if !file.IsDir() {
			content, err := os.ReadFile(configDir + "/" + file.Name())
			if err == nil {
				configVars[file.Name()] = string(content)
				log.Printf("Loaded config file: %s\n", file.Name())
			} else {
				log.Printf("Error reading config file %s: %v\n", file.Name(), err)
			}
		}
	}
	return configVars
}

func getHostIP() (string, string) {
	host, err := os.Hostname()
	if err != nil {
		log.Printf("Error getting hostname: %v\n", err)
		host = "Unknown"
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("Error getting network interface addresses: %v\n", err)
		return host, "Unknown"
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				log.Printf("Found IP address: %s\n", ipNet.IP.String())
				return host, ipNet.IP.String()
			}
		}
	}

	return host, "Unknown"
}
