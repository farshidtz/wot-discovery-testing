package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"testing"
)

var report [][]string

func initReportWriter() func() {
	header := []string{"ID", "Status", "Comments"}

	writer := csv.NewWriter(os.Stdout)

	err := writer.Write(header)
	if err != nil {
		fmt.Printf("Error writing header to report file: %s", err)
		os.Exit(1)
	}

	return func() {
		fmt.Println("========== REPORT ==========")
		for _, record := range report {
			err := writer.Write(record)
			if err != nil {
				fmt.Printf("Error writing the report: %s", err)
				os.Exit(1)
			}
		}

		writer.Flush()
		fmt.Println("============================")
		err = writer.Error()
		if err != nil {
			fmt.Printf("Error flushing the report file: %s", err)
			os.Exit(1)
		}
	}

}

func writeReportLine(id, status, comment string) {
	report = append(report, []string{id, status, comment})
}

func writeTestResult(id, comment string, t *testing.T) {
	if t.Failed() {
		writeReportLine(id, "fail", comment)
	} else {
		writeReportLine(id, "pass", comment)
	}
}
