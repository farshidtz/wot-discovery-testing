package directory

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"testing"
)

const outputFile = "report.csv"

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

func initReportWriter() (commit func()) {
	results = make(map[string]result)
	// csv header
	header := []string{"AssertionID", "Status", "Coverage", "Details"}

	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating report file: %s", err)
		os.Exit(1)
	}

	// write to both stdout and file
	writer := csv.NewWriter(io.MultiWriter(os.Stdout, file))

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
		sort.Slice(sortedResults, func(i, j int) bool {
			return sortedResults[i][0] < sortedResults[j][0]
		})

		fmt.Println("==================== REPORT ====================")
		for _, result := range sortedResults {
			err := writer.Write(result)
			if err != nil {
				fmt.Printf("Error writing the report: %s", err)
				os.Exit(1)
			}
		}
		writer.Flush()
		fmt.Println("================================================")

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
	} else if len(r.passed) > 0 {
		status = "passed"
	} else {
		status = "skipped"
	}

	// calculate ratio of passed tests
	var coverage float64
	if len(r.passed)+len(r.failed)+len(r.skipped) == 0 {
		coverage = 0
	} else {
		coverage = float64(len(r.passed)) / float64(len(r.passed)+len(r.failed)+len(r.skipped))
	}

	var details []string
	if len(r.passed) > 0 {
		// details = append(details, fmt.Sprint("passed:", strings.Join(r.passed, " ")))
		details = append(details, fmt.Sprint("passed:", len(r.passed)))
	}
	if len(r.failed) > 0 {
		details = append(details, fmt.Sprint("failed:", strings.Join(r.failed, " ")))
	}
	if len(r.skipped) > 0 {
		details = append(details, fmt.Sprint("skipped:", strings.Join(r.skipped, " ")))
	}

	return []string{
		assertionID,
		status,
		fmt.Sprintf("%.0f%%", coverage*100),
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
	r.comments = fmt.Sprintf(format, messages...)
	t.Fatal(r.comments)
}

func skip(t *testing.T, r *record, format string, messages ...interface{}) {
	r.comments = fmt.Sprintf(format, messages...)
	t.Skip(r.comments)
}
