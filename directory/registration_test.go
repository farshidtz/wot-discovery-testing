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

	td := mockedTD("") // without ID
	b, _ := json.Marshal(td)

	var response *http.Response

	t.Run("submit request", func(t *testing.T) {
		defer report(t,
			"tdd-reg-crudl",
			"tdd-reg-create-anonymous-td",
			"tdd-reg-create-body",
			"tdd-reg-create-contenttype")

		// submit POST request
		res, err := http.Post(serverURL+"/things/", MediaTypeThingDescription, bytes.NewReader(b))
		if err != nil {
			t.Fatalf("Error posting: %s", err)
		}
		response = res
		// defer res.Body.Close()
	})

	body := httpReadBody(response, t)

	t.Run("status code", func(t *testing.T) {
		defer report(t, "tdd-reg-create-anonymous-td-resp")
		assertStatusCode(t, response, http.StatusCreated, body)
	})

	var systemGeneratedID string
	t.Run("location header", func(t *testing.T) {
		defer report(t,
			"tdd-reg-create-anonymous-td-resp",
			"tdd-reg-anonymous-td-local-id",
		)

		// Check if system-generated id is in response
		location, err := response.Location()
		if err != nil {
			t.Fatalf(err.Error())
		}
		systemGeneratedID = location.String()
		if systemGeneratedID == "" {
			t.Fatalf("System-generated ID not in response. Got location header: %s", location)
		}
		_, err = url.ParseRequestURI(systemGeneratedID)
		if err != nil {
			t.Fatalf("System-generated ID not in a valid URI. Got: %s", location)
		}
		if !strings.Contains(systemGeneratedID, "urn:uuid:") {
			t.Fatalf("System-generated ID doesn't have URN UUID scheme. Got: %s", location)
		}
	})

	t.Run("result", func(t *testing.T) {
		defer report(t, "tdd-reg-create-anonymous-td")

		if systemGeneratedID == "" {
			t.Fatalf("previous errors")
		}

		// retrieve the stored TD
		storedTD := retrieveThing(systemGeneratedID, serverURL, t)

		// remove system-generated attributes
		delete(td, "registration")
		delete(storedTD, "registration")
		delete(storedTD, "id")

		if !serializedEqual(td, storedTD) {
			t.Fatalf("Expected:\n%v\nRetrieved:\n%v\n", marshalPrettyJSON(td), marshalPrettyJSON(storedTD))
		}
	})

	t.Run("registration info", func(t *testing.T) {
		defer report(t, "tdd-reg-anonymous-td-identifier")

		// retrieve the stored TD
		storedTD := retrieveThing(systemGeneratedID, serverURL, t)
		// get the ID. This should pass
		getID(t, storedTD)
	})

	// reject PUT of anonymous TD
	t.Run("reject PUT", func(t *testing.T) {
		defer report(t, "tdd-reg-create-known-vs-anonymous")

		td := mockedTD("") // no id
		b, _ := json.Marshal(td)

		// submit PUT request
		res, err := httpPut(serverURL+"/things/", MediaTypeThingDescription, b)
		if err != nil {
			t.Fatalf("Error putting: %s", err)
		}
		defer res.Body.Close()

		if res.StatusCode < 400 && res.StatusCode >= 500 {
			t.Fatalf("Anonymous TD submission with PUT not rejected. Got status: %d", res.StatusCode)
		}
	})

	t.Run("reject invalid", func(t *testing.T) {
		td := mockedTD("")  // no id
		delete(td, "title") // remove the mandatory field

		b, _ := json.Marshal(td)

		// submit POST request
		res, err := http.Post(serverURL+"/things/", MediaTypeThingDescription, bytes.NewReader(b))
		if err != nil {
			t.Fatalf("Error posting: %s", err)
		}
		defer res.Body.Close()

		body = httpReadBody(res, t)

		t.Run("status", func(t *testing.T) {
			defer report(t, "tdd-validation-syntactic")

			assertStatusCode(t, res, http.StatusBadRequest, nil)
		})

		t.Run("response", func(t *testing.T) {
			defer report(t, "tdd-http-error-response")

			assertErrorResponse(t, res, body)
		})

		t.Run("validation", func(t *testing.T) {
			defer report(t,
				"tdd-validation-result",
				"tdd-validation-response",
			)

			assertValidationResponse(t, res, body)
		})
	})
}

