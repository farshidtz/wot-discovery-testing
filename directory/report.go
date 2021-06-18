package directory

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

var results map[string]result

type record struct {
	name       string
	status     string
	assertions []string
	comments   string
}

type result struct {
	passed  []string
	failed  []string
	skipped []string
}

func initReportWriter(path string) (commit func()) {
	results = make(map[string]result)
	// csv header
	header := []string{"AssertionID", "Status", "Details"}

	file, err := os.Create(path)
	if err != nil {
		fmt.Printf("Error creating report file: %s", err)
		os.Exit(1)
	}
	writer := csv.NewWriter(file)

	err = writer.Write(header)
	if err != nil {
		fmt.Printf("Error writing header to report file: %s", err)
		os.Exit(1)
	}

	// return commit function and run after all tests
	var sortedResults [][]string
	commit = func() {
		// sort by id
		for id, result := range results {
			sortedResults = append(sortedResults, resultToCSV(id, result))
		}

		// insert unchecked assertions
		for _, id := range tddAssertions {
			if _, found := results[id]; !found {
				sortedResults = append(sortedResults, []string{id, "skipped", "untested"})
			}
		}

		sort.Slice(sortedResults, func(i, j int) bool {
			return sortedResults[i][0] < sortedResults[j][0]
		})

		fmt.Println("Writing report to", file.Name())
		for _, result := range sortedResults {
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

	return commit
}

func ingestRecord(t *testing.T, r record) {
	for _, a := range r.assertions {
		result := results[a]
		if t.Failed() {
			result.failed = append(result.failed, r.name)
		} else if t.Skipped() {
			result.skipped = append(result.skipped, r.name)
			// result.skipped = append(result.skipped, fmt.Sprintf("%s(%s)", r.name, strings.ReplaceAll(r.comments, " ", "_")))
		} else {
			result.passed = append(result.passed, r.name)
		}
		results[a] = result
	}

}

func resultToCSV(assertionID string, r result) []string {
	var status string
	if len(r.failed) > 0 {
		status = "failed"
	} else if len(r.skipped) > 0 {
		status = "skipped"
	} else {
		status = "passed"
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
func report(t *testing.T, r *record) {
	if r == nil {
		r = &record{}
	}

	r.name = t.Name()

	for _, a := range r.assertions {
		if strings.Contains(a, ",") {
			panic("Assertion should not contain commas: " + a)
		}
		if strings.Contains(a, " ") {
			panic("Assertion should not contain spaces: " + a)
		}
	}

	if r.status != "" {
		panic("status should not be set explicitly: " + r.status)
	}

	ingestRecord(t, *r)
}

func fatal(t *testing.T, r *record, format string, messages ...interface{}) {
	t.Helper()
	r.comments = fmt.Sprintf(format, messages...)
	t.Fatal(r.comments)
}

func skip(t *testing.T, r *record, format string, messages ...interface{}) {
	t.Helper()
	r.comments = fmt.Sprintf(format, messages...)
	t.Skip(r.comments)
}
