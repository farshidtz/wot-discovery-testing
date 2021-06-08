package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		createThing(id, td, t)
	}

	t.Run("Filter", func(t *testing.T) {
		res, err := http.Get(serverURL + fmt.Sprintf("/search/jsonpath?query=$[?(@.tag=='%s')]", tag))
		if err != nil {
			t.Fatalf("Error getting TDs: %s", err)
		}
		defer res.Body.Close()

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Error reading response body: %s", err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected status %v, got: %d. Response body:\n%s", http.StatusOK, res.StatusCode, b)
		}

		var collectionPage []mapAny
		err = json.Unmarshal(b, &collectionPage)
		if err != nil {
			t.Fatalf("Error decoding page: %s", err)
		}

		storedTDs := retrieveAllThings(t)

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

	t.Run("Filter fail", func(t *testing.T) {
		res, err := http.Get(serverURL + "/search/jsonpath?query=*/id")
		if err != nil {
			t.Fatalf("Error getting TDs: %s", err)
		}
		defer res.Body.Close()

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Error reading response body: %s", err)
		}
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected status %v, got: %d. Response body:\n%s", http.StatusBadRequest, res.StatusCode, b)
		}
	})

}
