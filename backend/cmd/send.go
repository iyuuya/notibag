package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Config struct {
	Host string `json:"host"`
}

type CreateNotificationRequest struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

func loadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(homeDir, ".notibag", "config.json")

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Host: "http://localhost:8080"}, nil
		}
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	var host = flag.String("host", config.Host, "Server host URL")
	var title = flag.String("title", "", "Notification title")
	var message = flag.String("message", "", "Notification message")
	flag.Parse()

	if *title == "" || *message == "" {
		fmt.Println("Usage: send -title <title> -message <message> [-host <host>]")
		os.Exit(1)
	}

	req := CreateNotificationRequest{
		Title:   *title,
		Message: *message,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	url := *host + "/api/notifications"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	fmt.Println("Notification sent successfully")
}