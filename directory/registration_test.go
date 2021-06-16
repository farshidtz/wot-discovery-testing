package directory

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
)

func TestCreateAnonymousThing(t *testing.T) {
	defer report(t, nil)

	td := mockedTD("") // without ID
	b, _ := json.Marshal(td)

	var response *http.Response

	t.Run("submit request", func(t *testing.T) {
		r := &record{
			assertions: []string{
				"tdd-reg-crudl",
				"tdd-reg-create-anonymous-td",
				"tdd-reg-create-body",
				"tdd-reg-create-contenttype",
			},
		}
		defer report(t, r)

		// submit POST request
		res, err := http.Post(serverURL+"/things/", MediaTypeThingDescription, bytes.NewReader(b))
		if err != nil {
			fatal(t, r, "Error posting: %s", err)
		}
		response = res
		// defer res.Body.Close()
	})

	body := httpReadBody(response, t)

	t.Run("status code", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-create-anonymous-td-resp"},
		}
		defer report(t, r)
		assertStatusCode(t, r, response, http.StatusCreated, body)
	})

	var systemGeneratedID string
	t.Run("location header", func(t *testing.T) {
		r := &record{
			assertions: []string{
				"tdd-reg-create-anonymous-td-resp",
				"tdd-reg-anonymous-td-local-id",
			},
		}
		defer report(t, r)

		// Check if system-generated id is in response
		location, err := response.Location()
		if err != nil {
			fatal(t, r, err.Error())
		}
		systemGeneratedID = location.String()
		if systemGeneratedID == "" {
			fatal(t, r, "System-generated ID not in response. Got location header: %s", location)
		}
		_, err = url.ParseRequestURI(systemGeneratedID)
		if err != nil {
			fatal(t, r, "System-generated ID not in a valid URI. Got: %s", location)
		}
		if !strings.Contains(systemGeneratedID, "urn:uuid:") {
			fatal(t, r, "System-generated ID doesn't have URN UUID scheme. Got: %s", location)
		}
	})

	t.Run("result", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-create-anonymous-td"},
		}
		defer report(t, r)

		if systemGeneratedID == "" {
			fatal(t, r, "previous errors")
		}

		// retrieve the stored TD
		storedTD := retrieveThing(systemGeneratedID, serverURL, t)

		// remove system-generated attributes
		delete(td, "registration")
		delete(storedTD, "registration")
		delete(storedTD, "id")

		if !serializedEqual(td, storedTD) {
			t.Logf("Expected:\n%v\nRetrieved:\n%v\n", marshalPrettyJSON(td), marshalPrettyJSON(storedTD))
			fatal(t, r, "Stored TD was does not match the expectations; see logs.")
		}
	})

	t.Run("registration info", func(t *testing.T) {
		// retrieve the stored TD
		storedTD := retrieveThing(systemGeneratedID, serverURL, t)

		testRegistrionInfo(t, storedTD)
	})

	// reject PUT of anonymous TD
	t.Run("reject PUT", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-create-known-vs-anonymous"},
		}
		defer report(t, r)

		td := mockedTD("") // no id
		b, _ := json.Marshal(td)

		// submit PUT request
		res, err := httpPut(serverURL+"/things/", MediaTypeThingDescription, b)
		if err != nil {
			fatal(t, r, "Error putting: %s", err)
		}
		defer res.Body.Close()

		if res.StatusCode < 400 && res.StatusCode >= 500 {
			fatal(t, r, "Anonymous TD submission with PUT not rejected. Got status: %d", res.StatusCode)
		}
	})

	t.Run("reject invalid", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-http-error-response"},
		}
		defer report(t, r)

		td := mockedTD("")  // no id
		delete(td, "title") // remove the mandatory field

		b, _ := json.Marshal(td)

		// submit POST request
		res, err := http.Post(serverURL+"/things/", MediaTypeThingDescription, bytes.NewReader(b))
		if err != nil {
			fatal(t, r, "Error posting: %s", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			fatal(t, r, "Expected status %d, got : %d", http.StatusBadRequest, res.StatusCode)
		}

		assertErrorResponse(t, r, res, nil, true)
	})
}