func TestCreateThing(t *testing.T) {

	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	b, _ := json.Marshal(td)

	var response *http.Response

	t.Run("submit request", func(t *testing.T) {
		defer report(t, "tdd-reg-crudl",
			"tdd-reg-create-known-td",
			"tdd-reg-create-body",
			"tdd-reg-create-contenttype",
		)

		// submit PUT request
		res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
		if err != nil {
			t.Fatalf("Error posting: %s", err)
		}
		response = res
		// defer res.Body.Close()
	})

	body := httpReadBody(response, t)

	t.Run("status code", func(t *testing.T) {
		defer report(t, "tdd-reg-create-known-td-resp")
		assertStatusCode(t, response, http.StatusCreated, body)
	})

	t.Run("result", func(t *testing.T) {
		defer report(t,
			"tdd-reg-create-known-td",
			"tdd-reg-create-body",
		)

		// retrieve the stored TD
		storedTD := retrieveThing(id, serverURL, t)

		// remove system-generated attributes
		delete(td, "registration")
		delete(storedTD, "registration")

		if !serializedEqual(td, storedTD) {
			t.Fatalf("Expected:\n%v\nRetrieved:\n%v\n", marshalPrettyJSON(td), marshalPrettyJSON(storedTD))
		}
	})

	// t.Run("registration info", func(t *testing.T) {
	// 	// retrieve the stored TD
	// 	storedTD := retrieveThing(id, serverURL, t)

	// 	testRegistrionInfo(t, storedTD)
	// })

	t.Run("reject id mismatch", func(t *testing.T) {
		defer report(t)
		t.Skipf("no relevant assertions")

		id := "urn:uuid:" + uuid.NewV4().String()
		anotherID := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(anotherID)
		b, _ := json.Marshal(td)

		var response *http.Response

		t.Run("status code", func(t *testing.T) {
			defer report(t)

			// submit PUT request
			res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
			if err != nil {
				t.Fatalf("Error posting: %s", err)
			}
			response = res
			// defer res.Body.Close()
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			defer report(t)
			assertStatusCode(t, response, http.StatusConflict, body)
		})
	})

	t.Run("reject invalid", func(t *testing.T) {
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		delete(td, "title") // remove the mandatory field

		b, _ := json.Marshal(td)

		// submit PUT request
		res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
		if err != nil {
			t.Fatalf("Error putting: %s", err)
		}
		defer res.Body.Close()

		body = httpReadBody(res, t)

		t.Run("status", func(t *testing.T) {
			defer report(t, "tdd-validation-syntactic")

			assertStatusCode(t, res, http.StatusBadRequest, body)
		})

		t.Run("response", func(t *testing.T) {
			defer report(t, "tdd-http-error-response")
			assertErrorResponse(t, res, body)
		})

		t.Run("validation", func(t *testing.T) {
			defer report(t, "tdd-validation-result", "tdd-validation-response")
			assertValidationResponse(t, res, body)
		})
	})

	// reject POST with id: not disallowed explicitly
	// t.Run("reject POST", func(t *testing.T) {
	// 	defer report(t, "tdd-reg-create-known-vs-anonymous")

	// 	id := "urn:uuid:" + uuid.NewV4().String()
	// 	td := mockedTD(id)
	// 	b, _ := json.Marshal(td)

	// 	// submit POST request
	// 	res, err := http.Post(serverURL+"/things/", MediaTypeThingDescription, bytes.NewReader(b))
	// 	if err != nil {
	// 		t.Fatalf( "Error posting: %s", err)
	// 	}
	// 	defer res.Body.Close()

	// 	body := httpReadBody(res, t)

	// 	t.Run("status code", func(t *testing.T) {
	// 		assertStatusCode(t, res, http.StatusBadRequest, body)
	// 	})
	// })

}

