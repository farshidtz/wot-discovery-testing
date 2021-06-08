package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/linksmart/thing-directory/wot"
)

type any = interface{}
type mapAny = map[string]any

func mockedTD(id string) map[string]any {
	var td = map[string]any{
		"@context": "https://www.w3.org/2019/wot/td/v1",
		"title":    "example thing",
		"security": []string{"nosec_sc"},
		"securityDefinitions": map[string]any{
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

func retrieveThing(id string, t *testing.T) mapAny {
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
		t.Fatalf("Error decoding body: %s", err)
	}
	return retrievedTD
}

func createThing(id string, td mapAny, t *testing.T) mapAny {
	b, _ := json.Marshal(td)

	res, err := httpPut(serverURL+"/things/"+id, wot.MediaTypeThingDescription, b)
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

	storedTD := retrieveThing(id, t)

	// add the system-generated attributes
	td["registration"] = storedTD["registration"]
	return td
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

func httpRequest(method, url, contentType string, b []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
