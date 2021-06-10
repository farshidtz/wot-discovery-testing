package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"reflect"
	"testing"
)

type any = interface{}
type mapAny = map[string]any

// retrieveThing is a helper function to support tests unrelated to retrieval of a TD
func retrieveThing(id, serverURL string, t *testing.T) mapAny {
	t.Helper()
	res, err := http.Get(serverURL + "/things/" + id)
	if err != nil {
		t.Fatalf("Error getting TD: %s", err)
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %s", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Error retrieving test data: %d: %s", res.StatusCode, b)
	}

	var retrievedTD mapAny
	err = json.Unmarshal(b, &retrievedTD)
	if err != nil {
		t.Fatalf("Error decoding body: %s. Body:\n%s", err, b)
	}
	return retrievedTD
}

// createThing is a helper function to support tests unrelated to creation of a TD
func createThing(id string, td mapAny, serverURL string, t *testing.T) mapAny {
	t.Helper()
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
		t.Fatalf("Error creating test data: %d: %s", res.StatusCode, b)
	}

	storedTD := retrieveThing(id, serverURL, t)

	// add the system-generated attributes
	td["registration"] = storedTD["registration"]
	return td
}

// retrieveAllThings is a helper function to support tests unrelated to retrieval of all TDs
func retrieveAllThings(serverURL string, t *testing.T) []mapAny {
	t.Helper()
	res, err := http.Get(serverURL + "/things")
	if err != nil {
		t.Fatalf("Error getting TD: %s", err)
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %s", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Error retrieving test data: %d: %s", res.StatusCode, b)
	}

	var retrievedTDs []mapAny
	err = json.Unmarshal(b, &retrievedTDs)
	if err != nil {
		t.Fatalf("Error decoding body: %s", err)
	}
	return retrievedTDs
}

func serializedEqual(td1 mapAny, td2 mapAny) bool {
	// serialize to ease comparison of interfaces and concrete types
	tdBytes, _ := json.Marshal(td1)
	storedTDBytes, _ := json.Marshal(td2)

	return reflect.DeepEqual(tdBytes, storedTDBytes)
}

func httpPut(url, contentType string, b []byte) (*http.Response, error) {
	return httpRequest(http.MethodPut, url, contentType, b)
}

func httpPatch(url, contentType string, b []byte) (*http.Response, error) {
	return httpRequest(http.MethodPatch, url, contentType, b)
}

func httpDelete(url string) (*http.Response, error) {
	return httpRequest(http.MethodDelete, url, "", nil)
}

func httpRequest(method, url, contentType string, b []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func httpReadBody(res *http.Response, t *testing.T) []byte {
	t.Helper()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %s", err)
	}
	return b
}

func marshalPrettyJSON(i interface{}) string {
	b, _ := json.MarshalIndent(i, "", "\t")
	return string(b)
}

func prettifyJSON(in []byte) []byte {
	var out bytes.Buffer
	json.Indent(&out, in, "", "\t")
	return out.Bytes()
}

func assertStatusCode(got, expected int, body []byte, t *testing.T) {
	t.Helper()
	if got != expected {
		body = prettifyJSON(body)
		if len(body) == 0 {
			t.Fatalf("Expected response %d, got: %d.", expected, got)
		} else {
			t.Fatalf("Expected response %d, got: %d. Body:\n%s", expected, got, body)
		}
	}
}

func assertContentMediaType(got, expected string, t *testing.T) {
	mediaType, _, err := mime.ParseMediaType(got)
	if err != nil {
		t.Fatalf("Error parsing content media type: %s", err)
	}
	if mediaType != expected {
		t.Fatalf("Expected Content-Type: %s, got %s", expected, got)
	}
}
