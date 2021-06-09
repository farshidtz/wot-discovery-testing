package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
	"testing"

	uuid "github.com/satori/go.uuid"
)

const (
	MediaTypeJSON             = "application/json"
	MediaTypeJSONLD           = "application/ld+json"
	MediaTypeThingDescription = "application/td+json"
	MediaTypeMergePatch       = "application/merge-patch+json"
)

func TestCreateAnonymousThing(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("create-anonymous-thing", "", t)
	})

	td := mockedTD("") // without ID
	b, _ := json.Marshal(td)

	res, err := http.Post(serverURL+"/things/", MediaTypeThingDescription, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("Error posting: %s", err)
	}
	defer res.Body.Close()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %s", err)
	}

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected response %v, got: %d. Response body: %s", http.StatusCreated, res.StatusCode, b)
	}

	// Check if system-generated id is in response
	location, err := res.Location()
	if err != nil {
		t.Fatal(err.Error())
	}
	if !strings.Contains(location.String(), "urn:uuid:") {
		t.Fatalf("System-generated ID is not a UUID. Get response location: %s\n", location)
	}

	storedTD := retrieveThing(location.String(), t)

	// manually change attributes of the reference TD
	// set the system-generated attributes
	td["id"] = storedTD["id"]
	td["registration"] = storedTD["registration"]

	if !serializedEqual(td, storedTD) {
		t.Fatalf("Posted:\n%v\n Retrieved:\n%v\n", td, storedTD)
	}

}

func TestCreateThing(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("create-thing", "", t)
	})

	t.Run("PUT", func(t *testing.T) {
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		b, _ := json.Marshal(td)

		res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
		if err != nil {
			t.Fatalf("Error posting: %s", err)
		}
		defer res.Body.Close()

		b, err = ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Error reading response body: %s", err)
		}

		if res.StatusCode != http.StatusCreated {
			t.Fatalf("Expected response %v, got: %d. Response body: %s", http.StatusCreated, res.StatusCode, b)
		}

		storedTD := retrieveThing(id, t)

		// manually change attributes of the reference TD
		// set the system-generated attributes
		td["registration"] = storedTD["registration"]

		if !serializedEqual(td, storedTD) {
			t.Fatalf("Posted:\n%v\n Retrieved:\n%v\n", td, storedTD)
		}
	})

	t.Run("PUT fail id mismatch", func(t *testing.T) {
		t.Skip("not normative")

		id := "urn:uuid:" + uuid.NewV4().String()
		anotherID := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(anotherID)
		b, _ := json.Marshal(td)

		res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
		if err != nil {
			t.Fatalf("Error posting: %s", err)
		}
		defer res.Body.Close()

		b, err = ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Error reading response body: %s", err)
		}

		if res.StatusCode != http.StatusConflict {
			t.Fatalf("Expected response %v, got: %d. Response body: %s", http.StatusConflict, res.StatusCode, b)
		}
	})

	t.Run("POST fail", func(t *testing.T) {
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		b, _ := json.Marshal(td)

		res, err := http.Post(serverURL+"/things/", MediaTypeThingDescription, bytes.NewReader(b))
		if err != nil {
			t.Fatalf("Error posting: %s", err)
		}
		defer res.Body.Close()

		b, err = ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Error reading response body: %s", err)
		}

		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected response %v, got: %d. Response body: %s", http.StatusBadRequest, res.StatusCode, b)
		}
	})

}

func TestRetrieveThing(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("retrieve-thing", "", t)
	})

	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	storedTD := createThing(id, td, t)

	res, err := http.Get(serverURL + "/td/" + id)
	if err != nil {
		t.Fatalf("Error getting TD: %s", err)
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %s", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected response %v, got: %d. Response body: %s", http.StatusOK, res.StatusCode, b)
	}

	mediaType, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		t.Fatalf("Error parsing media type: %s", err)
	}
	if mediaType != MediaTypeThingDescription {
		t.Fatalf("Expected Content-Type: %s, got %s", MediaTypeThingDescription, res.Header.Get("Content-Type"))
	}

	var retrievedTD mapAny
	err = json.Unmarshal(b, &retrievedTD)
	if err != nil {
		t.Fatalf("Error decoding body: %s", err)
	}

	if !serializedEqual(td, storedTD) {
		t.Fatalf("The retrieved TD is not the same as the added one:\n Added:\n %v \n Retrieved: \n %v", td, retrievedTD)
	}
}

func TestUpdateThing(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("update-thing", "", t)
	})

	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	createThing(id, td, t)

	td["title"] = "updated title"
	b, _ := json.Marshal(td)
	res, err := httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
	if err != nil {
		t.Fatalf("Error putting TD: %s", err)
	}
	defer res.Body.Close()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %s", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected response %v, got: %d. Reponse body: %s", http.StatusOK, res.StatusCode, b)
	}

	storedTD := retrieveThing(id, t)

	// manually change attributes of the reference TD
	// set system-generated attributes
	td["registration"] = storedTD["registration"]

	if !serializedEqual(td, storedTD) {
		t.Fatalf("Expected:\n%v\n Retrieved:\n%v\n", td, storedTD)
	}

}

