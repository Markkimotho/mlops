package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	base := env("MLAIOPS_URL", "http://localhost:8080")
	client := &http.Client{Timeout: 15 * time.Second}
	var method, path string
	var body any
	switch strings.Join(os.Args[1:], " ") {
	case "project list":
		method, path = http.MethodGet, "/api/v1/projects"
	case "pipeline list":
		method, path = http.MethodGet, "/api/v1/pipelines/runs"
	case "model list":
		method, path = http.MethodGet, "/api/v1/models"
	case "agent list":
		method, path = http.MethodGet, "/api/v1/agents"
	case "tool list":
		method, path = http.MethodGet, "/api/v1/tools"
	case "connection list":
		method, path = http.MethodGet, "/api/v1/connections"
	case "audit list":
		method, path = http.MethodGet, "/api/v1/audit"
	default:
		if len(os.Args) == 4 && os.Args[1] == "pipeline" && os.Args[2] == "submit" {
			method, path, body = http.MethodPost, "/api/v1/pipelines/submit", map[string]string{"project_id": os.Args[3], "name": "training-pipeline"}
		} else {
			usage()
			os.Exit(2)
		}
	}
	var reader io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	}
	request, err := http.NewRequest(method, base+path, reader)
	fatal(err)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-MLAIOps-Actor", env("USER", "cli"))
	response, err := client.Do(request)
	fatal(err)
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	fatal(err)
	if response.StatusCode >= 300 {
		fmt.Fprintln(os.Stderr, string(raw))
		os.Exit(1)
	}
	var value any
	if json.Unmarshal(raw, &value) == nil {
		formatted, _ := json.MarshalIndent(value, "", "  ")
		fmt.Println(string(formatted))
	} else {
		fmt.Println(string(raw))
	}
}
func usage() {
	fmt.Fprintln(os.Stderr, "usage: mlaiops <project|pipeline|model|agent|tool|connection|audit> list | mlaiops pipeline submit <project-id>")
}
func fatal(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
