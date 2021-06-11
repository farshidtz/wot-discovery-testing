package directory

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestCreateAnonymousThing(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("", "", t)
	})

	td := mockedTD("") // without ID
	b, _ := json.Marshal(td)

	var response *http.Response

	t.Run("submit request", func(t *testing.T) {
		r := &record{
			assertions: []string{"tdd-reg-create-anonymous-td", "tdd-reg-create-body"},
		}
		defer report(r, t)
		// t.Cleanup(func() {
		// 	// writeTestResult("tdd-reg-create-anonymous-td tdd-reg-create-body", "", t)
		// 	report(r, t)
		// })
		// t.Cleanup(func() { report(r, t) })

		// submit POST request
		res, err := http.Post(serverURL+"/things/", MediaTypeThingDescription, bytes.NewReader(b))
		if err != nil {
			// r.comments = fmt.Sprintf("Error posting: %s", err)
			// t.Fatal(r.comments)
			reportError(r, t, "Error posting: %s", err)
		}
		response = res
		// defer res.Body.Close()
	})

	body := httpReadBody(response, t)

	t.Run("status code", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-create-anonymous-td-resp", "", t)
		})
		assertStatusCode(response, http.StatusCreated, body, t)
	})

	var systemGeneratedID string
	t.Run("location header", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-create-anonymous-td-resp", "", t)
		})
		// Check if system-generated id is in response
		location, err := response.Location()
		if err != nil {
			t.Fatal(err.Error())
		}
		systemGeneratedID = location.String()
		if systemGeneratedID == "" {
			t.Fatalf("System-generated ID not in response. Get response location: %s", location)
		}
		if !strings.Contains(systemGeneratedID, "_:") {
			t.Fatalf("System-generated ID is not a Blank Node Identifier. Get response location: %s", location)
		}
	})

	t.Run("result", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-create-anonymous-td", "", t)
		})
		if systemGeneratedID == "" {
			t.Skip()
		}
		// retrieve the stored TD
		storedTD := retrieveThing(systemGeneratedID, serverURL, t)

		// manually change attributes of the reference TD
		// set the system-generated attributes
		td["id"] = storedTD["id"]
		td["registration"] = storedTD["registration"]

		if !serializedEqual(td, storedTD) {
			t.Fatalf("Expected:\n%v\n Retrieved:\n%v\n", td, storedTD)
		}
	})
}

func TestCreateThing(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("", "", t)
	})

	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	b, _ := json.Marshal(td)

	var response *http.Response

	t.Run("submit request", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-create-known-td tdd-reg-create-body", "", t)
		})
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
		t.Cleanup(func() {
			writeTestResult("tdd-reg-create-known-td-resp", "", t)
		})
		assertStatusCode(response, http.StatusCreated, body, t)
	})

	t.Run("result", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-create-known-td tdd-reg-create-body", "", t)
		})
		// retrieve the stored TD
		storedTD := retrieveThing(id, serverURL, t)

		// manually change attributes of the reference TD
		// set the system-generated attributes
		td["registration"] = storedTD["registration"]

		if !serializedEqual(td, storedTD) {
			t.Fatalf("Expected:\n%v\n Retrieved:\n%v\n", td, storedTD)
		}
	})

	t.Run("reject id mismatch", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("", "no relevant assertions", t)
		})
		t.SkipNow() // this is sadly not an expected normative behavior

		id := "urn:uuid:" + uuid.NewV4().String()
		anotherID := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(anotherID)
		b, _ := json.Marshal(td)

		// submit PUT request
		res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
		if err != nil {
			t.Fatalf("Error posting: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(res, http.StatusConflict, body, t)
		})
	})

	t.Run("reject POST", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("", "no relevant assertions", t)
		})
		t.SkipNow()

		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		b, _ := json.Marshal(td)

		// submit POST request
		res, err := http.Post(serverURL+"/things/", MediaTypeThingDescription, bytes.NewReader(b))
		if err != nil {
			t.Fatalf("Error posting: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(res, http.StatusBadRequest, body, t)
		})
	})

}

func TestRetrieveThing(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("", "", t)
	})

	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	storedTD := createThing(id, td, serverURL, t)

	var response *http.Response

	t.Run("submit request", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-retrieve", "", t)
		})
		// submit GET request
		res, err := http.Get(serverURL + "/td/" + id)
		if err != nil {
			t.Fatalf("Error getting TD: %s", err)
		}
		response = res
		// defer res.Body.Close()
	})

	body := httpReadBody(response, t)

	t.Run("status code", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-retrieve-resp", "", t)
		})
		assertStatusCode(response, http.StatusOK, body, t)
	})

	t.Run("content type", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-retrieve-resp", "", t)
		})
		assertContentMediaType(response, MediaTypeThingDescription, t)
	})

	t.Run("result", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-retrieve", "", t)
		})
		var retrievedTD mapAny
		err := json.Unmarshal(body, &retrievedTD)
		if err != nil {
			t.Fatalf("Error decoding body: %s", err)
		}

		if !serializedEqual(td, storedTD) {
			t.Fatalf("The retrieved TD is not the same as the added one:\n Added:\n %v \n Retrieved: \n %v", td, retrievedTD)
		}
	})

	t.Run("enriched result", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("", "TODO", t)
		})
		t.SkipNow()
	})
}