func TestPatch(t *testing.T) {
	t.Cleanup(func() {
		writeTestResult("partially-update-thing", "", t)
	})

	t.Run("Update title", func(t *testing.T) {
		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, t)

		// update the title
		jsonTD := `{"title": "new title"}`

		// submit PATCH request
		res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
		if err != nil {
			t.Fatalf("Error patching TD: %s", err)
		}
		defer res.Body.Close()

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Error reading response body: %s", err)
		}

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected response %v, got: %d. Response body: %s", http.StatusOK, res.StatusCode, b)
		}

		// retrieve the changed TD
		storedTD := retrieveThing(id, t)

		// manually change attributes of the reference TD
		td["title"] = "new title"
		// set system-generated attributes
		td["registration"] = storedTD["registration"]

		if !serializedEqual(td, storedTD) {
			t.Fatalf("Expected:\n%v\n Retrieved:\n%v\n", td, storedTD)
		}
	})

	t.Run("Remove description", func(t *testing.T) {
		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		td["description"] = "this is a test descr"
		createThing(id, td, t)

		// set description to null to remove it
		jsonTD := `{"description": null}`

		// submit PATCH request
		res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
		if err != nil {
			t.Fatalf("Error patching TD: %s", err)
		}
		defer res.Body.Close()

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Error reading response body: %s", err)
		}

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected response %v, got: %d. Response body: %s", http.StatusOK, res.StatusCode, b)
		}

		// retrieve the changed TD
		storedTD := retrieveThing(id, t)

		// manually change attributes of the reference TD
		delete(td, "description")
		// set system-generated attributes
		td["registration"] = storedTD["registration"]

		if !serializedEqual(td, storedTD) {
			t.Fatalf("Posted:\n%v\n Retrieved:\n%v\n", td, storedTD)
		}
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
		createThing(id, td, t)

		// patch with new property
		jsonTD := `{"properties": {"new_property": {"forms": [{"href": "https://mylamp.example.com/new_property"}]}}}`

		// submit PATCH request
		res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
		if err != nil {
			t.Fatalf("Error patching TD: %s", err)
		}
		defer res.Body.Close()

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Error reading response body: %s", err)
		}

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected response %v, got: %d. Response body: %s", http.StatusOK, res.StatusCode, b)
		}

		// retrieve the changed TD
		storedTD := retrieveThing(id, t)

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
		createThing(id, td, t)

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

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Error reading response body: %s", err)
		}

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected response %v, got: %d. Reponse body: %s", http.StatusOK, res.StatusCode, b)
		}

		// retrieve the changed TD
		storedTD := retrieveThing(id, t)

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

	t.Run("Fail removing mandatory title", func(t *testing.T) {
		// add a new TD
		id := "urn:uuid:" + uuid.NewV4().String()
		td := mockedTD(id)
		createThing(id, td, t)

		// set title to null to remove it
		jsonTD := `{"title": null}`

		// submit PATCH request
		res, err := httpPatch(serverURL+"/things/"+id, MediaTypeMergePatch, []byte(jsonTD))
		if err != nil {
			t.Fatalf("Error patching TD: %s", err)
		}
		defer res.Body.Close()

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Error reading response body: %s", err)
		}

		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected response %v, got: %d. Response body: %s", http.StatusBadRequest, res.StatusCode, b)
		}
	})
}

