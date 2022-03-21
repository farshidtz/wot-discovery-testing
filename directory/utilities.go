package directory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"reflect"
	"testing"

	"github.com/r3labs/sse/v2"
	"gopkg.in/cenkalti/backoff.v1"
)

type any = interface{}
type mapAny = map[string]any

func mockedTD(id string) mapAny {
	var td = mapAny{
		"@context": "https://www.w3.org/2019/wot/td/v1",
		"title":    "example thing",
		"security": []string{"nosec_sc"},
		"securityDefinitions": mapAny{
			"nosec_sc": map[string]string{
				"scheme": "nosec",
			},
		},
	}
	if id != "" {
		td["id"] = id
	}
	return td
}

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
func createThing(id string, td mapAny, serverURL string, t *testing.T) {
	t.Helper()
	b, _ := json.Marshal(td)

	var res *http.Response
	var err error
	if id == "" { // anonymous TD
		res, err = http.Post(serverURL+"/things", MediaTypeThingDescription, bytes.NewReader(b))
		if err != nil {
			t.Fatalf("Error posting: %s", err)
		}
	} else {
		res, err = httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
		if err != nil {
			t.Fatalf("Error posting: %s", err)
		}
	}
	defer res.Body.Close()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %s", err)
	}

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Error creating test data: %d: %s", res.StatusCode, b)
	}

	// storedTD := retrieveThing(id, serverURL, t)

	// add the system-generated attributes
	// td["registration"] = storedTD["registration"]
	// return td
}

// updateThing is a helper function to support tests unrelated to updating the TD
func updateThing(id string, td mapAny, serverURL string, t *testing.T) {
	t.Helper()
	b, _ := json.Marshal(td)

	var res *http.Response
	var err error

	res, err = httpPut(serverURL+"/things/"+id, MediaTypeThingDescription, b)
	if err != nil {
		t.Fatalf("Error updateing: %s", err)
	}

	defer res.Body.Close()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %s", err)
	}

	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Error updating test data: %d: %s", res.StatusCode, b)
	}

}

// deleteThing is a helper function to support tests unrelated to deleting the TD
func deleteThing(id string, serverURL string, t *testing.T) {
	t.Helper()
	var res *http.Response
	var err error

	res, err = httpDelete(serverURL + "/things/" + id)
	if err != nil {
		t.Fatalf("Error updateing: %s", err)
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %s", err)
	}

	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Error updating test data: %d: %s", res.StatusCode, b)
	}
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

func serializedEqual(td1, td2 mapAny) bool {
	// ignore context for now
	// See https://github.com/w3c/wot-discovery/issues/291
	delete(td1, "@context")
	delete(td2, "@context")

	// serialize to ease comparison of interfaces and concrete types
	tdBytes, _ := json.Marshal(td1)
	td2Bytes, _ := json.Marshal(td2)

	return reflect.DeepEqual(tdBytes, td2Bytes)
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
	if res == nil {
		t.FailNow()
	}

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

func assertStatusCode(t *testing.T, res *http.Response, expected int, body []byte) {
	t.Helper()
	if res == nil {
		t.Fatalf("previous errors")
	}
	got := res.StatusCode
	if got != expected {
		body = prettifyJSON(body)
		if len(body) > 0 {
			t.Logf("Body: %s", body)
		}
		t.Fatalf("Expected status %d, got: %d", expected, got)
	}
}

func assertContentMediaType(t *testing.T, res *http.Response, expected string) {
	t.Helper()
	if res == nil {
		t.Fatalf("previous errors")
	}
	got := res.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(got)
	if err != nil {
		t.Fatalf("Error parsing content media type: %s", err)
	}
	if mediaType != expected {
		t.Fatalf("Expected Content-Type: %s, got %s", expected, got)
	}
}

func getID(t *testing.T, td mapAny) string {
	t.Helper()
	var id string
	if _, found := td["id"]; found {
		if _, ok := td["id"].(string); ok {
			id = td["id"].(string)
		}
	}
	if id == "" {
		t.Fatalf("No ID in TD: %s", marshalPrettyJSON(td))
	}
	return id
}

type httpError struct {
	message string
	code    int
}

func (e *httpError) Error() string {
	return fmt.Sprintf("%d: %s", e.code, e.message)
}
func subscribeEvent(t *testing.T, url string, eventCh chan *sse.Event, errCh chan error) *sse.Client {
	t.Helper()
	client := sse.NewClient(url)
	client.ResponseValidator = func(c *sse.Client, resp *http.Response) error {
		if resp.StatusCode != http.StatusOK {
			err := &httpError{message: "request failed", code: resp.StatusCode}
			return backoff.Permanent(err)
		}
		return nil
	}
	go func() {
		err := client.SubscribeChanRaw(eventCh)
		if err != nil {
			errCh <- err
		}
	}()
	return client
}

func unsubscribeEvent(t *testing.T, client *sse.Client, eventCh chan *sse.Event) {
	t.Helper()
	go client.Unsubscribe(eventCh)
}
