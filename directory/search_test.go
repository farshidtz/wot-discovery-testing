package directory

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	uuid "github.com/satori/go.uuid"
)

// RFC2119 Assertions IDs:
// tdd-search-apis-jsonPath
// tdd-search-jsonpath-method
// tdd-search-jsonpath-parameter
// tdd-search-jsonpath-response
func TestJSONPath(t *testing.T) {
	defer report(t, nil)

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
			r := &record{
				assertions: []string{"tdd-search-apis-jsonPath", "tdd-search-jsonpath-method", "tdd-search-jsonpath-parameter"},
			}
			defer report(t, r)

			// submit the request
			res, err := http.Get(serverURL + fmt.Sprintf("/search/jsonpath?query=$[?(@.tag=='%s')]", tag))
			if err != nil {
				fatal(t, r, "Error getting TDs: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			r := &record{
				assertions: []string{"tdd-search-jsonpath-response"},
			}
			defer report(t, r)

			assertStatusCode(t, r, response, http.StatusOK, body)
		})

		t.Run("payload", func(t *testing.T) {
			r := &record{
				assertions: []string{"tdd-search-jsonpath-response"},
			}
			defer report(t, r)

			var collection []mapAny
			err := json.Unmarshal(body, &collection)
			if err != nil {
				fatal(t, r, "Error decoding page: %s", err)
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
			r := &record{
				assertions: []string{"tdd-search-apis-jsonPath", "tdd-search-jsonpath-method", "tdd-search-jsonpath-parameter"},
			}
			defer report(t, r)

			res, err := http.Get(serverURL + "/search/jsonpath?query=*/id")
			if err != nil {
				fatal(t, r, "Error getting TDs: %s", err)
			}
			defer res.Body.Close()
			response = res
		})

		t.Run("status code", func(t *testing.T) {
			r := &record{
				assertions: []string{"tdd-search-jsonpath-response"},
			}
			defer report(t, r)

			assertStatusCode(t, r, response, http.StatusBadRequest, nil)
		})
	})
}

// RFC2119 Assertions IDs:
// tdd-search-apis-xPath
// tdd-search-xpath-method
// tdd-search-xpath-parameter
// tdd-search-xpath-response
func TestXPath(t *testing.T) {
	defer report(t, &record{
		comments: "TODO",
		assertions: []string{
			"tdd-search-apis-xPath",
			"tdd-search-xpath-method",
			"tdd-search-xpath-parameter",
			"tdd-search-xpath-response",
		},
	})
	t.SkipNow()
}

// RFC2119 Assertions IDs:
// tdd-search-apis-sparql
// tdd-search-semantic-protocol
// tdd-search-semantic-method-get
// tdd-search-semantic-method-post
// tdd-search-semantic-parameter
// tdd-distributed-search-semantic
// tdd-distributed-search-semantic-imp
func TestSPARQL(t *testing.T) {
	defer report(t, &record{
		comments: "TODO",
		assertions: []string{
			"tdd-search-apis-sparql",
			"tdd-search-semantic-protocol",
			"tdd-search-semantic-method-get",
			"tdd-search-semantic-method-post",
			"tdd-search-semantic-parameter",
			"tdd-distributed-search-semantic",
			"tdd-distributed-search-semantic-imp",
		},
	})
	t.SkipNow()
}
