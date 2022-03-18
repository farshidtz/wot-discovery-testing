package directory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestJSONPath(t *testing.T) {

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
			defer report(t,
				"tdd-search-jsonpath",
				"tdd-search-jsonpath-method",
				"tdd-search-jsonpath-parameter",
			)

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
			defer report(t, "tdd-search-jsonpath-response")

			assertStatusCode(t, response, http.StatusOK, body)
		})

		t.Run("content type", func(t *testing.T) {
			defer report(t, "tdd-search-jsonpath-response")

			assertContentMediaType(t, response, MediaTypeJSON)
		})

		t.Run("payload", func(t *testing.T) {
			defer report(t, "tdd-search-jsonpath-response")

			var filterredTDs []mapAny
			err := json.Unmarshal(body, &filterredTDs)
			if err != nil {
				t.Fatalf("Error decoding page: %s", err)
			}

			if len(createdTD) != len(filterredTDs) {
				t.Fatalf("Filtering returned %d TDs, expected %d", len(filterredTDs), len(createdTD))
			}

			createdTDsMap := make(map[string]mapAny)
			for _, createdTD := range createdTD {
				id := getID(t, createdTD)
				createdTDsMap[id] = createdTD
			}

			// compare created and filterred
			for _, filterredTD := range filterredTDs {
				id := getID(t, filterredTD)
				if _, found := createdTDsMap[id]; !found {
					t.Fatalf("Result does not include the TD with id: %s", id)
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
		defer report(t, "tdd-reg-anonymous-td-identifier")

		// add an anonymous TD
		createdTD := mockedTD("") // no id
		// tag the TDs to find later
		tag := uuid.NewV4().String()
		createdTD["tag"] = tag
		createThing("", createdTD, serverURL, t)

		// submit the request
		response, err := http.Get(serverURL + fmt.Sprintf("/search/jsonpath?query=$[?(@.tag=='%s')]", tag))
		if err != nil {
			t.Fatalf("Error getting TDs: %s", err)
		}
		defer response.Body.Close()

		body := httpReadBody(response, t)

		var filterredTDs []mapAny
		err = json.Unmarshal(body, &filterredTDs)
		if err != nil {
			t.Fatalf("Error decoding page: %s", err)
		}

		if len(filterredTDs) != 1 {
			t.Fatalf("Filtering returned %d TDs, expected 1", len(filterredTDs))
		}

		// try to get the ID. This should pass
		getID(t, filterredTDs[0])
	})

	t.Run("reject bad query", func(t *testing.T) {
		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			defer report(t,
				"tdd-search-jsonpath",
				"tdd-search-jsonpath-method",
				"tdd-search-jsonpath-parameter",
			)

			res, err := http.Get(serverURL + "/search/jsonpath?query=*/id")
			if err != nil {
				t.Fatalf("Error getting TDs: %s", err)
			}
			defer res.Body.Close()
			response = res
		})

		t.Run("status code", func(t *testing.T) {
			defer report(t, "tdd-search-jsonpath-response")

			assertStatusCode(t, response, http.StatusBadRequest, nil)
		})
	})
}

func TestXPath(t *testing.T) {
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
			defer report(t,
				"tdd-search-xpath",
				"tdd-search-xpath-method",
				"tdd-search-xpath-parameter",
			)

			// submit the request
			res, err := http.Get(serverURL + fmt.Sprintf("/search/xpath?query=*[tag='%s']", tag))
			if err != nil {
				t.Fatalf("Error getting TDs: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			defer report(t, "tdd-search-xpath-response")

			assertStatusCode(t, response, http.StatusOK, body)
		})

		t.Run("content type", func(t *testing.T) {
			defer report(t, "tdd-search-xpath-response")

			assertContentMediaType(t, response, MediaTypeJSON)
		})

		t.Run("payload", func(t *testing.T) {
			defer report(t, "tdd-search-xpath-response")

			var filterredTDs []mapAny
			err := json.Unmarshal(body, &filterredTDs)
			if err != nil {
				t.Fatalf("Error decoding page: %s", err)
			}

			if len(createdTD) != len(filterredTDs) {
				t.Fatalf("Filtering returned %d TDs, expected %d", len(filterredTDs), len(createdTD))
			}

			createdTDsMap := make(map[string]mapAny)
			for _, createdTD := range createdTD {
				id := getID(t, createdTD)
				createdTDsMap[id] = createdTD
			}

			// compare created and filterred
			for _, filterredTD := range filterredTDs {
				id := getID(t, filterredTD)
				if _, found := createdTDsMap[id]; !found {
					t.Fatalf("Result does not include the TD with id: %s", id)
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
		defer report(t, "tdd-reg-anonymous-td-identifier")

		// add an anonymous TD
		createdTD := mockedTD("") // no id
		// tag the TDs to find later
		tag := uuid.NewV4().String()
		createdTD["tag"] = tag
		createThing("", createdTD, serverURL, t)

		// submit the request
		response, err := http.Get(serverURL + fmt.Sprintf("/search/xpath?query=*[tag='%s']", tag))
		if err != nil {
			t.Fatalf("Error getting TDs: %s", err)
		}
		defer response.Body.Close()

		body := httpReadBody(response, t)

		var filterredTDs []mapAny
		err = json.Unmarshal(body, &filterredTDs)
		if err != nil {
			t.Fatalf("Error decoding page: %s", err)
		}

		if len(filterredTDs) != 1 {
			t.Fatalf("Filtering returned %d TDs, expected 1", len(filterredTDs))
		}

		// try to get the ID. This should pass
		getID(t, filterredTDs[0])
	})

	t.Run("reject bad query", func(t *testing.T) {
		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			defer report(t,
				"tdd-search-xpath",
				"tdd-search-xpath-method",
				"tdd-search-xpath-parameter",
			)

			res, err := http.Get(serverURL + "/search/xpath?query=$[:].id")
			if err != nil {
				t.Fatalf("Error getting TDs: %s", err)
			}
			defer res.Body.Close()
			response = res
		})

		t.Run("status code", func(t *testing.T) {
			defer report(t, "tdd-search-xpath-response")

			assertStatusCode(t, response, http.StatusBadRequest, nil)
		})
	})
}