func TestCreateThing(t *testing.T) {
	defer report(t, nil)

	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	b, _ := json.Marshal(td)

	var response *http.Response

	t.Run("submit request", func(t *testing.T) {
		r := &record{
			assertions: []string{
				"tdd-reg-crudl",
				"tdd-reg-create-known-td",
				"tdd-reg-create-body",
				"tdd-reg-create-contenttype",
			},
		}
		defer report(t, r)

		// submit PUT request
		res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
		if err != nil {
			fatal(t, r, "Error posting: %s", err)
		}
		response = res
		// defer res.Body.Close()
	})

	body := httpReadBody(response, t)

	t.Run("status code", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-create-known-td-resp"},
		}
		defer report(t, r)

		assertStatusCode(t, r, response, http.StatusCreated, body)
	})

	t.Run("result", func(t *testing.T) {
		r := &record{
			assertions: []string{
				"tdd-reg-create-known-td",
				"tdd-reg-create-body",
			},
		}
		defer report(t, r)

		// retrieve the stored TD
		storedTD := retrieveThing(id, serverURL, t)

		// remove system-generated attributes
		delete(td, "registration")
		delete(storedTD, "registration")

		if !serializedEqual(td, storedTD) {
			t.Logf("Expected:\n%v\nRetrieved:\n%v\n", marshalPrettyJSON(td), marshalPrettyJSON(storedTD))
			fatal(t, r, "Unexpected result body; see logs.")
		}
	})

	t.Run("reject id mismatch", func(t *testing.T) {
		r := &record{
			assertions: []string{},
		}
		defer report(t, r)
		skip(t, r, "no relevant assertions")

		id := "urn:uuid:" + uuid.NewV4().String()
		anotherID := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(anotherID)
		b, _ := json.Marshal(td)

		var response *http.Response

		t.Run("status code", func(t *testing.T) {
			r := &record{
				assertions: []string{},
			}
			defer report(t, r)

			// submit PUT request
			res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
			if err != nil {
				fatal(t, r, "Error posting: %s", err)
			}
			response = res
			// defer res.Body.Close()
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			r := &record{
				assertions: []string{},
			}
			defer report(t, r)
			assertStatusCode(t, r, response, http.StatusConflict, body)
		})
	})

	t.Run("reject invalid", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-http-error-response"},
		}
		defer report(t, r)

		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		delete(td, "title") // remove the mandatory field

		b, _ := json.Marshal(td)

		// submit PUT request
		res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
		if err != nil {
			fatal(t, r, "Error putting: %s", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			fatal(t, r, "Expected status %d, got : %d", http.StatusBadRequest, res.StatusCode)
		}

		assertErrorResponse(t, r, res, nil, true)
	})

	// reject POST with id
	t.Run("reject POST", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-create-known-vs-anonymous"},
		}
		defer report(t, r)
		skip(t, r, "not disallowed explicitly")

		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		b, _ := json.Marshal(td)

		// submit POST request
		res, err := http.Post(serverURL+"/things/", MediaTypeThingDescription, bytes.NewReader(b))
		if err != nil {
			fatal(t, r, "Error posting: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(t, r, res, http.StatusBadRequest, body)
		})
	})

}

func TestRetrieveThing(t *testing.T) {
	defer report(t, nil)

	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	createThing(id, td, serverURL, t)

	var response *http.Response

	t.Run("submit request", func(t *testing.T) {
		r := &record{
			assertions: []string{
				"tdd-reg-crudl",
				"tdd-reg-retrieve",
			},
		}
		defer report(t, r)

		// submit GET request
		res, err := http.Get(serverURL + "/td/" + id)
		if err != nil {
			fatal(t, r, "Error getting TD: %s", err)
		}
		response = res
		// defer res.Body.Close()
	})

	body := httpReadBody(response, t)

	t.Run("status code", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-retrieve-resp"},
		}
		defer report(t, r)

		assertStatusCode(t, r, response, http.StatusOK, body)
	})

	t.Run("content type", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-retrieve-resp"},
		}
		defer report(t, r)

		assertContentMediaType(t, r, response, MediaTypeThingDescription)
	})

	t.Run("result", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-retrieve"},
		}
		defer report(t, r)

		var retrievedTD mapAny
		err := json.Unmarshal(body, &retrievedTD)
		if err != nil {
			fatal(t, r, "Error decoding body: %s", err)
		}

		// remove system-generated attributes
		delete(retrievedTD, "registration")

		if !serializedEqual(td, retrievedTD) {
			t.Logf("Expected:\n%v\nRetrieved:\n%v", marshalPrettyJSON(td), marshalPrettyJSON(retrievedTD))
			fatal(t, r, "The retrieved TD is not the same as the added one; see logs.")
		}
	})

	t.Run("enriched result", func(t *testing.T) {
		r := &record{
			assertions: []string{},
		}
		defer report(t, r)

		fatal(t, r, "TODO")
	})

	t.Run("anonymous td id", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-anonymous-td-identifier"},
		}
		defer report(t, r)

		fatal(t, r, "TODO")
	})
}

