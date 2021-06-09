package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/linksmart/thing-directory/wot"
	uuid "github.com/satori/go.uuid"
)

func TestCreateAnonymousThing(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("reg-create-anonymous-thing", "", t)
	})

	td := mockedTD("") // without ID
	b, _ := json.Marshal(td)

	// submit POST request
	res, err := http.Post(serverURL+"/things/", MediaTypeThingDescription, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("Error posting: %s", err)
	}
	defer res.Body.Close()

	body := httpReadBody(res, t)
	t.Run("status code", func(t *testing.T) {
		assertStatusCode(res.StatusCode, http.StatusCreated, body, t)
	})

	var systemGeneratedID string
	t.Run("location header", func(t *testing.T) {
		// Check if system-generated id is in response
		location, err := res.Location()
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
		writeTestResult("reg-create-thing", "", t)
	})

	t.Run("PUT", func(t *testing.T) {
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		b, _ := json.Marshal(td)

		// submit PUT request
		res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
		if err != nil {
			t.Fatalf("Error posting: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(res.StatusCode, http.StatusCreated, body, t)
		})

		t.Run("result", func(t *testing.T) {
			// retrieve the stored TD
			storedTD := retrieveThing(id, serverURL, t)

			// manually change attributes of the reference TD
			// set the system-generated attributes
			td["registration"] = storedTD["registration"]

			if !serializedEqual(td, storedTD) {
				t.Fatalf("Expected:\n%v\n Retrieved:\n%v\n", td, storedTD)
			}
		})
	})

	t.Run("PUT fail id mismatch", func(t *testing.T) {
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
			assertStatusCode(res.StatusCode, http.StatusConflict, body, t)
		})
	})

	t.Run("POST fail", func(t *testing.T) {
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
			assertStatusCode(res.StatusCode, http.StatusBadRequest, body, t)
		})
	})

}

func TestRetrieveThing(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("reg-retrieve-thing", "", t)
	})

	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	storedTD := createThing(id, td, serverURL, t)

	// submit GET request
	res, err := http.Get(serverURL + "/td/" + id)
	if err != nil {
		t.Fatalf("Error getting TD: %s", err)
	}
	defer res.Body.Close()

	body := httpReadBody(res, t)

	t.Run("status code", func(t *testing.T) {
		assertStatusCode(res.StatusCode, http.StatusOK, body, t)
	})

	t.Run("content type", func(t *testing.T) {
		assertContentMediaType(res.Header.Get("Content-Type"), MediaTypeThingDescription, t)
	})

	t.Run("result", func(t *testing.T) {
		var retrievedTD mapAny
		err = json.Unmarshal(body, &retrievedTD)
		if err != nil {
			t.Fatalf("Error decoding body: %s", err)
		}

		if !serializedEqual(td, storedTD) {
			t.Fatalf("The retrieved TD is not the same as the added one:\n Added:\n %v \n Retrieved: \n %v", td, retrievedTD)
		}
	})
}