func TestSPARQL(t *testing.T) {

	const query = `select * { ?s ?p ?o }limit 5`
	const federatedQuery = `select * {
		service <https://dbpedia.org/sparql>{
			 ?s ?p ?o
		}
	}limit 5`

	var expectedResult = sparqlResultsSample()

	t.Run("search using GET", func(t *testing.T) {
		defer report(t,
			"tdd-search-sparql",
			"tdd-search-sparql-method-get",
			"tdd-search-sparql-resp",
		)

		// submit GET request
		res, err := http.Get(serverURL + "/search/sparql?query=" + url.QueryEscape(query))
		if err != nil {
			t.Fatalf("Error solving query SPARQL: %s", err)
		}
		body := httpReadBody(res, t)

		var responseMap mapAny
		err = json.Unmarshal(body, &responseMap)
		if err != nil {
			t.Fatalf("Error decoding response: %s", err)
		}

		t.Log(responseMap)
		delete(responseMap, "results")

		if !serializedEqual(responseMap, expectedResult) {
			t.Fatalf("Expected:\n%v\nGot:\n%v\n",
				marshalPrettyJSON(expectedResult), marshalPrettyJSON(responseMap))
		}
	})

	t.Run("search using POST", func(t *testing.T) {
		defer report(t,
			"tdd-search-sparql",
			"tdd-search-sparql-method-post",
			"tdd-search-sparql-resp",
		)

		// submit POST request
		res, err := http.Post(serverURL+"/search/sparql",
			"application/sparql-query",
			bytes.NewReader([]byte(query)))
		if err != nil {
			t.Fatalf("Error solving query SPARQL: %s", err)
		}
		body := httpReadBody(res, t)

		var responseMap mapAny
		err = json.Unmarshal(body, &responseMap)
		if err != nil {
			t.Fatalf("Error decoding response: %s", err)
		}

		t.Log(responseMap)
		delete(responseMap, "results")

		if !serializedEqual(responseMap, expectedResult) {
			t.Fatalf("Expected:\n%v\nGot:\n%v\n",
				marshalPrettyJSON(expectedResult), marshalPrettyJSON(responseMap))
		}
	})

	t.Run("federated search using GET", func(t *testing.T) {
		defer report(t,
			"tdd-search-sparql",
			"tdd-search-sparql-method-get",
			"tdd-search-sparql-resp",
			"tdd-search-sparql-federation",
		)

		// submit GET request
		res, err := http.Get(serverURL + "/search/sparql?query=" + url.QueryEscape(federatedQuery))
		if err != nil {
			t.Fatalf("Error solving query SPARQL: %s", err)
		}
		body := httpReadBody(res, t)

		var responseMap mapAny
		err = json.Unmarshal(body, &responseMap)
		if err != nil {
			t.Fatalf("Error decoding response: %s", err)
		}

		t.Log(responseMap)
		delete(responseMap, "results")

		if !serializedEqual(responseMap, expectedResult) {
			t.Fatalf("Expected:\n%v\nGot:\n%v\n",
				marshalPrettyJSON(expectedResult), marshalPrettyJSON(responseMap))
		}
	})
}

func sparqlResultsSample() mapAny {
	var qr = mapAny{
		"head": mapAny{
			"vars": []string{
				"s",
				"p",
				"o",
			},
		},
	}
	return qr
}