func TestRetrieveThing(t *testing.T) {

	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	createThing(id, td, serverURL, t)

	var response *http.Response

	t.Run("submit request", func(t *testing.T) {
		defer report(t,
			"tdd-reg-crudl",
			"tdd-reg-retrieve",
		)

		// submit GET request
		res, err := http.Get(serverURL + "/things/" + id)
		if err != nil {
			t.Fatalf("Error getting TD: %s", err)
		}
		response = res
		// defer res.Body.Close()
	})

	body := httpReadBody(response, t)

	t.Run("status code", func(t *testing.T) {
		defer report(t, "tdd-reg-retrieve-resp")
		assertStatusCode(t, response, http.StatusOK, body)
	})

	t.Run("content type", func(t *testing.T) {
		defer report(t, "tdd-reg-retrieve-resp")
		assertContentMediaType(t, response, MediaTypeThingDescription)
	})

	t.Run("result", func(t *testing.T) {
		defer report(t, "tdd-reg-retrieve")

		var retrievedTD mapAny
		err := json.Unmarshal(body, &retrievedTD)
		if err != nil {
			t.Fatalf("Error decoding body: %s", err)
		}

		// remove system-generated attributes
		delete(retrievedTD, "registration")

		if !serializedEqual(td, retrievedTD) {
			t.Fatalf("Expected:\n%v\nRetrieved:\n%v", marshalPrettyJSON(td), marshalPrettyJSON(retrievedTD))
		}
	})

	t.Run("registration info", func(t *testing.T) {
		// retrieve the stored TD
		storedTD := retrieveThing(id, serverURL, t)

		testRegistrionInfo(t, storedTD)
	})

	// t.Run("anonymous td id", func(t *testing.T) {
	// 	defer report(t, "tdd-reg-anonymous-td-identifier")

	// 	t.Skipf( "Tested under TestCreateAnonymousThing")
	// })

	t.Run("HEAD", func(t *testing.T) {
		defer report(t, "tdd-http-head")

		res, err := httpRequest(http.MethodHead, serverURL+"/things/"+id, "", nil)
		if err != nil {
			t.Fatalf("Error making HEAD request: %s", err)
		}
		body := httpReadBody(res, t)
		assertStatusCode(t, res, http.StatusOK, body)
	})
}

func TestUpdateThing(t *testing.T) {

	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	createThing(id, td, serverURL, t)

	// update an attribute
	td["title"] = "updated title"
	b, _ := json.Marshal(td)

	var response *http.Response

	t.Run("submit request", func(t *testing.T) {
		defer report(t,
			"tdd-reg-crudl",
			"tdd-reg-update-types",
			"tdd-reg-update",
			"tdd-reg-update-contenttype",
		)

		// submit PUT request
		res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
		if err != nil {
			t.Fatalf("Error putting TD: %s", err)
		}
		response = res
		// defer res.Body.Close()
	})

	body := httpReadBody(response, t)

	t.Run("status code", func(t *testing.T) {
		defer report(t, "tdd-reg-update-resp")
		assertStatusCode(t, response, http.StatusNoContent, body)
	})

	t.Run("result", func(t *testing.T) {
		defer report(t, "tdd-reg-update-types", "tdd-reg-update", "tdd-reg-update-contenttype")

		// retrieve the stored TD
		storedTD := retrieveThing(id, serverURL, t)

		// remove system-generated attributes
		delete(td, "registration")
		delete(storedTD, "registration")

		if !serializedEqual(td, storedTD) {
			t.Logf("Expected:\n%v\n Retrieved:\n%v\n", marshalPrettyJSON(td), marshalPrettyJSON(storedTD))
			t.Fatalf("Unexpected result body; see logs.")
		}
	})

	t.Run("reject invalid", func(t *testing.T) {
		delete(td, "title") // remove the mandatory field

		b, _ := json.Marshal(td)

		// submit PUT request
		res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
		if err != nil {
			t.Fatalf("Error putting: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		t.Run("status", func(t *testing.T) {
			defer report(t, "tdd-validation-syntactic")

			assertStatusCode(t, res, http.StatusBadRequest, body)
		})

		t.Run("response", func(t *testing.T) {
			defer report(t, "tdd-http-error-response")

			assertErrorResponse(t, res, body)
		})

		t.Run("validation", func(t *testing.T) {
			defer report(t, "tdd-validation-result", "tdd-validation-response")
			assertValidationResponse(t, res, body)
		})
	})
}

