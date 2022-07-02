package directory

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"testing"
)

const (
	MediaTypeJSON             = "application/json"
	MediaTypeJSONLD           = "application/ld+json"
	MediaTypeThingDescription = "application/td+json"
	MediaTypeMergePatch       = "application/merge-patch+json"
)

var (
	serverURL               string
	testJSONPath, testXPath bool
)

func TestMain(m *testing.M) {
	// CLI arguments
	reportPath := flag.String("report", "", "Path to create report")
	flag.BoolVar(&testJSONPath, "testJSONPath", false, "Enable JSONPath testing")
	flag.BoolVar(&testXPath, "testXPath", false, "Enable XPath testing")
	flag.StringVar(&serverURL, "server", "", "Base URL of the directory service")
	flag.Parse()

	if *reportPath != "" {
		fmt.Printf("Bad input. Report path is now hardcoded to %s\n", reportFile)
		os.Exit(1)
	}

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

	writeReport := initReportWriter()

	code := m.Run()

	writeReport()

	if code != 0 {
		fmt.Println("Some tests failed, but the reporting is complete.")
	}
	os.Exit(0)
}
