package main

import (
	"fmt"
	"net/url"
	"os"
	"testing"
)

const (
	EnvURL = "URL"
)

var (
	serverURL string
)

func TestMain(m *testing.M) {
	parsedURL, err := url.Parse(os.Getenv(EnvURL))
	if err != nil {
		fmt.Printf("Error parsing server URL: %s", err)
		os.Exit(1)
	}
	serverURL = parsedURL.String()
	if serverURL == "" {
		fmt.Println("Server URL is not set!")
		os.Exit(1)
	}
	fmt.Printf("Server URL: %s\n", serverURL)

	writeReport := initReportWriter()

	if m.Run() == 1 {
		os.Exit(1)
	}

	writeReport()
	os.Exit(0)
}
