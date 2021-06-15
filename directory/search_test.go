package directory

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestJSONPath(t *testing.T) {
	defer report(t, nil)

	t.Run("filter", func(t *testing.T) {
		tag := uuid.NewV4().String()
		var createdTD []mapAny
		for i := 0; i < 3; i++ {
			id := "urn:uuid:" + uuid.NewV4().String()
			td := mockedTD(id)
			// tag the TDs to find later
			td["tag"] = tag
			createdTD = append(createdTD, td)
			createThing(id, td, serverURL, t)
		}

		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			r := &record{
				assertions: []string{
					"tdd-search-jsonpath",
					"tdd-search-jsonpath-method",
					"tdd-search-jsonpath-parameter",
				},
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

		t.Run("content type", func(t *testing.T) {
			r := &record{
				assertions: []string{"tdd-search-jsonpath-response"},
			}
			defer report(t, r)

			assertContentMediaType(t, r, response, MediaTypeJSON)
		})

		t.Run("payload", func(t *testing.T) {
			r := &record{
				assertions: []string{"tdd-search-jsonpath-response"},
			}
			defer report(t, r)

			var filterredTDs []mapAny
			err := json.Unmarshal(body, &filterredTDs)
			if err != nil {
				fatal(t, r, "Error decoding page: %s", err)
			}

			if len(createdTD) != len(filterredTDs) {
				fatal(t, r, "Filtering returned %d TDs, expected %d", len(filterredTDs), len(createdTD))
			}

			createdTDsMap := make(map[string]mapAny)
			for _, createdTD := range createdTD {
				id := getID(t, r, createdTD)
				createdTDsMap[id] = createdTD
			}

			// compare created and filterred
			for _, filterredTD := range filterredTDs {
				id := getID(t, r, filterredTD)
				if _, found := createdTDsMap[id]; !found {
					fatal(t, r, "Result does not include the TD with id: %s", id)
				}

				// remove system-generated attributes
				delete(filterredTD, "registration")

				if !serializedEqual(createdTDsMap[id], filterredTD) {
					t.Fatalf("Expected:\n%v\nGot:\n%v\n",
						marshalPrettyJSON(createdTDsMap[id]), marshalPrettyJSON(filterredTD))
				}
			}
		})
	})

	t.Run("filter anonymous", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-anonymous-td-identifier"},
		}
		defer report(t, r)

		// add an anonymous TD
		createdTD := mockedTD("") // no id
		// tag the TDs to find later
		tag := uuid.NewV4().String()
		createdTD["tag"] = tag
		createThing("", createdTD, serverURL, t)

		// submit the request
		response, err := http.Get(serverURL + fmt.Sprintf("/search/jsonpath?query=$[?(@.tag=='%s')]", tag))
		if err != nil {
			fatal(t, r, "Error getting TDs: %s", err)
		}
		defer response.Body.Close()

		body := httpReadBody(response, t)

		var filterredTDs []mapAny
		err = json.Unmarshal(body, &filterredTDs)
		if err != nil {
			fatal(t, r, "Error decoding page: %s", err)
		}

		if len(filterredTDs) != 1 {
			fatal(t, r, "Filtering returned %d TDs, expected 1", len(filterredTDs))
		}

		// try to get the ID. This should pass
		getID(t, r, filterredTDs[0])
	})

	t.Run("reject bad query", func(t *testing.T) {
		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			r := &record{
				assertions: []string{
					"tdd-search-jsonpath",
					"tdd-search-jsonpath-method",
					"tdd-search-jsonpath-parameter",
				},
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

func TestXPath(t *testing.T) {
	defer report(t, nil)

	t.Run("filter", func(t *testing.T) {
		tag := uuid.NewV4().String()
		var createdTD []mapAny
		for i := 0; i < 3; i++ {
			id := "urn:uuid:" + uuid.NewV4().String()
			td := mockedTD(id)
			// tag the TDs to find later
			td["tag"] = tag
			createdTD = append(createdTD, td)
			createThing(id, td, serverURL, t)
		}

		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			r := &record{
				assertions: []string{
					"tdd-search-xpath",
					"tdd-search-xpath-method",
					"tdd-search-xpath-parameter",
				},
			}
			defer report(t, r)

			// submit the request
			res, err := http.Get(serverURL + fmt.Sprintf("/search/xpath?query=*[tag='%s']", tag))
			if err != nil {
				fatal(t, r, "Error getting TDs: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			r := &record{
				assertions: []string{"tdd-search-xpath-response"},
			}
			defer report(t, r)

			assertStatusCode(t, r, response, http.StatusOK, body)
		})

		t.Run("content type", func(t *testing.T) {
			r := &record{
				assertions: []string{"tdd-search-xpath-response"},
			}
			defer report(t, r)

			assertContentMediaType(t, r, response, MediaTypeJSON)
		})

		t.Run("payload", func(t *testing.T) {
			r := &record{
				assertions: []string{"tdd-search-xpath-response"},
			}
			defer report(t, r)

			var filterredTDs []mapAny
			err := json.Unmarshal(body, &filterredTDs)
			if err != nil {
				fatal(t, r, "Error decoding page: %s", err)
			}

			if len(createdTD) != len(filterredTDs) {
				fatal(t, r, "Filtering returned %d TDs, expected %d", len(filterredTDs), len(createdTD))
			}

			createdTDsMap := make(map[string]mapAny)
			for _, createdTD := range createdTD {
				id := getID(t, r, createdTD)
				createdTDsMap[id] = createdTD
			}

			// compare created and filterred
			for _, filterredTD := range filterredTDs {
				id := getID(t, r, filterredTD)
				if _, found := createdTDsMap[id]; !found {
					fatal(t, r, "Result does not include the TD with id: %s", id)
				}

				// remove system-generated attributes
				delete(filterredTD, "registration")

				if !serializedEqual(createdTDsMap[id], filterredTD) {
					t.Fatalf("Expected:\n%v\nGot:\n%v\n",
						marshalPrettyJSON(createdTDsMap[id]), marshalPrettyJSON(filterredTD))
				}
			}
		})
	})

	t.Run("filter anonymous", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-anonymous-td-identifier"},
		}
		defer report(t, r)

		// add an anonymous TD
		createdTD := mockedTD("") // no id
		// tag the TDs to find later
		tag := uuid.NewV4().String()
		createdTD["tag"] = tag
		createThing("", createdTD, serverURL, t)

		// submit the request
		response, err := http.Get(serverURL + fmt.Sprintf("/search/xpath?query=*[tag='%s']", tag))
		if err != nil {
			fatal(t, r, "Error getting TDs: %s", err)
		}
		defer response.Body.Close()

		body := httpReadBody(response, t)

		var filterredTDs []mapAny
		err = json.Unmarshal(body, &filterredTDs)
		if err != nil {
			fatal(t, r, "Error decoding page: %s", err)
		}

		if len(filterredTDs) != 1 {
			fatal(t, r, "Filtering returned %d TDs, expected 1", len(filterredTDs))
		}

		// try to get the ID. This should pass
		getID(t, r, filterredTDs[0])
	})

	t.Run("reject bad query", func(t *testing.T) {
		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			r := &record{
				assertions: []string{
					"tdd-search-xpath",
					"tdd-search-xpath-method",
					"tdd-search-xpath-parameter",
				},
			}
			defer report(t, r)

			res, err := http.Get(serverURL + "/search/xpath?query=$[:].id")
			if err != nil {
				fatal(t, r, "Error getting TDs: %s", err)
			}
			defer res.Body.Close()
			response = res
		})

		t.Run("status code", func(t *testing.T) {
			r := &record{
				assertions: []string{"tdd-search-xpath-response"},
			}
			defer report(t, r)

			assertStatusCode(t, r, response, http.StatusBadRequest, nil)
		})
	})
}

func TestSPARQL(t *testing.T) {
	r := &record{
		assertions: []string{
			"tdd-search-sparql",
			"tdd-search-sparql-version",
			"tdd-search-sparql-method-get",
			"tdd-search-sparql-method-post",
			"tdd-search-sparql-resp",
			"tdd-search-sparql-federation",
			"tdd-search-sparql-federation-imp",
		},
	}
	defer report(t, r)

	fatal(t, r, "TODO")
}
