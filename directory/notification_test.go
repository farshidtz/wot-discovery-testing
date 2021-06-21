package directory

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/r3labs/sse/v2"
	uuid "github.com/satori/go.uuid"
	net "github.com/subchord/go-sse"
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

// RFC2119 Assertions IDs:
// tdd-notification-event-types
// tdd-notification-filter-type
// tdd-notification-data
// tdd-notification-data-tdid
// tdd-notification-data-create-full
// tdd-notification-data-update-diff
// tdd-notification-data-update-id
// tdd-notification-data-delete-diff
// tdd-notification-data-diff-unsupported

type Event struct {
	ID   string `json:"id"`
	Type string `json:"event"`
	Data mapAny `json:"data"`
}

func TestCreateEvent3(t *testing.T) {
	defer report(t, &record{comments: "TODO"})

	// subscribe to create events
	client := sse.NewClient(serverURL + "/events/blah")
	client.OnDisconnect(func(c *sse.Client) {
		t.Fatal("disconnected")
	})

	c := make(chan *sse.Event)
	go func() {}()
	client.on
	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	createThing(id, td, serverURL, t)

	select {
	case res := <-c:
		fmt.Println(res)
	case <-time.After(3 * time.Second):
		t.Fatal("timedout")
	}
	client.Unsubscribe(c)
	//t.SkipNow()
}

func TestCreateEvent2(t *testing.T) {
	defer report(t, &record{comments: "TODO"})
	feed, err := net.ConnectWithSSEFeed(serverURL+"/events/create", nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	sub, err := feed.Subscribe("create")
	if err != nil {
		return
	}

	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	createThing(id, td, serverURL, t)

	select {
	case evt := <-sub.Feed():
		log.Print(evt)
	case err := <-sub.ErrFeed():
		log.Fatal(err)
		return
	case <-time.After(3 * time.Second):
		t.Fatal("timedout")
	}

	sub.Close()
	feed.Close()
	//t.SkipNow()
}