func TestPatch(t *testing.T) {

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
			defer report(t, requestAssertions...)

			// submit PATCH request
			res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
			if err != nil {
				t.Fatalf("Error patching TD: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			defer report(t, statusAssertions...)

			assertStatusCode(t, response, http.StatusNoContent, body)
		})

		t.Run("result", func(t *testing.T) {
			defer report(t, resultAssertions...)

			// retrieve the changed TD
			storedTD := retrieveThing(id, serverURL, t)

			// manually change attributes of the reference TD
			td["title"] = "new title"
			// remove system-generated attributes
			delete(td, "registration")
			delete(storedTD, "registration")

			if !serializedEqual(td, storedTD) {
				t.Fatalf("Expected:\n%s\n Retrieved:\n%s\n", marshalPrettyJSON(td), marshalPrettyJSON(storedTD))
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
			defer report(t, requestAssertions...)

			// submit PATCH request
			res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
			if err != nil {
				t.Fatalf("Error patching TD: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			defer report(t, statusAssertions...)
			assertStatusCode(t, response, http.StatusNoContent, body)
		})

		t.Run("result", func(t *testing.T) {
			defer report(t, resultAssertions...)

			// retrieve the changed TD
			storedTD := retrieveThing(id, serverURL, t)

			// manually change attributes of the reference TD
			delete(td, "description")
			// remove system-generated attributes
			delete(td, "registration")
			delete(storedTD, "registration")

			if !serializedEqual(td, storedTD) {
				t.Fatalf("Expected:\n%s\n Retrieved:\n%s\n", marshalPrettyJSON(td), marshalPrettyJSON(storedTD))
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
			defer report(t, requestAssertions...)

			// submit PATCH request
			res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
			if err != nil {
				t.Fatalf("Error patching TD: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			defer report(t, statusAssertions...)

			assertStatusCode(t, response, http.StatusNoContent, body)
		})

		t.Run("result", func(t *testing.T) {
			defer report(t, resultAssertions...)

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
				t.Fatalf("Unexpected result body; see logs.")
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
			defer report(t, requestAssertions...)

			// submit PATCH request
			res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
			if err != nil {
				t.Fatalf("Error patching TD: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			defer report(t, statusAssertions...)

			assertStatusCode(t, response, http.StatusNoContent, body)
		})

		t.Run("result", func(t *testing.T) {
			defer report(t, resultAssertions...)

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
				t.Fatalf("Unexpected result body; see logs.")
			}
		})
	})

	t.Run("reject invalid", func(t *testing.T) {
		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, serverURL, t)

		// set title to null to remove it
		jsonTD := `{"title": null}`

		// submit PATCH request
		res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
		if err != nil {
			t.Fatalf("Error patching TD: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		t.Run("status", func(t *testing.T) {
			defer report(t, "tdd-validation-syntactic")

			assertStatusCode(t, res, http.StatusBadRequest, body)
		})

		t.Run("response", func(t *testing.T) {
			defer report(t, "tdd-http-error-response")

			assertErrorResponse(t, res, body)
		})

		t.Run("validation", func(t *testing.T) {
			defer report(t, "tdd-validation-result", "tdd-validation-response")

			assertValidationResponse(t, res, body)
		})
	})
}

func TestDelete(t *testing.T) {

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
			defer report(t, "tdd-reg-crudl", requestAssertions)

			// submit DELETE request
			res, err := httpDelete(serverURL + "/things/" + id)
			if err != nil {
				t.Fatalf("Error deleting TD: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			defer report(t, statusAssertions)

			assertStatusCode(t, response, http.StatusNoContent, body)
		})

		t.Run("result", func(t *testing.T) {
			defer report(t, resultAssertions)

			// try to retrieve the deleted TD
			res, err := http.Get(serverURL + "/things/" + id)
			if err != nil {
				t.Fatalf("Error getting TD: %s", err)
			}
			defer res.Body.Close()

			body = httpReadBody(res, t)

			t.Run("status code", func(t *testing.T) {
				assertStatusCode(t, res, http.StatusNotFound, body)
			})
		})
	})

	t.Run("non-existing", func(t *testing.T) {
		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			defer report(t, requestAssertions)

			// submit DELETE request
			res, err := httpDelete(serverURL + "/things/non-exiting-td")
			if err != nil {
				t.Fatalf("Error deleting TD: %s", err)
			}
			// defer res.Body.Close()
			response = res
		})

		body := httpReadBody(response, t)

		t.Run("status code", func(t *testing.T) {
			defer report(t, statusAssertions)
			assertStatusCode(t, response, http.StatusNotFound, body)
		})
	})

}