func TestDelete(t *testing.T) {
	// add a new TD
	id := "urn:uuid:" + uuid.NewV4().String()
	td := mockedTD(id)
	createThing(id, td, t)

	t.Run("Remove existing", func(t *testing.T) {
		// submit DELETE request
		res, err := httpDelete(serverURL + "/things/" + id)
		if err != nil {
			t.Fatalf("Error deleting TD: %s", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Server should return %v, got instead: %d", http.StatusOK, res.StatusCode)
		}

		// try to retrieve the delete TD
		res, err = http.Get(serverURL + "/things/" + id)
		if err != nil {
			t.Fatalf("Error getting TD: %s", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected response %v, got: %d", http.StatusNotFound, res.StatusCode)
		}
	})

	t.Run("Remove non-existing", func(t *testing.T) {
		// submit DELETE request
		res, err := httpDelete(serverURL + "/things/does-not-exist")
		if err != nil {
			t.Fatalf("Error deleting TD: %s", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected response %v, got: %d", http.StatusNotFound, res.StatusCode)
		}
	})

}

// func TestGetAll(t *testing.T) {
// 	controller, testServer := setupTestHTTPServer(t)

// 	for i := 0; i < 3; i++ {
// 		// add through controller
// 		td := mockedTD("urn:example:test/thing_" + strconv.Itoa(i))
// 		_, err := controller.add(td)
// 		if err != nil {
// 			t.Fatalf("Error adding through controller: %s", err)
// 		}
// 	}

// 	t.Run("Response total", func(t *testing.T) {
// 		res, err := http.Get(testServer.URL + "/td")
// 		if err != nil {
// 			t.Fatalf("Error getting TDs: %s", err)
// 		}
// 		defer res.Body.Close()

// 		var collectionPage ThingDescriptionPage
// 		err = json.NewDecoder(res.Body).Decode(&collectionPage)
// 		if err != nil {
// 			t.Fatalf("Error decoding page: %s", err)
// 		}

// 		items, ok := collectionPage.Items.([]interface{})
// 		if !ok {
// 			t.Fatalf("Items in catalog are not TDs. Got: %v", collectionPage.Items)
// 		}
// 		if len(items) != 3 {
// 			t.Fatalf("Expected 3 items in page, got %d", len(items))
// 		}
// 		if collectionPage.Total != 3 {
// 			t.Fatalf("Expected total value of 3, got %d", collectionPage.Total)
// 		}
// 	})

// 	t.Run("Response headers", func(t *testing.T) {
// 		res, err := http.Get(testServer.URL + "/td")
// 		if err != nil {
// 			t.Fatalf("Error getting TDs: %s", err)
// 		}
// 		defer res.Body.Close()

// 		if res.StatusCode != http.StatusOK {
// 			t.Fatalf("Expected status %v, got: %d", http.StatusOK, res.StatusCode)
// 		}

// 		mediaType, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
// 		if err != nil {
// 			t.Fatalf("Error parsing media type: %s", err)
// 		}
// 		if mediaType != wot.MediaTypeJSONLD {
// 			t.Fatalf("Expected Content-Type: %s, got %s", wot.MediaTypeJSONLD, res.Header.Get("Content-Type"))
// 		}
// 	})

// 	t.Run("Response items", func(t *testing.T) {
// 		res, err := http.Get(testServer.URL + "/td")
// 		if err != nil {
// 			t.Fatalf("Error getting TDs: %s", err)
// 		}
// 		defer res.Body.Close()

// 		var collectionPage ThingDescriptionPage
// 		err = json.NewDecoder(res.Body).Decode(&collectionPage)
// 		if err != nil {
// 			t.Fatalf("Error decoding page: %s", err)
// 		}

// 		// get all through controller
// 		storedTDs, _, err := controller.list(1, 10)
// 		if err != nil {
// 			t.Fatal("Error getting list of TDs:", err.Error())
// 		}

// 		// compare response and stored
// 		for i, sd := range collectionPage.Items.([]interface{}) {
// 			if !reflect.DeepEqual(storedTDs[i], sd) {
// 				t.Fatalf("TD responded over HTTP is different with the one stored:\n Stored:\n%v\n Listed\n%v\n",
// 					storedTDs[i], sd)
// 			}
// 		}
// 	})

// 	t.Run("Filter bad JSONPath", func(t *testing.T) {
// 		res, err := http.Get(testServer.URL + "/td?jsonpath=*/id")
// 		if err != nil {
// 			t.Fatalf("Error getting TDs: %s", err)
// 		}
// 		defer res.Body.Close()

// 		b, err := ioutil.ReadAll(res.Body)
// 		if err != nil {
// 			t.Fatalf("Error reading response body: %s", err)
// 		}
// 		if res.StatusCode != http.StatusBadRequest {
// 			t.Fatalf("Expected status %v, got: %d. Response body:\n%s", http.StatusBadRequest, res.StatusCode, b)
// 		}
// 	})

// 	t.Run("Filter bad XPath", func(t *testing.T) {
// 		res, err := http.Get(testServer.URL + "/td?xpath=$[:].id")
// 		if err != nil {
// 			t.Fatalf("Error getting TDs: %s", err)
// 		}
// 		defer res.Body.Close()

// 		b, err := ioutil.ReadAll(res.Body)
// 		if err != nil {
// 			t.Fatalf("Error reading response body: %s", err)
// 		}
// 		if res.StatusCode != http.StatusBadRequest {
// 			t.Fatalf("Expected status %v, got: %d. Response body:\n%s", http.StatusBadRequest, res.StatusCode, b)
// 		}
// 	})

// 	t.Run("Filter multiple paths", func(t *testing.T) {
// 		res, err := http.Get(testServer.URL + "/td?jsonpath=$[:].id&xpath=*/id")
// 		if err != nil {
// 			t.Fatalf("Error getting TDs: %s", err)
// 		}
// 		defer res.Body.Close()

// 		b, err := ioutil.ReadAll(res.Body)
// 		if err != nil {
// 			t.Fatalf("Error reading response body: %s", err)
// 		}
// 		if res.StatusCode != http.StatusBadRequest {
// 			t.Fatalf("Expected status %v, got: %d. Response body:\n%s", http.StatusBadRequest, res.StatusCode, b)
// 		}
// 	})
// }