func TestUpdateThing(t *testing.T) {
	defer report(t, nil)

	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	createThing(id, td, serverURL, t)

	// update an attribute
	td["title"] = "updated title"
	b, _ := json.Marshal(td)

	var response *http.Response

	t.Run("submit request", func(t *testing.T) {
		r := &record{
			assertions: []string{
				"tdd-reg-crudl",
				"tdd-reg-update-types",
				"tdd-reg-update",
				"tdd-reg-update-contenttype"},
		}
		defer report(t, r)

		// submit PUT request
		res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
		if err != nil {
			fatal(t, r, "Error putting TD: %s", err)
		}
		response = res
		// defer res.Body.Close()
	})

	body := httpReadBody(response, t)

	t.Run("status code", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-update-resp"},
		}
		defer report(t, r)

		assertStatusCode(t, r, response, http.StatusNoContent, body)
	})

	t.Run("result", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-update-types", "tdd-reg-update", "tdd-reg-update-contenttype"},
		}
		defer report(t, r)

		// retrieve the stored TD
		storedTD := retrieveThing(id, serverURL, t)

		// remove system-generated attributes
		delete(td, "registration")
		delete(storedTD, "registration")

		if !serializedEqual(td, storedTD) {
			t.Logf("Expected:\n%v\n Retrieved:\n%v\n", marshalPrettyJSON(td), marshalPrettyJSON(storedTD))
			fatal(t, r, "Unexpected result body; see logs.")
		}
	})

	t.Run("reject invalid", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-http-error-response"},
		}
		defer report(t, r)

		delete(td, "title") // remove the mandatory field

		b, _ := json.Marshal(td)

		// submit PUT request
		res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
		if err != nil {
			fatal(t, r, "Error putting: %s", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			fatal(t, r, "Expected status %d, got : %d", http.StatusBadRequest, res.StatusCode)
		}

		assertErrorResponse(t, r, res, nil, true)
	})
}

