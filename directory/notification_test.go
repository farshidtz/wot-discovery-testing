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
	EventTypeCreate = "thing_created"
	EventTypeUpdate = "thing_updated"
	EventTypeDelete = "thing_deleted"
)

func TestCreateEvent(t *testing.T) {

	t.Run("create event subscriber", func(t *testing.T) {

		// subscribe to create events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events/"+EventTypeCreate, eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, serverURL, t)

		for {
			select {
			case res := <-eventCh:
				if ignoreUnknownEvents && string(res.Event) != EventTypeCreate {
					t.Logf("Unknown event type: '%s'", res.Event)
					continue
				} else {
					t.Run("get event type", func(t *testing.T) {
						defer report(t,
							"tdd-notification",
							"tdd-notification-sse",
							"tdd-notification-event-types",
							"tdd-notification-filter-type",
						)
						if string(res.Event) != EventTypeCreate {
							t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeCreate)
						}
					})

					t.Run("get event ID", func(t *testing.T) {
						defer report(t, "tdd-notification-sse", "tdd-notification-event-id")
						if string(res.ID) == "" {
							t.Fatal("missing event ID")
						}
					})

					var data mapAny
					t.Run("check event data", func(t *testing.T) {
						defer report(t, "tdd-notification-data")
						err := json.Unmarshal(res.Data, &data)
						if err != nil {
							t.Fatalf("unable to unmarshal the event data: %s", res.Data)
						}
					})

					t.Run("check event data td id", func(t *testing.T) {
						defer report(t, "tdd-notification-data-td-id")
						if id != data["id"] {
							t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
						}

					})
				}
				return
			case err := <-errCh:
				t.Fatalf("unexpected error while subscribing to notification: %s", err)

			case <-time.After(timeoutDuration):
				t.Fatal("timed out waiting for subscription")
			}

		}
	})

	t.Run("create event with diff subscriber", func(t *testing.T) {
		// subscribe to create events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events/"+EventTypeCreate+"?diff=true", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, serverURL, t)

		for {
			select {
			case res := <-eventCh:
				if ignoreUnknownEvents && string(res.Event) != EventTypeCreate {
					t.Logf("Unknown event type: '%s'", res.Event)
					continue
				} else {
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

					t.Run("get event ID", func(t *testing.T) {
						defer report(t, "tdd-notification-sse", "tdd-notification-event-id")
						if string(res.ID) == "" {
							t.Fatal("missing event ID")
						}
					})

					var data mapAny
					t.Run("check event data", func(t *testing.T) {
						defer report(t, "tdd-notification-data")
						err := json.Unmarshal(res.Data, &data)
						if err != nil {
							t.Fatalf("unable to unmarshal the event data: %s", res.Data)
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

						assertEqualTitle(t, td, data)
					})
				}
				return
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
				t.Fatal("timed out waiting for subscription")
			}
		}
	})

	// t.Run("all event subscriber", func(t *testing.T) {
	// 	// subscribe to create events
	// 	eventCh := make(chan *sse.Event)
	// 	errCh := make(chan error)
	// 	client := subscribeEvent(t, serverURL+"/events", eventCh, errCh)
	// 	defer unsubscribeEvent(t, client, eventCh)

	// 	time.Sleep(waitDuration)

	// 	// add a new TD
	// 	id := "urn:uuid:" + uuid.NewV4().String()
	// 	td := mockedTD(id)
	// 	createThing(id, td, serverURL, t)

	// 	for {
	// 		select {
	// 		case res := <-eventCh:
	// 			if ignoreUnknownEvents &&
	// 				(string(res.Event) != EventTypeCreate ||
	// 					string(res.Event) != EventTypeUpdate ||
	// 					string(res.Event) != EventTypeDelete) {
	// 				t.Logf("Unknown event '%s' when subscribing to '%s'", string(res.Event), EventTypeCreate)
	// 				continue
	// 			} else {
	// 				t.Run("get event type", func(t *testing.T) {
	// 					defer report(t,
	// 						"tdd-notification-sse",
	// 						"tdd-notification-event-types",
	// 					)
	// 					if string(res.Event) != EventTypeCreate {
	// 						t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeCreate)
	// 					}
	// 				})

	// 				t.Run("get event ID", func(t *testing.T) {
	// 					defer report(t, "tdd-notification-sse", "tdd-notification-event-id")
	// 					if string(res.ID) == "" {
	// 						t.Fatal("missing event ID")
	// 					}
	// 				})

	// 				var data mapAny
	// 				t.Run("check event data", func(t *testing.T) {
	// 					defer report(t, "tdd-notification-data")
	// 					err := json.Unmarshal(res.Data, &data)
	// 					if err != nil {
	// 						t.Fatal("unable to unmarshal the event data to TDD")
	// 					}
	// 				})

	// 				t.Run("check event data td id", func(t *testing.T) {
	// 					defer report(t, "tdd-notification-data-td-id")
	// 					if id != data["id"] {
	// 						t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
	// 					}

	// 				})
	// 			}
	// 		case err := <-errCh:
	// 			t.Fatalf("unexpected error while subscribing to notification: %s", err)

	// 		case <-time.After(timeoutDuration):
	// 			t.Fatal("timed out waiting for subscription")
	// 		}
	// 	}
	// })
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
		client := subscribeEvent(t, serverURL+"/events/"+EventTypeUpdate, eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// update an attribute
		td["title"] = "updated title for update event subscriber"
		updateThing(id, td, serverURL, t)

		for {
			select {
			case res := <-eventCh:
				if ignoreUnknownEvents && string(res.Event) != EventTypeUpdate {
					t.Logf("Unknown event type: '%s'", res.Event)
					continue
				} else {
					t.Run("get event ID", func(t *testing.T) {
						defer report(t,
							"tdd-notification",
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
				}
				return
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
		}
	})

	t.Run("update event with diff subscriber", func(t *testing.T) {
		// subscribe to update events
		eventCh := make(chan *sse.Event)
		errCh := make(chan error)
		client := subscribeEvent(t, serverURL+"/events/"+EventTypeUpdate+"?diff=true", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// update an attribute
		td["title"] = "updated title for update diff event subscriber"
		updateThing(id, td, serverURL, t)

		for {
			select {
			case res := <-eventCh:
				if ignoreUnknownEvents && string(res.Event) != EventTypeUpdate {
					t.Logf("Unknown event type: '%s'", res.Event)
					continue
				} else {
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
				}
				return
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
		}
	})

	// t.Run("all event subscriber", func(t *testing.T) {
	// 	// subscribe to create events
	// 	eventCh := make(chan *sse.Event)
	// 	errCh := make(chan error)
	// 	client := subscribeEvent(t, serverURL+"/events", eventCh, errCh)
	// 	defer unsubscribeEvent(t, client, eventCh)

	// 	time.Sleep(waitDuration)

	// 	// update an attribute
	// 	td["title"] = "updated title for all event subscriber"
	// 	updateThing(id, td, serverURL, t)

	// 	select {
	// 	case res := <-eventCh:
	// 		t.Run("get event ID", func(t *testing.T) {
	// 			defer report(t, "tdd-notification-sse", "tdd-notification-event-id")
	// 			if string(res.ID) == "" {
	// 				t.Fatal("missing event ID")
	// 			}
	// 		})

	// 		t.Run("get event type", func(t *testing.T) {
	// 			defer report(t,
	// 				"tdd-notification-sse",
	// 				"tdd-notification-event-types",
	// 			)
	// 			if string(res.Event) != EventTypeUpdate {
	// 				t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeUpdate)
	// 			}
	// 		})

	// 		var data mapAny
	// 		t.Run("check event data", func(t *testing.T) {
	// 			defer report(t, "tdd-notification-data")
	// 			err := json.Unmarshal(res.Data, &data)
	// 			if err != nil {
	// 				t.Fatal("unable to unmarshal the event data to TDD")
	// 			}
	// 		})

	// 		t.Run("check event data td id", func(t *testing.T) {
	// 			defer report(t, "tdd-notification-data-td-id")
	// 			if id != data["id"] {
	// 				t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
	// 			}

	// 		})
	// 	case err := <-errCh:
	// 		t.Run("event subscription errors", func(t *testing.T) {
	// 			defer report(t, "tdd-notification-sse")
	// 			t.Fatalf("unexpected error while subscribing to notification: %s", err)
	// 		})
	// 	case <-time.After(timeoutDuration):
	// 		t.Run("event subscription timeout", func(t *testing.T) {
	// 			defer report(t, "tdd-notification-sse")
	// 			t.Fatal("timed out waiting for data")
	// 		})
	// 	}
	// })
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
		client := subscribeEvent(t, serverURL+"/events/"+EventTypeDelete, eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		deleteThing(id, serverURL, t)

		for {
			select {
			case res := <-eventCh:
				if ignoreUnknownEvents && string(res.Event) != EventTypeDelete {
					t.Logf("Unknown event type: '%s'", res.Event)
					continue
				} else {
					t.Run("get event ID", func(t *testing.T) {
						defer report(t,
							"tdd-notification",
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
				}
				return
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
		client := subscribeEvent(t, serverURL+"/events/"+EventTypeDelete+"?diff=true", eventCh, errCh)
		defer unsubscribeEvent(t, client, eventCh)

		time.Sleep(waitDuration)

		// delete the created thing
		deleteThing(id, serverURL, t)

		for {
			select {
			case res := <-eventCh:
				if ignoreUnknownEvents && string(res.Event) != EventTypeDelete {
					t.Logf("Unknown event type: '%s'", res.Event)
					continue
				} else {
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
				}
				return
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
		}
	})

	// t.Run("all event subscriber", func(t *testing.T) {
	// 	// add a new TD
	// 	id := "urn:uuid:" + uuid.NewV4().String()
	// 	td := mockedTD(id)
	// 	createThing(id, td, serverURL, t)

	// 	time.Sleep(waitDuration)

	// 	// subscribe to delete events
	// 	eventCh := make(chan *sse.Event)
	// 	errCh := make(chan error)
	// 	client := subscribeEvent(t, serverURL+"/events", eventCh, errCh)
	// 	defer unsubscribeEvent(t, client, eventCh)

	// 	time.Sleep(waitDuration)

	// 	// delete the created thing
	// 	deleteThing(id, serverURL, t)

	// 	select {
	// 	case res := <-eventCh:
	// 		t.Run("get event ID", func(t *testing.T) {
	// 			defer report(t, "tdd-notification-sse", "tdd-notification-event-id")
	// 			if string(res.ID) == "" {
	// 				t.Fatal("missing event ID")
	// 			}
	// 		})

	// 		t.Run("get event type", func(t *testing.T) {
	// 			defer report(t,
	// 				"tdd-notification-sse",
	// 				"tdd-notification-event-types",
	// 			)
	// 			if string(res.Event) != EventTypeDelete {
	// 				t.Fatalf("Unexpected event type: %s, expected: %s", string(res.Event), EventTypeDelete)
	// 			}
	// 		})

	// 		var data mapAny
	// 		t.Run("check event data", func(t *testing.T) {
	// 			defer report(t, "tdd-notification-data")
	// 			err := json.Unmarshal(res.Data, &data)
	// 			if err != nil {
	// 				t.Fatal("unable to unmarshal the event data to TDD")
	// 			}
	// 		})

	// 		t.Run("check event data td id", func(t *testing.T) {
	// 			defer report(t, "tdd-notification-data-td-id")
	// 			if id != data["id"] {
	// 				t.Fatalf("td id did not match: expected %s, got %s", id, data["id"])
	// 			}

	// 		})
	// 	case err := <-errCh:
	// 		t.Run("event subscription errors", func(t *testing.T) {
	// 			defer report(t, "tdd-notification-sse")
	// 			t.Fatalf("unexpected error while subscribing to notification: %s", err)
	// 		})
	// 	case <-time.After(timeoutDuration):
	// 		t.Run("event subscription timeout", func(t *testing.T) {
	// 			defer report(t, "tdd-notification-sse")
	// 			t.Fatal("timed out waiting for data")
	// 		})
	// 	}
	// })

}
