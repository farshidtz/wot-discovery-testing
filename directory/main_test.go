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

const (
	discoveryRepoBranch = "https://raw.githubusercontent.com/w3c/wot-discovery/main"
	assertionsTemplate  = discoveryRepoBranch + "/testing/template.csv"
	assertionsManual    = discoveryRepoBranch + "/testing/manual.csv"
)

var (
	serverURL               string
	testJSONPath, testXPath bool
	templateURL, manualURL  string
	ignoreUnknownEvents     bool
)

func TestMain(m *testing.M) {
	// CLI arguments
	usage := flag.Bool("usage", false, "Print CLI usage help")
	flag.BoolVar(&testJSONPath, "testJSONPath", false, "Enable JSONPath testing")
	flag.BoolVar(&testXPath, "testXPath", false, "Enable XPath testing")
	flag.StringVar(&serverURL, "server", "", "Base URL of the directory service")
	flag.StringVar(&templateURL, "templateURL", assertionsTemplate, "URL to download assertions template")
	flag.StringVar(&manualURL, "manualURL", assertionsManual, "URL to download template for assertions that are tested manually")
	flag.BoolVar(&ignoreUnknownEvents, "ignoreUnknownEvents", false, "Ignore unknown events instead of failing the tests")
	flag.Parse()
	if *usage {
		flag.Usage()
		return
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

	writeReport := initReportWriter(templateURL, manualURL)

	code := m.Run()

	writeReport()

	if code != 0 {
		fmt.Println("Some tests failed, but the reporting is complete.")
	}
	os.Exit(0)
}
