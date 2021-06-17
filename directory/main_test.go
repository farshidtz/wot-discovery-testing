package directory

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"testing"
)

var serverURL string

func TestMain(m *testing.M) {
	// CLI arguments
	reportPath := flag.String("report", "report.csv", "Path to create report")
	flag.StringVar(&serverURL, "server", "", "URL of the directory service")
	flag.Parse()

	_, err := url.Parse(serverURL)
	if err != nil {
		fmt.Printf("Error parsing server URL: %s", err)
		os.Exit(1)
	}
	if serverURL == "" {
		fmt.Println("Server URL is not set!")
		os.Exit(1)
	}
	fmt.Printf("Server URL: %s\n", serverURL)

	writeReport := initReportWriter(*reportPath)

	code := m.Run()

	writeReport()

	if code != 0 {
		fmt.Println("Some tests failed, but the reporting is complete.")
	}
	os.Exit(0)
}