func TestPatch(t *testing.T) {
	defer report(t, nil)

	var (
		requestAssertions = []string{
			"tdd-reg-update-partial",
			"tdd-reg-update-partial-partialtd",
			"tdd-reg-update-partial-contenttype",
		}
		statusAssertions = []string{"tdd-reg-update-partial-resp"}
		resultAssertions = []string{
			"tdd-reg-update-partial",
			"tdd-reg-update-partial-mergepatch",
		}
	)

	t.Run("replace title", func(t *testing.T) {
		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, serverURL, t)

		// update the title
		jsonTD := `{"title": "new title"}`

		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			r := &record{
				assertions: requestAssertions,
			}
			defer report(t, r)

			// submit PATCH request
			res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
			if err != nil {
				fatal(t, r, "Error patching TD: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			r := &record{
				assertions: statusAssertions,
			}
			defer report(t, r)

			assertStatusCode(t, r, response, http.StatusNoContent, body)
		})

		t.Run("result", func(t *testing.T) {
			r := &record{
				assertions: resultAssertions,
			}
			defer report(t, r)

			// retrieve the changed TD
			storedTD := retrieveThing(id, serverURL, t)

			// manually change attributes of the reference TD
			td["title"] = "new title"
			// remove system-generated attributes
			delete(td, "registration")
			delete(storedTD, "registration")

			if !serializedEqual(td, storedTD) {
				t.Logf("Expected:\n%s\n Retrieved:\n%s\n", marshalPrettyJSON(td), marshalPrettyJSON(storedTD))
				fatal(t, r, "Unexpected result body; see logs.")
			}
		})
	})

	t.Run("remove description", func(t *testing.T) {
		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		td["description"] = "this is a test descr"
		createThing(id, td, serverURL, t)

		// set description to null to remove it
		jsonTD := `{"description": null}`

		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			r := &record{
				assertions: requestAssertions,
			}
			defer report(t, r)

			// submit PATCH request
			res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
			if err != nil {
				fatal(t, r, "Error patching TD: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			r := &record{
				assertions: statusAssertions,
			}
			defer report(t, r)

			assertStatusCode(t, r, response, http.StatusNoContent, body)
		})

		t.Run("result", func(t *testing.T) {
			r := &record{
				assertions: resultAssertions,
			}
			defer report(t, r)

			// retrieve the changed TD
			storedTD := retrieveThing(id, serverURL, t)

			// manually change attributes of the reference TD
			delete(td, "description")
			// remove system-generated attributes
			delete(td, "registration")
			delete(storedTD, "registration")

			if !serializedEqual(td, storedTD) {
				t.Logf("Expected:\n%s\n Retrieved:\n%s\n", marshalPrettyJSON(td), marshalPrettyJSON(storedTD))
				fatal(t, r, "Unexpected result body; see logs.")
			}
		})
	})

	t.Run("update properties", func(t *testing.T) {
		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		td["properties"] = map[string]interface{}{
			"status": map[string]interface{}{
				"forms": []map[string]interface{}{
					{"href": "https://mylamp.example.com/status"},
				},
			},
		}
		createThing(id, td, serverURL, t)

		// patch with new property
		jsonTD := `{"properties": {"new_property": {"forms": [{"href": "https://mylamp.example.com/new_property"}]}}}`

		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			r := &record{
				assertions: requestAssertions,
			}
			defer report(t, r)

			// submit PATCH request
			res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
			if err != nil {
				fatal(t, r, "Error patching TD: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			r := &record{
				assertions: statusAssertions,
			}
			defer report(t, r)

			assertStatusCode(t, r, response, http.StatusNoContent, body)
		})

		t.Run("result", func(t *testing.T) {
			r := &record{
				assertions: resultAssertions,
			}
			defer report(t, r)

			// retrieve the changed TD
			storedTD := retrieveThing(id, serverURL, t)

			// manually change attributes of the reference TD
			td["properties"] = map[string]interface{}{
				"status": map[string]interface{}{
					"forms": []map[string]interface{}{
						{"href": "https://mylamp.example.com/status"},
					},
				},
				"new_property": map[string]interface{}{
					"forms": []map[string]interface{}{
						{"href": "https://mylamp.example.com/new_property"},
					},
				},
			}
			// remove system-generated attributes
			delete(td, "registration")
			delete(storedTD, "registration")

			if !serializedEqual(td, storedTD) {
				t.Logf("Expected:\n%s\n Retrieved:\n%s\n", marshalPrettyJSON(td), marshalPrettyJSON(storedTD))
				fatal(t, r, "Unexpected result body; see logs.")
			}
		})
	})

	t.Run("replace array", func(t *testing.T) {
		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		td["properties"] = map[string]interface{}{
			"status": map[string]interface{}{
				"forms": []map[string]interface{}{
					{"href": "https://mylamp.example.com/status"},
				},
			},
		}
		createThing(id, td, serverURL, t)

		// patch with different array
		jsonTD := `{"properties": {"status": {"forms": [
					{"href": "https://mylamp.example.com/status"},
					{"href": "coaps://mylamp.example.com/status"}
				]}}}`

		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			r := &record{
				assertions: requestAssertions,
			}
			defer report(t, r)

			// submit PATCH request
			res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
			if err != nil {
				fatal(t, r, "Error patching TD: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			r := &record{
				assertions: statusAssertions,
			}
			defer report(t, r)

			assertStatusCode(t, r, response, http.StatusNoContent, body)
		})

		t.Run("result", func(t *testing.T) {
			r := &record{
				assertions: resultAssertions,
			}
			defer report(t, r)

			// retrieve the changed TD
			storedTD := retrieveThing(id, serverURL, t)

			// manually change attributes of the reference TD
			td["properties"] = map[string]interface{}{
				"status": map[string]interface{}{
					"forms": []map[string]interface{}{
						{"href": "https://mylamp.example.com/status"},
						{"href": "coaps://mylamp.example.com/status"},
					},
				},
			}
			// remove system-generated attributes
			delete(td, "registration")
			delete(storedTD, "registration")

			if !serializedEqual(td, storedTD) {
				t.Logf("Expected:\n%s\n Retrieved:\n%s\n", marshalPrettyJSON(td), marshalPrettyJSON(storedTD))
				fatal(t, r, "Unexpected result body; see logs.")
			}
		})
	})

	t.Run("fail to remove mandatory title", func(t *testing.T) {
		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, serverURL, t)

		// set title to null to remove it
		jsonTD := `{"title": null}`

		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			r := &record{
				assertions: requestAssertions,
			}
			defer report(t, r)

			// submit PATCH request
			res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
			if err != nil {
				fatal(t, r, "Error patching TD: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			r := &record{
				assertions: append(statusAssertions, "tdd-validation-syntactic"),
			}
			defer report(t, r)

			assertStatusCode(t, r, response, http.StatusBadRequest, body)
		})

		t.Run("response", func(t *testing.T) {
			r := &record{
				assertions: []string{"tdd-http-error-response"},
			}
			defer report(t, r)

			assertErrorResponse(t, r, response, body, true)
		})
	})
}

