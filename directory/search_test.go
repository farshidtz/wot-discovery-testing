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
		writeTestResult(t.Name(), "", "", t)
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
		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			t.Cleanup(func() {
				writeTestResult(t.Name(), "tdd-search-apis-jsonPath tdd-search-jsonpath-method tdd-search-jsonpath-parameter", "", t)
			})
			// submit the request
			res, err := http.Get(serverURL + fmt.Sprintf("/search/jsonpath?query=$[?(@.tag=='%s')]", tag))
			if err != nil {
				t.Fatalf("Error getting TDs: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			t.Cleanup(func() {
				writeTestResult(t.Name(), "tdd-search-jsonpath-response", "", t)
			})
			assertStatusCode(response.StatusCode, http.StatusOK, body, t)
		})

		t.Run("payload", func(t *testing.T) {
			t.Cleanup(func() {
				writeTestResult(t.Name(), "tdd-search-jsonpath-response", "", t)
			})
			var collection []mapAny
			err := json.Unmarshal(body, &collection)
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

	t.Run("reject bad query", func(t *testing.T) {
		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			t.Cleanup(func() {
				writeTestResult(t.Name(), "tdd-search-apis-jsonPath tdd-search-jsonpath-method tdd-search-jsonpath-parameter", "", t)
			})
			res, err := http.Get(serverURL + "/search/jsonpath?query=*/id")
			if err != nil {
				t.Fatalf("Error getting TDs: %s", err)
			}
			defer res.Body.Close()
			response = res
		})

		t.Run("status code", func(t *testing.T) {
			t.Cleanup(func() {
				writeTestResult(t.Name(), "tdd-search-jsonpath-response", "", t)
			})
			assertStatusCode(response.StatusCode, http.StatusBadRequest, nil, t)
		})
	})

}

func TestXPath(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult(t.Name(), "", "TODO", t)
	})
	t.SkipNow()
}
