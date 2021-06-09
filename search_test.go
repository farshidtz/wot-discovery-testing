package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestJSONPath(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("search-jsonpath", "", t)
	})

	tag := uuid.NewV4().String()
	for i := 0; i < 3; i++ {
		// add through controller
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		// tag the TDs to find later
		td["tag"] = tag
		createThing(id, td, serverURL, t)
	}

	t.Run("filter", func(t *testing.T) {
		res, err := http.Get(serverURL + fmt.Sprintf("/search/jsonpath?query=$[?(@.tag=='%s')]", tag))
		if err != nil {
			t.Fatalf("Error getting TDs: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(res.StatusCode, http.StatusOK, body, t)
		})

		t.Run("payload", func(t *testing.T) {
			var collection []mapAny
			err = json.Unmarshal(body, &collection)
			if err != nil {
				t.Fatalf("Error decoding page: %s", err)
			}

			storedTDs := retrieveAllThings(serverURL, t)

			// compare response and stored
			for i, sd := range storedTDs {
				if sd["tag"] == tag {
					if !reflect.DeepEqual(storedTDs[i], sd) {
						t.Fatalf("TD responded over HTTP is different with the one stored:\n Stored:\n%v\n Listed\n%v\n",
							storedTDs[i], sd)
					}
				}
			}
		})
	})

	t.Run("fail bad query", func(t *testing.T) {
		res, err := http.Get(serverURL + "/search/jsonpath?query=*/id")
		if err != nil {
			t.Fatalf("Error getting TDs: %s", err)
		}
		defer res.Body.Close()

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(res.StatusCode, http.StatusBadRequest, nil, t)
		})
	})

}

func TestXPath(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("search-xpath", "TODO", t)
	})
	t.SkipNow()
}