func TestDelete(t *testing.T) {
	defer report(t, nil)

	const (
		requestAssertions = "tdd-reg-delete"
		statusAssertions  = "tdd-reg-delete-resp"
		resultAssertions  = "tdd-reg-delete"
	)

	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	createThing(id, td, serverURL, t)

	t.Run("existing", func(t *testing.T) {
		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			r := &record{
				assertions: []string{"tdd-reg-crudl", requestAssertions},
			}
			defer report(t, r)

			// submit DELETE request
			res, err := httpDelete(serverURL + "/things/" + id)
			if err != nil {
				fatal(t, r, "Error deleting TD: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			r := &record{
				assertions: []string{statusAssertions},
			}
			defer report(t, r)

			assertStatusCode(t, r, response, http.StatusNoContent, body)
		})

		t.Run("result", func(t *testing.T) {
			r := &record{
				assertions: []string{resultAssertions},
			}
			defer report(t, r)

			// try to retrieve the deleted TD
			res, err := http.Get(serverURL + "/things/" + id)
			if err != nil {
				fatal(t, r, "Error getting TD: %s", err)
			}
			defer res.Body.Close()

			body = httpReadBody(res, t)

			t.Run("status code", func(t *testing.T) {
				assertStatusCode(t, r, res, http.StatusNotFound, body)
			})
		})
	})

	t.Run("non-existing", func(t *testing.T) {
		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			r := &record{
				assertions: []string{requestAssertions},
			}
			defer report(t, r)

			// submit DELETE request
			res, err := httpDelete(serverURL + "/things/non-exiting-td")
			if err != nil {
				fatal(t, r, "Error deleting TD: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			r := &record{
				assertions: []string{statusAssertions},
			}
			defer report(t, r)

			assertStatusCode(t, r, response, http.StatusNotFound, body)
		})
	})

}

