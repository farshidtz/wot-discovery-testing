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

// tdd-notification
// tdd-notification-data-diff-unsupported

const (
	timeoutDuration = 5 * time.Second
	waitDuration    = time.Second
	// TD event types
	EventTypeCreate = "create"
	EventTypeUpdate = "update"
	EventTypeDelete = "delete"
)

func TestCreateEvent(t *testing.T) {

	t.Run("create event subscriber", func(t *testing.T) {

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

		select {
		case res := <-eventCh:
			t.Run("get event ID", func(t *testing.T) {
				defer report(t, "tdd-notification-sse", "tdd-notification-event-id")
				if string(res.ID) == "" {
					t.Fatal("missing event ID")
				}
			})

			t.Run("get event type", func(t *testing.T) {
				defer report(t,
					"tdd-notification-sse",
					"tdd-notification-event-types",
					"tdd-notification-filter-type",
				)
				if string(res.Event) != EventTypeCreate {
					t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeCreate)
				}
			})

			var data mapAny
			t.Run("check event data", func(t *testing.T) {
				defer report(t, "tdd-notification-data")
				err := json.Unmarshal(res.Data, &data)
				if err != nil {
					t.Fatal("unable to unmarshal the event data to TDD")
				}
			})

			t.Run("check event data td id", func(t *testing.T) {
				defer report(t, "tdd-notification-data-td-id")
				if id != data["id"] {
					t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
				}

			})
		case err := <-errCh:
			t.Run("event subscription errors", func(t *testing.T) {
				defer report(t, "tdd-notification-sse")
				t.Fatalf("unexpected error while subscribing to notification: %s", err)
			})
		case <-time.After(timeoutDuration):
			t.Run("event subscription timeout", func(t *testing.T) {
				defer report(t, "tdd-notification-sse")
				t.Fatal("timed out waiting for subscription")
			})
		}
	})

	t.Run("create event with diff subscriber", func(t *testing.T) {
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
				defer report(t, "tdd-notification-sse", "tdd-notification-event-id")
				if string(res.ID) == "" {
					t.Fatal("missing event ID")
				}
			})

			t.Run("get event type", func(t *testing.T) {
				defer report(t,
					"tdd-notification-sse",
					"tdd-notification-event-types",
					"tdd-notification-filter-type",
				)
				if string(res.Event) != EventTypeCreate {
					t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeCreate)
				}
			})

			var data mapAny
			t.Run("check event data", func(t *testing.T) {
				defer report(t, "tdd-notification-data")
				err := json.Unmarshal(res.Data, &data)
				if err != nil {
					t.Fatal("unable to unmarshal the event data to TDD")
				}
			})

			t.Run("check event data td id", func(t *testing.T) {
				defer report(t, "tdd-notification-data-td-id")
				if id != data["id"] {
					t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
				}

			})
			t.Run("check event data create full", func(t *testing.T) {
				defer report(t, "tdd-notification-data-create-full")
				// remove system-generated attributes
				delete(data, "registration")

				if !serializedEqual(td, data) {
					t.Fatalf("notification data is not same as the one created: Expected:\n%v\nRetrieved:\n%v", marshalPrettyJSON(td), marshalPrettyJSON(data))
				}

			})
		case err := <-errCh:
			t.Run("event subscription diff unsupported", func(t *testing.T) {
				defer report(t, "tdd-notification-data-diff-unsupported")
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
				defer report(t, "tdd-notification-sse", "tdd-notification-event-id")
				if string(res.ID) == "" {
					t.Fatal("missing event ID")
				}
			})

			t.Run("get event type", func(t *testing.T) {
				defer report(t,
					"tdd-notification-sse",
					"tdd-notification-event-types",
				)
				if string(res.Event) != EventTypeCreate {
					t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeCreate)
				}
			})

			var data mapAny
			t.Run("check event data", func(t *testing.T) {
				defer report(t, "tdd-notification-data")
				err := json.Unmarshal(res.Data, &data)
				if err != nil {
					t.Fatal("unable to unmarshal the event data to TDD")
				}
			})

			t.Run("check event data td id", func(t *testing.T) {
				defer report(t, "tdd-notification-data-td-id")
				if id != data["id"] {
					t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
				}

			})
		case err := <-errCh:
			t.Run("event subscription errors", func(t *testing.T) {
				defer report(t, "tdd-notification-sse")
				t.Fatalf("unexpected error while subscribing to notification: %s", err)
			})
		case <-time.After(timeoutDuration):
			t.Run("event subscription timeout", func(t *testing.T) {
				defer report(t, "tdd-notification-sse")
				t.Fatal("timed out waiting for subscription")
			})
		}
	})

	t.Run("update event subscriber", func(t *testing.T) {
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
				defer report(t, "tdd-notification-filter-type")
				t.Fatal("unexpected update event received for TD create")
			})
		case err := <-errCh:
			t.Fatalf("unexpected response to update subscription %v", err)
		case <-time.After(timeoutDuration):
			t.Log("success: did not get any update event")
		}
	})
}

