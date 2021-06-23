package directory

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/r3labs/sse/v2"
	uuid "github.com/satori/go.uuid"
)

const (
	timeoutDuration = 5 * time.Second
	waitDuration    = time.Second
	// TD event types
	EventTypeCreate = "create"
	EventTypeUpdate = "update"
	EventTypeDelete = "delete"
)

type dummyLock struct {
}

func (l dummyLock) Lock() {

}

func (l dummyLock) Unlock() {

}

func TestCreateEvent(t *testing.T) {

	var seqMutex dummyLock

	seq := func() func() {
		seqMutex.Lock()
		return func() {
			seqMutex.Unlock()
		}
	}

	t.Run("create event subscriber", func(t *testing.T) {
		defer seq()()

		// subscribe to create events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events/create", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, serverURL, t)

		/*defer report(t, &record{assertions: []string{
			"tdd-notification-sse",
			//"tdd-notification-event-id",
			"tdd-notification-event-types",
			"tdd-notification-filter-type",
			"tdd-notification-data",
			"tdd-notification-data-tdid",
			//"tdd-notification-data-create-full",
			//"tdd-notification-data-diff-unsupported",
		}})*/
		select {
		case res := <-eventCh:
			t.Run("get event ID", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-sse", "tdd-notification-event-id"}})
				if string(res.ID) == "" {
					t.Fatal("missing event ID")
				}
			})

			t.Run("get event type", func(t *testing.T) {
				defer report(t, &record{assertions: []string{
					"tdd-notification-sse",
					"tdd-notification-event-types",
					"tdd-notification-filter-type",
				}})
				if string(res.Event) != EventTypeCreate {
					t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeCreate)
				}
			})

			var data mapAny
			t.Run("check event data", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-data"}})
				err := json.Unmarshal(res.Data, &data)
				if err != nil {
					t.Fatal("unable to unmarshal the event data to TDD")
				}
			})

			t.Run("check event data tdid", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-data-tdid"}})
				if id != data["id"] {
					t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
				}

			})
		case err := <-errCh:
			t.Run("event subscription errors", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-sse"}})
				t.Fatalf("unexpected error while subscribing to notification: %s", err)
			})
		case <-time.After(timeoutDuration):
			t.Run("event subscription timeout", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-sse"}})
				t.Fatal("timed out waiting for subscription")
			})
		}
	})

	t.Run("create event with diff subscriber", func(t *testing.T) {
		defer seq()()
		// subscribe to create events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events/create?diff=true", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, serverURL, t)

		select {
		case res := <-eventCh:
			t.Run("get event ID", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-sse", "tdd-notification-event-id"}})
				if string(res.ID) == "" {
					t.Fatal("missing event ID")
				}
			})

			t.Run("get event type", func(t *testing.T) {
				defer report(t, &record{assertions: []string{
					"tdd-notification-sse",
					"tdd-notification-event-types",
					"tdd-notification-filter-type",
				}})
				if string(res.Event) != EventTypeCreate {
					t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeCreate)
				}
			})

			var data mapAny
			t.Run("check event data", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-data"}})
				err := json.Unmarshal(res.Data, &data)
				if err != nil {
					t.Fatal("unable to unmarshal the event data to TDD")
				}
			})

			t.Run("check event data tdid", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-data-tdid"}})
				if id != data["id"] {
					t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
				}

			})
			t.Run("check event data create full", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-data-create-full"}})
				// remove system-generated attributes
				delete(data, "registration")

				if !serializedEqual(td, data) {
					t.Fatalf("notification data is not same as the one created: Expected:\n%v\nRetrieved:\n%v", marshalPrettyJSON(td), marshalPrettyJSON(data))
				}

			})
		case err := <-errCh:
			t.Run("event subscription diff unsupported", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-data-diff-unsupported"}})
				var httpErr *httpError
				if errors.As(err, &httpErr) {
					if httpErr.code != http.StatusNotImplemented {
						t.Fatalf("unexpected response code: %d", httpErr.code)
					}
				} else {
					t.Fatalf("unexpected error while subscribing to notification: %s", err)
				}
			})
		case <-time.After(timeoutDuration):
			t.Fatal("timed out")
		}
	})

	t.Run("all event subscriber", func(t *testing.T) {
		defer seq()()
		// subscribe to create events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, serverURL, t)

		select {
		case res := <-eventCh:
			t.Run("get event ID", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-sse", "tdd-notification-event-id"}})
				if string(res.ID) == "" {
					t.Fatal("missing event ID")
				}
			})

			t.Run("get event type", func(t *testing.T) {
				defer report(t, &record{assertions: []string{
					"tdd-notification-sse",
					"tdd-notification-event-types",
				}})
				if string(res.Event) != EventTypeCreate {
					t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeCreate)
				}
			})

			var data mapAny
			t.Run("check event data", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-data"}})
				err := json.Unmarshal(res.Data, &data)
				if err != nil {
					t.Fatal("unable to unmarshal the event data to TDD")
				}
			})

			t.Run("check event data tdid", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-data-tdid"}})
				if id != data["id"] {
					t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
				}

			})
		case err := <-errCh:
			t.Run("event subscription errors", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-sse"}})
				t.Fatalf("unexpected error while subscribing to notification: %s", err)
			})
		case <-time.After(timeoutDuration):
			t.Run("event subscription timeout", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-sse"}})
				t.Fatal("timed out waiting for subscription")
			})
		}
	})

	t.Run("update event subscriber", func(t *testing.T) {
		defer seq()()
		// subscribe to create events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events/update", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)
		time.Sleep(waitDuration)

		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, serverURL, t)

		select {
		case <-eventCh:
			t.Run("get event type", func(t *testing.T) {
				defer report(t, &record{assertions: []string{"tdd-notification-filter-type"}})
				t.Fatal("unexpected update event received for TD create")
			})
		case err := <-errCh:
			t.Fatalf("unexpected response to update subscription %v", err)
		case <-time.After(timeoutDuration):
			t.Log("success: did not get any update event")
		}
	})

	t.Log("done")

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
