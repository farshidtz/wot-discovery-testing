package directory

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"
)

const (
	assertionsTemplateURL = "https://raw.githubusercontent.com/w3c/wot-discovery/main/testing/template.csv"
	templateFile          = "report/template.csv"
	reportFile            = "report/tdd-auto.csv"
	manualReportFile      = "report/tdd-manual.csv"
)

var header = []string{"ID", "Status", "Comment"}

var results map[string]result

type result struct {
	passed  []string
	failed  []string
	skipped []string
}

func initReportWriter() (commit func()) {
	err := os.MkdirAll("report", 0755)
	if err != nil {
		fmt.Printf("Error creating report directory: %s\n", err)
		os.Exit(1)
	}

	tddAssertions := loadAssertions()

	// prepare the slice so tests can append to it
	results = make(map[string]result)

	// return commit function so it can be run after all tests
	return func() {
		// Generate auto testing report
		// convert to csv records (2D slice)
		var resultsSlice [][]string
		for id, result := range results {
			resultsSlice = append(resultsSlice, resultToCSVRecord(id, result))
		}
		// sort by id
		sort.Slice(resultsSlice, func(i, j int) bool {
			return resultsSlice[i][0] < resultsSlice[j][0]
		})
		writeCSVReport(reportFile, resultsSlice)

		// Generate manual testing report
		// find untested assertions
		var manualList [][]string
		for _, id := range tddAssertions {
			if _, found := results[id]; !found {
				manualList = append(manualList, []string{id, "null", ""})
			}
		}
		writeCSVReport(manualReportFile, manualList)

		fmt.Println("\nWarning: The following tested assertions do not exist in the list of normative assertions:")
		for i := range resultsSlice {
			id := resultsSlice[i][0]
			if !inSlice(tddAssertions, id) {
				fmt.Println("-", id)
			}
		}
	}
}

// loadAssertions returns the list of assertions.
// It will read from a local file.
// If the local file is not available, it will be downloaded from the source
func loadAssertions() []string {
	if _, err := os.Stat(templateFile); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Downloading assertions from", assertionsTemplateURL)
		resp, err := http.Get(assertionsTemplateURL)
		if err != nil {
			fmt.Printf("Error downloading assertions template: %s\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		file, err := os.Create(templateFile)
		if err != nil {
			fmt.Printf("Error creating assertions template file: %s\n", err)
			os.Exit(1)
		}
		defer file.Close()

		totalBytes, err := io.Copy(file, resp.Body)
		if err != nil {
			fmt.Printf("Error copying http response to assertions template file: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Wrote %d bytes to %s\n", totalBytes, templateFile)
	}

	file, err := os.Open(templateFile)
	if err != nil {
		fmt.Printf("Error opening assertions template file: %s\n", err)
		os.Exit(1)
	}

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		fmt.Printf("Error reading assertions template file: %s\n", err)
		os.Exit(1)
	}

	var tddAssertions []string
	for _, record := range records {
		id := record[0]
		if strings.HasPrefix(id, "tdd-") {
			tddAssertions = append(tddAssertions, id)
		}
	}

	return tddAssertions
}

func insertRecord(t *testing.T, name string, assertions []string) {
	for _, a := range assertions {
		result := results[a]
		if t.Failed() {
			result.failed = append(result.failed, name)
		} else if t.Skipped() {
			result.skipped = append(result.skipped, name)
			// result.skipped = append(result.skipped, fmt.Sprintf("%s(%s)", r.name, strings.ReplaceAll(r.comments, " ", "_")))
		} else {
			result.passed = append(result.passed, name)
		}
		results[a] = result
	}

}

func writeCSVReport(filename string, input [][]string) {
	// prepend the header
	input = append([][]string{header}, input...)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating report file: %s", err)
		os.Exit(1)
	}
	writer := csv.NewWriter(file)

	for _, result := range input {
		err := writer.Write(result)
		if err != nil {
			fmt.Printf("Error writing the report: %s", err)
			os.Exit(1)
		}
	}
	writer.Flush()

	err = writer.Error()
	if err != nil {
		fmt.Printf("Error flushing the report file: %s", err)
		os.Exit(1)
	}

	file.Close()
}

func resultToCSVRecord(assertionID string, r result) []string {
	var status string
	if len(r.failed) > 0 {
		status = "fail"
	} else if len(r.skipped) > 0 {
		status = "null"
	} else {
		status = "pass"
	}

	var details []string
	if len(r.failed) > 0 {
		details = append(details, fmt.Sprint("failed:", strings.Join(r.failed, " failed:")))
	}
	if len(r.skipped) > 0 {
		details = append(details, fmt.Sprint("skipped:", strings.Join(r.skipped, " skipped:")))
	}
	if len(r.passed) > 0 {
		details = append(details, fmt.Sprint("passed:", strings.Join(r.passed, " passed:")))
	}

	return []string{
		assertionID,
		status,
		strings.Join(details, " "),
	}
}

// report at the end of tests. Execute with defer statement.
func report(t *testing.T, assertions ...string) {
	// if len(assertions) == 0 {
	// 	panic("no assertions given")
	// }

	for _, a := range assertions {
		if strings.Contains(a, ",") {
			panic("Assertion should not contain commas: " + a)
		}
		if strings.Contains(a, " ") {
			panic("Assertion should not contain spaces: " + a)
		}
	}

	insertRecord(t, t.Name(), assertions)
}

// report at the end of tests. Execute with defer statement.
func reportGroup(t *testing.T, assertionGroups ...[]string) {
	for _, ag := range assertionGroups {
		report(t, ag...)
	}
}

func inSlice(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