func TestUpdateEvent(t *testing.T) {
	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	createThing(id, td, serverURL, t)

	t.Run("update event subscriber", func(t *testing.T) {

		// subscribe to update events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events/update", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// update an attribute
		td["title"] = "updated title for update event subscriber"
		updateThing(id, td, serverURL, t)

		select {
		case res := <-eventCh:
			t.Run("get event ID", func(t *testing.T) {
				defer report(t,
					"tdd-notification-sse",
					"tdd-notification-event-id",
				)
				if string(res.ID) == "" {
					t.Fatal("missing event ID")
				}
			})

			t.Run("get event type", func(t *testing.T) {
				defer report(t,
					"tdd-notification-sse",
					"tdd-notification-event-types",
					"tdd-notification-filter-type",
				)
				if string(res.Event) != EventTypeUpdate {
					t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeUpdate)
				}
			})

			var data mapAny
			t.Run("check event data", func(t *testing.T) {
				defer report(t, "tdd-notification-data")
				err := json.Unmarshal(res.Data, &data)
				if err != nil {
					t.Fatal("unable to unmarshal the event data to TDD")
				}
			})

			t.Run("check event data td id", func(t *testing.T) {
				defer report(t, "tdd-notification-data-td-id")
				if id != data["id"] {
					t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
				}

			})
		case err := <-errCh:
			t.Run("event subscription errors", func(t *testing.T) {
				defer report(t, "tdd-notification-sse")
				t.Fatalf("unexpected error while subscribing to notification: %s", err)
			})
		case <-time.After(timeoutDuration):
			t.Run("event subscription timeout", func(t *testing.T) {
				defer report(t, "tdd-notification-sse")
				t.Fatal("timed out waiting for data")
			})
		}
	})

	t.Run("update event with diff subscriber", func(t *testing.T) {
		// subscribe to update events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events/update?diff=true", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// update an attribute
		td["title"] = "updated title for update diff event subscriber"
		updateThing(id, td, serverURL, t)

		select {
		case res := <-eventCh:
			t.Run("get event ID", func(t *testing.T) {
				defer report(t, "tdd-notification-sse", "tdd-notification-event-id")
				if string(res.ID) == "" {
					t.Fatal("missing event ID")
				}
			})

			t.Run("get event type", func(t *testing.T) {
				defer report(t,
					"tdd-notification-sse",
					"tdd-notification-event-types",
					"tdd-notification-filter-type",
				)
				if string(res.Event) != EventTypeUpdate {
					t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeUpdate)
				}
			})

			var data mapAny
			t.Run("check event data", func(t *testing.T) {
				defer report(t, "tdd-notification-data")
				err := json.Unmarshal(res.Data, &data)
				if err != nil {
					t.Fatal("unable to unmarshal the event data to TDD")
				}
			})

			t.Run("check event data td id", func(t *testing.T) {
				defer report(t, "tdd-notification-data-td-id", "tdd-notification-data-update-id")
				if id != data["id"] {
					t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
				}

			})
			t.Run("check event data update diff", func(t *testing.T) {
				defer report(t, "tdd-notification-data-update-diff")
				// remove system-generated attributes
				delete(data, "registration")

				for key, _ := range data {
					if !(key == "id" || key == "title") {
						t.Fatalf("unexpected part in the merge patch : %s", key)
					}
				}
				if td["title"] != data["title"] {
					t.Fatalf("notification data does not reflect the changes in the title: Expected:\n%v\nRetrieved:\n%v", td["title"], data["title"])
				}

			})
		case err := <-errCh:
			t.Run("event subscription diff unsupported", func(t *testing.T) {
				defer report(t, "tdd-notification-data-diff-unsupported")
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
			t.Fatal("timed out waiting for data")
		}
	})

	t.Run("all event subscriber", func(t *testing.T) {
		// subscribe to create events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// update an attribute
		td["title"] = "updated title for all event subscriber"
		updateThing(id, td, serverURL, t)

		select {
		case res := <-eventCh:
			t.Run("get event ID", func(t *testing.T) {
				defer report(t, "tdd-notification-sse", "tdd-notification-event-id")
				if string(res.ID) == "" {
					t.Fatal("missing event ID")
				}
			})

			t.Run("get event type", func(t *testing.T) {
				defer report(t,
					"tdd-notification-sse",
					"tdd-notification-event-types",
				)
				if string(res.Event) != EventTypeUpdate {
					t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeUpdate)
				}
			})

			var data mapAny
			t.Run("check event data", func(t *testing.T) {
				defer report(t, "tdd-notification-data")
				err := json.Unmarshal(res.Data, &data)
				if err != nil {
					t.Fatal("unable to unmarshal the event data to TDD")
				}
			})

			t.Run("check event data td id", func(t *testing.T) {
				defer report(t, "tdd-notification-data-td-id")
				if id != data["id"] {
					t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
				}

			})
		case err := <-errCh:
			t.Run("event subscription errors", func(t *testing.T) {
				defer report(t, "tdd-notification-sse")
				t.Fatalf("unexpected error while subscribing to notification: %s", err)
			})
		case <-time.After(timeoutDuration):
			t.Run("event subscription timeout", func(t *testing.T) {
				defer report(t, "tdd-notification-sse")
				t.Fatal("timed out waiting for data")
			})
		}
	})

	t.Run("create event subscriber", func(t *testing.T) {
		// subscribe to create events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events/create", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// update an attribute
		td["title"] = "updated title for create event subscriber"
		updateThing(id, td, serverURL, t)

		select {
		case <-eventCh:
			t.Run("get event type", func(t *testing.T) {
				defer report(t, "tdd-notification-filter-type")
				t.Fatal("unexpected create event received for TD update")
			})
		case err := <-errCh:
			t.Fatalf("unexpected response to create subscription %v", err)
		case <-time.After(timeoutDuration):
			t.Log("success: did not get any create event")
		}
	})
}