func TestUpdateThing(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("reg-update-thing", "", t)
	})

	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	createThing(id, td, serverURL, t)

	// update an attribute
	td["title"] = "updated title"
	b, _ := json.Marshal(td)

	// submit PUT request
	res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
	if err != nil {
		t.Fatalf("Error putting TD: %s", err)
	}
	defer res.Body.Close()

	body := httpReadBody(res, t)

	t.Run("status code", func(t *testing.T) {
		assertStatusCode(res.StatusCode, http.StatusOK, body, t)
	})

	t.Run("result", func(t *testing.T) {
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
		writeTestResult("reg-partially-update-thing", "", t)
	})

	t.Run("Update title", func(t *testing.T) {
		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, serverURL, t)

		// update the title
		jsonTD := `{"title": "new title"}`

		// submit PATCH request
		res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
		if err != nil {
			t.Fatalf("Error patching TD: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(res.StatusCode, http.StatusOK, body, t)
		})

		t.Run("result", func(t *testing.T) {
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

	t.Run("Remove description", func(t *testing.T) {
		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		td["description"] = "this is a test descr"
		createThing(id, td, serverURL, t)

		// set description to null to remove it
		jsonTD := `{"description": null}`

		// submit PATCH request
		res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
		if err != nil {
			t.Fatalf("Error patching TD: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(res.StatusCode, http.StatusOK, body, t)
		})

		t.Run("result", func(t *testing.T) {
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

	t.Run("Patch properties object", func(t *testing.T) {
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

		// submit PATCH request
		res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
		if err != nil {
			t.Fatalf("Error patching TD: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(res.StatusCode, http.StatusOK, body, t)
		})

		t.Run("result", func(t *testing.T) {
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

	t.Run("Patch array", func(t *testing.T) {
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

		// submit PATCH request
		res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
		if err != nil {
			t.Fatalf("Error patching TD: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(res.StatusCode, http.StatusOK, body, t)
		})

		t.Run("result", func(t *testing.T) {
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

	t.Run("Fail removing mandatory title", func(t *testing.T) {
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

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(res.StatusCode, http.StatusBadRequest, body, t)
		})
	})
}

func TestDelete(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("reg-delete-thing", "", t)
	})

	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	createThing(id, td, serverURL, t)

	t.Run("Remove existing", func(t *testing.T) {
		// submit DELETE request
		res, err := httpDelete(serverURL + "/things/" + id)
		if err != nil {
			t.Fatalf("Error deleting TD: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(res.StatusCode, http.StatusOK, body, t)
		})

		// try to retrieve the deleted TD
		res, err = http.Get(serverURL + "/things/" + id)
		if err != nil {
			t.Fatalf("Error getting TD: %s", err)
		}
		defer res.Body.Close()

		body = httpReadBody(res, t)

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(res.StatusCode, http.StatusNotFound, body, t)
		})
	})

	t.Run("Remove non-existing", func(t *testing.T) {
		// submit DELETE request
		res, err := httpDelete(serverURL + "/things/does-not-exist")
		if err != nil {
			t.Fatalf("Error deleting TD: %s", err)
		}
		defer res.Body.Close()

		body := httpReadBody(res, t)

		t.Run("status code", func(t *testing.T) {
			assertStatusCode(res.StatusCode, http.StatusNotFound, body, t)
		})
	})

}

func TestListThings(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("reg-list-things", "", t)
	})

	t.Run("status code", func(t *testing.T) {
		res, err := http.Get(serverURL + "/things")
		if err != nil {
			t.Fatalf("Error getting list of TDs: %s", err)
		}
		defer res.Body.Close()

		assertStatusCode(res.StatusCode, http.StatusOK, nil, t)
	})

	t.Run("content type", func(t *testing.T) {
		res, err := http.Get(serverURL + "/things")
		if err != nil {
			t.Fatalf("Error getting list of TDs: %s", err)
		}
		defer res.Body.Close()

		assertContentMediaType(res.Header.Get("Content-Type"), MediaTypeJSONLD, t)
	})

	t.Run("payload", func(t *testing.T) {
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

		for _, td := range collection {
			if td["title"] == nil || td["title"].(string) == "" {
				t.Fatalf("Item in list may not be a TD: no mandatory title. Got:\n%s", marshalPrettyJSON(td))
			}
		}
	})

}

func TestMinimalValidation(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("reg-validation-minimal", "", t)
	})

	t.Run("missing context", func(t *testing.T) {
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
			assertStatusCode(res.StatusCode, http.StatusBadRequest, body, t)
		})

		var pd wot.ProblemDetails
		err = json.Unmarshal(body, &pd)
		if err != nil {
			t.Fatalf("Error decoding body: %s", err)
		}

		if pd.Status != 400 {
			t.Fatalf("Expected status set to 400, got: %d", pd.Status)
		}

		if len(pd.ValidationErrors) != 1 {
			t.Fatalf("Expected 1 error, got: %d. Body: %s", len(pd.ValidationErrors), body)
		}

		// if pd.ValidationErrors[0].Field != "(root)" { // not normative?
		// 	t.Fatalf("Expected error on root, got: %s. Body: %s", pd.ValidationErrors[0].Field, body)
		// }

		// if pd.ValidationErrors[0].Descr != "@context is required" { // not normative?
		// 	t.Fatalf("Expected error on root, got: %s. Body: %s", pd.ValidationErrors[0].Descr, body)
		// }
	})

}
