package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

const outputFile = "report.csv"

var report [][]string

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
		for _, record := range report {
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

func writeReportLine(id, status, assertions, comments string) {
	report = append(report, []string{id, status, assertions, comments})
}

func writeTestResult(id, assertions, comments string, t *testing.T) {
	if strings.Contains(id, ",") {
		t.Fatalf("ID cell should not contain commas.")
	}
	if strings.Contains(assertions, ",") {
		t.Fatalf("Assertions cell should not contain commas. Use space as separator.")
	}
	if t.Failed() {
		writeReportLine(id, "failed", assertions, comments)
	} else if t.Skipped() {
		writeReportLine(id, "skipped", assertions, comments)
	} else {
		writeReportLine(id, "passed", assertions, comments)
	}
}
