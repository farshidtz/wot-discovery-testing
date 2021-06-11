package directory

import (
	"testing"
)

func TestCreateEvent(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("", "TODO", t)
	})
	t.SkipNow()
}

func TestUpdateEvent(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("", "TODO", t)
	})
	t.SkipNow()
}

func TestDeleteEvent(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("", "TODO", t)
	})
	t.SkipNow()
}