func TestUpdateThing(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("", "", t)
	})

	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	createThing(id, td, serverURL, t)

	// update an attribute
	td["title"] = "updated title"
	b, _ := json.Marshal(td)

	var response *http.Response

	t.Run("submit request", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-update-types tdd-reg-update tdd-reg-update-contenttype", "", t)
		})
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
		// t.Cleanup(func() {
		defer writeTestResult("tdd-reg-update-resp", "", t)
		// })
		assertStatusCode(response, http.StatusNoContent, body, t)
	})

	t.Run("result", func(t *testing.T) {
		// t.Cleanup(func() {
		defer writeTestResult("tdd-reg-update-types tdd-reg-update tdd-reg-update-contenttype", "", t)
		// })
		// retrieve the stored TD
		storedTD := retrieveThing(id, serverURL, t)

		// manually change attributes of the reference TD
		// set system-generated attributes
		td["registration"] = storedTD["registration"]

		if !serializedEqual(td, storedTD) {
			t.Fatalf("Expected:\n%v\n Retrieved:\n%v\n", td, storedTD)
		}
	})
}

func TestPatch(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("", "", t)
	})

	const (
		requestAssertions = "tdd-reg-update-partial tdd-reg-update-partial-partialtd tdd-reg-update-partial-contenttype"
		statusAssertions  = "tdd-reg-update-partial-resp"
		resultAssertions  = "tdd-reg-update-partial tdd-reg-update-partial-mergepatch"
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
			t.Cleanup(func() {
				writeTestResult(requestAssertions, "", t)
			})
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
			t.Cleanup(func() {
				writeTestResult(statusAssertions, "", t)
			})
			assertStatusCode(response, http.StatusOK, body, t)
		})

		t.Run("result", func(t *testing.T) {
			t.Cleanup(func() {
				writeTestResult(resultAssertions, "", t)
			})
			// retrieve the changed TD
			storedTD := retrieveThing(id, serverURL, t)

			// manually change attributes of the reference TD
			td["title"] = "new title"
			// set system-generated attributes
			td["registration"] = storedTD["registration"]

			if !serializedEqual(td, storedTD) {
				t.Fatalf("Expected:\n%v\n Retrieved:\n%v\n", td, storedTD)
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
			t.Cleanup(func() {
				writeTestResult(requestAssertions, "", t)
			})
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
			t.Cleanup(func() {
				writeTestResult(statusAssertions, "", t)
			})
			assertStatusCode(response, http.StatusOK, body, t)
		})

		t.Run("result", func(t *testing.T) {
			t.Cleanup(func() {
				writeTestResult(resultAssertions, "", t)
			})
			// retrieve the changed TD
			storedTD := retrieveThing(id, serverURL, t)

			// manually change attributes of the reference TD
			delete(td, "description")
			// set system-generated attributes
			td["registration"] = storedTD["registration"]

			if !serializedEqual(td, storedTD) {
				t.Fatalf("Posted:\n%v\n Retrieved:\n%v\n", td, storedTD)
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
			t.Cleanup(func() {
				writeTestResult(requestAssertions, "", t)
			})
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
			t.Cleanup(func() {
				writeTestResult(statusAssertions, "", t)
			})
			assertStatusCode(response, http.StatusOK, body, t)
		})

		t.Run("result", func(t *testing.T) {
			t.Cleanup(func() {
				writeTestResult(resultAssertions, "", t)
			})
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
			// set system-generated attributes
			td["registration"] = storedTD["registration"]

			if !serializedEqual(td, storedTD) {
				t.Fatalf("Expected:\n%v\n Retrieved:\n%v\n", td, storedTD)
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
			t.Cleanup(func() {
				writeTestResult(requestAssertions, "", t)
			})
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
			t.Cleanup(func() {
				writeTestResult(statusAssertions, "", t)
			})
			assertStatusCode(response, http.StatusOK, body, t)
		})

		t.Run("result", func(t *testing.T) {
			t.Cleanup(func() {
				writeTestResult(resultAssertions, "", t)
			})
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
			// set system-generated attributes
			td["registration"] = storedTD["registration"]

			if !serializedEqual(td, storedTD) {
				t.Fatalf("Expected:\n%v\n Retrieved:\n%v\n", td, storedTD)
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
			t.Cleanup(func() {
				writeTestResult(requestAssertions, "", t)
			})
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
			t.Cleanup(func() {
				writeTestResult(statusAssertions+" td-validation-syntactic", "", t)
			})
			assertStatusCode(response, http.StatusBadRequest, body, t)
		})
	})
}

func TestDelete(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("", "", t)
	})

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
			t.Cleanup(func() {
				writeTestResult(requestAssertions, "", t)
			})
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
			t.Cleanup(func() {
				writeTestResult(statusAssertions, "", t)
			})
			assertStatusCode(response, http.StatusNoContent, body, t)
		})

		t.Run("result", func(t *testing.T) {
			t.Cleanup(func() {
				writeTestResult(resultAssertions, "", t)
			})
			// try to retrieve the deleted TD
			res, err := http.Get(serverURL + "/things/" + id)
			if err != nil {
				t.Fatalf("Error getting TD: %s", err)
			}
			defer res.Body.Close()

			body = httpReadBody(res, t)

			t.Run("status code", func(t *testing.T) {
				assertStatusCode(res, http.StatusNotFound, body, t)
			})
		})
	})

	t.Run("non-existing", func(t *testing.T) {
		var response *http.Response

		t.Run("submit request", func(t *testing.T) {
			t.Cleanup(func() {
				writeTestResult(requestAssertions, "", t)
			})
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
			t.Cleanup(func() {
				writeTestResult(statusAssertions, "", t)
			})
			assertStatusCode(response, http.StatusNotFound, body, t)
		})
	})

}

