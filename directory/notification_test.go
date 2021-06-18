package directory

import (
	"testing"
)

func TestCreateEvent(t *testing.T) {
	defer report(t, &record{assertions: []string{
		"tdd-notification-sse",
		"tdd-notification-event-id",
		"tdd-notification-event-types",
		"tdd-notification-filter-type",
		"tdd-notification-data",
		"tdd-notification-data-tdid",
		"tdd-notification-data-create-full",
		"tdd-notification-data-diff-unsupported",
	}})
	t.Skip("TODO")
}

func TestUpdateEvent(t *testing.T) {
	defer report(t, &record{assertions: []string{
		"tdd-notification-sse",
		"tdd-notification-event-id",
		"tdd-notification-event-types",
		"tdd-notification-filter-type",
		"tdd-notification-data",
		"tdd-notification-data-tdid",
		"tdd-notification-data-update-diff",
		"tdd-notification-data-update-id",
		"tdd-notification-data-diff-unsupported",
	}})
	t.Skip("TODO")
}

func TestDeleteEvent(t *testing.T) {
	defer report(t, &record{assertions: []string{
		"tdd-notification-sse",
		"tdd-notification-event-id",
		"tdd-notification-event-types",
		"tdd-notification-filter-type",
		"tdd-notification-data",
		"tdd-notification-data-tdid",
		"tdd-notification-data-delete-diff",
		"tdd-notification-data-diff-unsupported",
	}})
	t.Skip("TODO")
}
