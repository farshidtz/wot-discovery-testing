package main

import (
	"testing"
)

func TestCreateEvent(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("notification-create", "TODO", t)
	})
	t.SkipNow()
}

func TestUpdateEvent(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("notification-update", "TODO", t)
	})
	t.SkipNow()
}

func TestDeleteEvent(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("notification-delete", "TODO", t)
	})
	t.SkipNow()
}