func TestListThings(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("", "", t)
	})

	var response *http.Response

	t.Run("submit request", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-list-method", "", t)
		})
		res, err := http.Get(serverURL + "/things")
		if err != nil {
			t.Fatalf("Error getting list of TDs: %s", err)
		}
		// defer res.Body.Close()
		response = res
	})

	t.Run("status code", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-list-method", "", t)
		})
		assertStatusCode(response, http.StatusOK, nil, t)
	})

	t.Run("content type", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-list-resp", "", t)
		})
		assertContentMediaType(response, MediaTypeJSONLD, t)
	})

	t.Run("payload", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-list-resp", "", t)
		})

		body := httpReadBody(response, t)

		var collection []mapAny
		err := json.Unmarshal(body, &collection)
		if err != nil {
			t.Fatalf("Error decoding page: %s", err)
		}

		for _, td := range collection {
			if td["title"] == nil || td["title"].(string) == "" {
				t.Fatalf("Item in list may not be a TD: no mandatory title. Got:\n%s", marshalPrettyJSON(td))
			}
		}
	})

	t.Run("http11 chunking", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-list-http11", "TODO", t)
		})
		t.SkipNow()
	})

	t.Run("http2 streaming", func(t *testing.T) {
		t.Cleanup(func() {
			writeTestResult("tdd-reg-list-http2", "TODO", t)
		})
		t.SkipNow()
	})

	t.Run("pagination", func(t *testing.T) {
		t.Cleanup(func() {
			// tdd-reg-list-pagination tdd-reg-list-pagination-limit tdd-reg-list-pagination-header-nextlink tdd-reg-list-pagination-header-nextlink-attr tdd-reg-list-pagination-header-canonicallink tdd-reg-list-pagination-order-default tdd-reg-list-pagination-order tdd-reg-list-pagination-order-unsupported tdd-reg-list-pagination-order-nextlink
			writeTestResult("", "TODO", t)
		})
		t.SkipNow()
	})

}

func TestMinimalValidation(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("", "TODO: validate as part of POST/PUT/PATCH", t)
	})
	t.SkipNow()

	t.Run("reject missing context", func(t *testing.T) {
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)

		// remove the context field
		delete(td, "@context")

		b, _ := json.Marshal(td)

		// submit with PUT request
		res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
		if err != nil {
			t.Fatalf("Error posting: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(res, http.StatusBadRequest, body, t)
		})

		var problemDetails map[string]any
		err = json.Unmarshal(body, &problemDetails)
		if err != nil {
			t.Fatalf("Error decoding body: %s", err)
		}

		problemDetailsStatus, ok := problemDetails["status"].(float64) // JSON number is float64
		if !ok {
			t.Fatalf("Problem Details: missing status field. Body: %s", body)
		}
		if problemDetailsStatus != 400 {
			t.Fatalf("Problem Details: expected status 400 in body, got: %f", problemDetailsStatus)
		}

		validationErrors, ok := problemDetails["validationErrors"].([]any)
		if !ok {
			t.Fatalf("Problem Details: missing validationErrors field. Body: %s", body)
		}
		if len(validationErrors) != 1 {
			t.Fatalf("Problem Details: expected 1 validation error, got: %d. Body: %s", len(validationErrors), body)
		}

		// if pd.ValidationErrors[0].Field != "(root)" { // not normative?
		// 	t.Fatalf("Expected error on root, got: %s. Body: %s", pd.ValidationErrors[0].Field, body)
		// }

		// if pd.ValidationErrors[0].Descr != "@context is required" { // not normative?
		// 	t.Fatalf("Expected error on root, got: %s. Body: %s", pd.ValidationErrors[0].Descr, body)
		// }
	})

}