func TestListThings(t *testing.T) {

	var response *http.Response
	var body []byte

	tag := uuid.NewV4().String()
	t.Run("submit request", func(t *testing.T) {
		defer report(t, "tdd-reg-crudl", "tdd-reg-list-method")

		for i := 0; i < 3; i++ {
			id := "urn:uuid:" + uuid.NewV4().String()
			td := mockedTD(id)
			// tag the TDs to find later
			td["tag"] = tag
			createThing(id, td, serverURL, t)
		}

		res, err := http.Get(serverURL + "/things")
		if err != nil {
			t.Fatalf("Error getting list of TDs: %s", err)
		}
		// defer res.Body.Close()
		body = httpReadBody(res, t)
		response = res
	})

	t.Run("status code", func(t *testing.T) {
		defer report(t, "tdd-reg-list-method")

		assertStatusCode(t, response, http.StatusOK, body)
	})

	t.Run("content type", func(t *testing.T) {
		defer report(t, "tdd-reg-list-resp")
		assertContentMediaType(t, response, MediaTypeJSONLD)
	})

	t.Run("payload", func(t *testing.T) {
		defer report(t, "tdd-reg-list-resp")

		var collection []mapAny
		err := json.Unmarshal(body, &collection)
		if err != nil {
			t.Fatalf("Error decoding page: %s", err)
		}

		if len(collection) == 0 {
			t.Fatalf("Unexpected empty collection.")
		}

		var listedTDs []mapAny
		for _, td := range collection {
			if td["title"] == nil || td["title"].(string) == "" {
				t.Fatalf("Object in array may not be a TD: no mandatory title. Body:\n%s", marshalPrettyJSON(td))
			}
			if td["tag"] != nil && td["tag"].(string) == tag {
				listedTDs = append(listedTDs, td)
			}
		}

		if len(listedTDs) != 3 {
			t.Fatalf("Unexpected items in collection: %d. Expected 3 with tag: %s", len(listedTDs), tag)
		}
	})

	t.Run("registration info", func(t *testing.T) {
		defer report(t, "tdd-reg-list-resp")

		var collection []mapAny
		err := json.Unmarshal(body, &collection)
		if err != nil {
			t.Fatalf("Error decoding page: %s", err)
		}

		if len(collection) == 0 {
			t.Fatalf("Unexpected empty collection.")
		}

		// just test the first TD
		testRegistrionInfo(t, collection[0])
	})

	t.Run("anonymous td id", func(t *testing.T) {
		defer report(t, "tdd-reg-anonymous-td-identifier")

		// add an anonymous TD
		createdTD := mockedTD("") // no id
		// tag the TDs to find later
		tag2 := uuid.NewV4().String()
		createdTD["tag"] = tag2
		createThing("", createdTD, serverURL, t)

		// submit the request
		res, err := http.Get(serverURL + "/things")
		if err != nil {
			t.Fatalf("Error getting list of TDs: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		var collection []mapAny
		err = json.Unmarshal(body, &collection)
		if err != nil {
			t.Fatalf("Error decoding page: %s", err)
		}

		if len(collection) == 0 {
			t.Fatalf("Unexpected empty collection.")
		}

		var found bool
		for _, td := range collection {
			if td["tag"] != nil && td["tag"].(string) == tag2 {
				// try to get the ID. This should pass
				getID(t, td)
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Could not find the created anonymous TD with tag: %s", tag2)
		}
	})

	// t.Run("http11 chunking", func(t *testing.T) {
	// 	defer report(t, "tdd-reg-list-http11-chunks")

	// 	if response == nil {
	// 		t.Fatalf( "previous errors")
	// 	}

	// 	// encoding := response.Header.Get("Transfer-Encoding")
	// 	// t.Log(response.Header)

	// 	t.Skipf( "TODO")
	// })

	// t.Run("http2 streaming", func(t *testing.T) {
	// 	defer report(t, "tdd-reg-list-http2-frames")

	// 	t.Skipf( "TODO")
	// })

	t.Run("pagination", func(t *testing.T) {
		defer report(t,
			"tdd-reg-list-pagination",
			"tdd-reg-list-pagination-limit",
			"tdd-reg-list-pagination-header-nextlink",
			"tdd-reg-list-pagination-header-nextlink-attr",
			"tdd-reg-list-pagination-header-canonicallink",
			"tdd-reg-list-pagination-order-default",
			"tdd-reg-list-pagination-order",
			"tdd-reg-list-pagination-order-unsupported",
			"tdd-reg-list-pagination-order-nextlink",
		)

		t.Skipf("TODO")
	})

	t.Run("HEAD", func(t *testing.T) {
		defer report(t, "tdd-http-head")

		res, err := httpRequest(http.MethodHead, serverURL+"/things", "", nil)
		if err != nil {
			t.Fatalf("Error making HEAD request: %s", err)
		}
		body := httpReadBody(res, t)
		assertStatusCode(t, res, http.StatusOK, body)
	})
}

func testRegistrionInfo(t *testing.T, td mapAny) {

	t.Run("created", func(t *testing.T) {
		defer report(t, "tdd-registrationinfo-vocab-created")

		regInfo, ok := td["registration"].(mapAny)
		if !ok {
			t.Fatalf("invalid or missing registration object: %v", td["registration"])
		}

		createdStr, ok := regInfo["created"].(string)
		if !ok {
			t.Fatalf("invalid or missing registration.created: %v", regInfo["created"])
		}
		created, err := time.Parse(time.RFC3339, createdStr)
		if err != nil {
			t.Fatalf("invalid registration.created format: %s", err)
		}
		age := time.Since(created)
		if age < 0 && age > time.Minute {
			t.Fatalf("registration.created is in future or too old: %s", created)
		}
	})

	t.Run("modified", func(t *testing.T) {
		defer report(t, "tdd-registrationinfo-vocab-modified")

		regInfo, ok := td["registration"].(mapAny)
		if !ok {
			t.Fatalf("invalid or missing registration object: %v", td["registration"])
		}

		modifiedStr, ok := regInfo["modified"].(string)
		if !ok {
			t.Fatalf("invalid or missing registration.modified: %v", regInfo["modified"])
		}
		modified, err := time.Parse(time.RFC3339, modifiedStr)
		if err != nil {
			t.Fatalf("invalid registration.modified format: %s", err)
		}
		age := time.Since(modified)
		if age < 0 && age > time.Minute {
			t.Fatalf("registration.modified is in future or too old: %s", modified)
		}
	})

	t.Run("expires", func(t *testing.T) {
		defer report(t, "tdd-registrationinfo-vocab-expires")

		t.Skipf("TODO")
	})

	t.Run("ttl", func(t *testing.T) {
		defer report(t, "tdd-registrationinfo-vocab-ttl")

		t.Skipf("TODO")
	})

	t.Run("retrieved", func(t *testing.T) {
		defer report(t, "tdd-registrationinfo-vocab-retrieved")

		t.Skipf("TODO")
	})

	t.Run("purge expired", func(t *testing.T) {
		defer report(t, "tdd-registrationinfo-expiry-purge")

		t.Skipf("TODO")
	})

}
