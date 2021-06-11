package directory

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

	code := m.Run()

	writeReport()

	if code != 0 {
		fmt.Println("Some tests failed, but the reporting is complete.")
	}
	os.Exit(0)
}