func TestListThings(t *testing.T) {
	defer report(t, nil)

	var response *http.Response

	t.Run("submit request", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-crudl", "tdd-reg-list-method"},
		}
		defer report(t, r)

		res, err := http.Get(serverURL + "/things")
		if err != nil {
			fatal(t, r, "Error getting list of TDs: %s", err)
		}
		// defer res.Body.Close()
		response = res
	})

	t.Run("status code", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-list-method"},
		}
		defer report(t, r)

		assertStatusCode(t, r, response, http.StatusOK, nil)
	})

	t.Run("content type", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-list-resp"},
		}
		defer report(t, r)

		assertContentMediaType(t, r, response, MediaTypeJSONLD)
	})

	t.Run("payload", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-list-resp"},
		}
		defer report(t, r)

		body := httpReadBody(response, t)

		var collection []mapAny
		err := json.Unmarshal(body, &collection)
		if err != nil {
			fatal(t, r, "Error decoding page: %s", err)
		}

		for _, td := range collection {
			if td["title"] == nil || td["title"].(string) == "" {
				t.Logf("Body:\n%s", marshalPrettyJSON(td))
				fatal(t, r, "Object in array may not be a TD: no mandatory title. See logs.")
			}
		}

	})

	t.Run("anonymous td id", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-anonymous-td-identifier"},
		}
		defer report(t, r)

		fatal(t, r, "TODO")
	})

	// t.Run("http11 chunking", func(t *testing.T) {
	// 	r := &record{
	// 		assertions: []string{"tdd-reg-list-http11-chunks"},
	// 	}
	// 	defer report(t, r)

	// 	if response == nil {
	// 		fatal(t, r, "previous errors")
	// 	}

	// 	// encoding := response.Header.Get("Transfer-Encoding")
	// 	// t.Log(response.Header)

	// 	skip(t, r, "TODO")
	// })

	// t.Run("http2 streaming", func(t *testing.T) {
	// 	r := &record{
	// 		assertions: []string{"tdd-reg-list-http2-frames"},
	// 	}
	// 	defer report(t, r)

	// 	skip(t, r, "TODO")
	// })

	t.Run("pagination", func(t *testing.T) {
		r := &record{
			assertions: []string{
				"tdd-reg-list-pagination",
				"tdd-reg-list-pagination-limit",
				"tdd-reg-list-pagination-header-nextlink",
				"tdd-reg-list-pagination-header-nextlink-attr",
				"tdd-reg-list-pagination-header-canonicallink",
				"tdd-reg-list-pagination-order-default",
				"tdd-reg-list-pagination-order",
				"tdd-reg-list-pagination-order-unsupported",
				"tdd-reg-list-pagination-order-nextlink",
			},
		}
		defer report(t, r)

		fatal(t, r, "TODO")
	})
}

func testRegistrionInfo(t *testing.T, td mapAny) {

	t.Run("created", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-registrationinfo-vocab-created"},
		}
		defer report(t, r)

		regInfo, ok := td["registration"].(mapAny)
		if !ok {
			fatal(t, r, "invalid or missing registration object: %v", td["registration"])
		}

		createdStr, ok := regInfo["created"].(string)
		if !ok {
			fatal(t, r, "invalid or missing registration.created: %v", regInfo["created"])
		}
		created, err := time.Parse(time.RFC3339, createdStr)
		if err != nil {
			fatal(t, r, "invalid registration.created format: %s", err)
		}
		age := time.Since(created)
		if age < 0 && age > time.Minute {
			fatal(t, r, "registration.created is in future or too old: %s", created)
		}
	})

	t.Run("modified", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-registrationinfo-vocab-modified"},
		}
		defer report(t, r)

		regInfo, ok := td["registration"].(mapAny)
		if !ok {
			fatal(t, r, "invalid or missing registration object: %v", td["registration"])
		}

		modifiedStr, ok := regInfo["modified"].(string)
		if !ok {
			fatal(t, r, "invalid or missing registration.modified: %v", regInfo["modified"])
		}
		modified, err := time.Parse(time.RFC3339, modifiedStr)
		if err != nil {
			fatal(t, r, "invalid registration.modified format: %s", err)
		}
		age := time.Since(modified)
		if age < 0 && age > time.Minute {
			fatal(t, r, "registration.modified is in future or too old: %s", modified)
		}
	})

	t.Run("expires", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-registrationinfo-vocab-expires"},
		}
		defer report(t, r)

		fatal(t, r, "TODO")
	})

	t.Run("ttl", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-registrationinfo-vocab-ttl"},
		}
		defer report(t, r)

		fatal(t, r, "TODO")
	})

	t.Run("retrieved", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-registrationinfo-vocab-retrieved"},
		}
		defer report(t, r)

		fatal(t, r, "TODO")
	})

}