func TestDeleteEvent(t *testing.T) {

	t.Run("delete event subscriber", func(t *testing.T) {

		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, serverURL, t)

		time.Sleep(waitDuration)

		// subscribe to delete events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events/delete", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		deleteThing(id, serverURL, t)

		select {
		case res := <-eventCh:
			t.Run("get event ID", func(t *testing.T) {
				defer report(t,
					"tdd-notification-sse",
					"tdd-notification-event-id",
				)
				if string(res.ID) == "" {
					t.Fatal("missing event ID")
				}
			})

			t.Run("get event type", func(t *testing.T) {
				defer report(t,
					"tdd-notification-sse",
					"tdd-notification-event-types",
					"tdd-notification-filter-type",
				)
				if string(res.Event) != EventTypeDelete {
					t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeDelete)
				}
			})

			var data mapAny
			t.Run("check event data", func(t *testing.T) {
				defer report(t, "tdd-notification-data")
				err := json.Unmarshal(res.Data, &data)
				if err != nil {
					t.Fatal("unable to unmarshal the event data to TDD")
				}
			})

			t.Run("check event data td id", func(t *testing.T) {
				defer report(t, "tdd-notification-data-td-id")
				if id != data["id"] {
					t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
				}

			})
		case err := <-errCh:
			t.Run("event subscription errors", func(t *testing.T) {
				defer report(t, "tdd-notification-sse")
				t.Fatalf("unexpected error while subscribing to notification: %s", err)
			})
		case <-time.After(timeoutDuration):
			t.Run("event subscription timeout", func(t *testing.T) {
				defer report(t, "tdd-notification-sse")
				t.Fatal("timed out waiting for data")
			})
		}
	})

	t.Run("delete event with diff subscriber", func(t *testing.T) {
		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, serverURL, t)

		time.Sleep(waitDuration)

		// subscribe to delete events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events/delete?diff=true", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// delete the created thing
		deleteThing(id, serverURL, t)

		select {
		case res := <-eventCh:
			t.Run("get event ID", func(t *testing.T) {
				defer report(t, "tdd-notification-sse", "tdd-notification-event-id")
				if string(res.ID) == "" {
					t.Fatal("missing event ID")
				}
			})

			t.Run("get event type", func(t *testing.T) {
				defer report(t,
					"tdd-notification-sse",
					"tdd-notification-event-types",
					"tdd-notification-filter-type",
				)
				if string(res.Event) != EventTypeDelete {
					t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeDelete)
				}
			})

			var data mapAny
			t.Run("check event data", func(t *testing.T) {
				defer report(t, "tdd-notification-data")
				err := json.Unmarshal(res.Data, &data)
				if err != nil {
					t.Fatal("unable to unmarshal the event data to TDD")
				}
			})

			t.Run("check event data td id", func(t *testing.T) {
				defer report(t, "tdd-notification-data-td-id")
				if id != data["id"] {
					t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
				}

			})
			t.Run("check event data", func(t *testing.T) {
				defer report(t, "tdd-notification-data-delete-diff")
				for key, _ := range data {
					if key != "id" {
						t.Fatalf("unexpected part in the delete notification : %s", key)
					}
				}
			})
		case err := <-errCh:
			t.Run("event subscription diff unsupported", func(t *testing.T) {
				defer report(t, "tdd-notification-data-diff-unsupported")
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
			t.Fatal("timed out waiting for data")
		}
	})

	t.Run("all event subscriber", func(t *testing.T) {
		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, serverURL, t)

		time.Sleep(waitDuration)

		// subscribe to delete events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// delete the created thing
		deleteThing(id, serverURL, t)

		select {
		case res := <-eventCh:
			t.Run("get event ID", func(t *testing.T) {
				defer report(t, "tdd-notification-sse", "tdd-notification-event-id")
				if string(res.ID) == "" {
					t.Fatal("missing event ID")
				}
			})

			t.Run("get event type", func(t *testing.T) {
				defer report(t,
					"tdd-notification-sse",
					"tdd-notification-event-types",
				)
				if string(res.Event) != EventTypeDelete {
					t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeDelete)
				}
			})

			var data mapAny
			t.Run("check event data", func(t *testing.T) {
				defer report(t, "tdd-notification-data")
				err := json.Unmarshal(res.Data, &data)
				if err != nil {
					t.Fatal("unable to unmarshal the event data to TDD")
				}
			})

			t.Run("check event data td id", func(t *testing.T) {
				defer report(t, "tdd-notification-data-td-id")
				if id != data["id"] {
					t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
				}

			})
		case err := <-errCh:
			t.Run("event subscription errors", func(t *testing.T) {
				defer report(t, "tdd-notification-sse")
				t.Fatalf("unexpected error while subscribing to notification: %s", err)
			})
		case <-time.After(timeoutDuration):
			t.Run("event subscription timeout", func(t *testing.T) {
				defer report(t, "tdd-notification-sse")
				t.Fatal("timed out waiting for data")
			})
		}
	})

	t.Run("create event subscriber", func(t *testing.T) {
		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, serverURL, t)

		time.Sleep(waitDuration)

		// subscribe to delete events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events/create", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// delete the created thing
		deleteThing(id, serverURL, t)

		select {
		case <-eventCh:
			t.Run("get event type", func(t *testing.T) {
				defer report(t, "tdd-notification-filter-type")
				t.Fatal("unexpected create event received for TD delete")
			})
		case err := <-errCh:
			t.Fatalf("unexpected response to create subscription %v", err)
		case <-time.After(timeoutDuration):
			t.Log("success: did not get any create event")
		}
	})
}
