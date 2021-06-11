package directory

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

const outputFile = "report.csv"

var records [][]string

func initReportWriter() (commit func()) {
	// csv header
	header := []string{"ID", "Status", "Related Assertions from Spec", "Comments"}

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
	return func() {
		fmt.Println("==================== REPORT ====================")
		for _, record := range records {
			err := writer.Write(record)
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

}

func appendRecord(id, status, assertions, comments string) {
	records = append(records, []string{id, status, assertions, comments})
}

func writeTestResult(assertions, comments string, t *testing.T) {
	id := t.Name()
	if strings.Contains(id, ",") {
		t.Fatalf("ID cell should not contain commas.")
	}
	if strings.Contains(assertions, ",") {
		t.Fatalf("Assertions cell should not contain commas. Use space as separator.")
	}
	if t.Failed() {
		appendRecord(id, "failed", assertions, comments)
	} else if t.Skipped() {
		appendRecord(id, "skipped", assertions, comments)
	} else {
		appendRecord(id, "passed", assertions, comments)
	}
}

type record struct {
	id         string
	status     string
	assertions []string
	comments   string
}

func report(t *testing.T, r *record) {
	if r == nil {
		r = &record{}
	}
	if r.id != "" {
		panic("id should not be set explicitly: " + r.id)
	}
	// take id from test name
	r.id = t.Name()
	if strings.Contains(r.id, ",") {
		panic("ID cell should not contain commas: " + r.id)
	}

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
	if t.Failed() {
		r.status = "failed"
	} else if t.Skipped() {
		r.status = "skipped"
	} else {
		r.status = "passed"
	}

	appendRecord(r.id, r.status, strings.Join(r.assertions, " "), r.comments)
}

func fatal(t *testing.T, r *record, format string, messages ...interface{}) {
	r.comments = fmt.Sprintf(format, messages...)
	t.Fatal(r.comments)
}

func skip(t *testing.T, r *record, format string, messages ...interface{}) {
	r.comments = fmt.Sprintf(format, messages...)
	t.Skip(r.comments)
}
