package payloadpro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/google/jsonapi"
)

// Bin resources are a simple way to organize your requests
type Bin struct {
	ID          string `jsonapi:"primary,bin"`
	Name        string `jsonapi:"attr,name"`
	Description string `jsonapi:"attr,description"`
}

// CreateBin creates a new bin resource
func (client *Client) CreateBin(bin *Bin) (*Bin, error) {
	log.Printf("[DEBUG] creating bin %s", bin.Name)

	in := bytes.NewBuffer(nil)
	jsonapi.MarshalPayload(in, bin)
	log.Printf("[DEBUG] 	request: POST %s %#v", "/bins", bin)

	req, err := client.newRequest("POST", "/bins", in.Bytes())
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] %#v", req)
	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	log.Printf("[DEBUG] 	response: %d %s", resp.StatusCode, bodyString)

	if resp.StatusCode >= 300 {
		errorResp := new(errorResponse)
		if err = json.Unmarshal(bodyBytes, &errorResp); err != nil {
			return nil, fmt.Errorf("Error creating bin: %s", bin.Name)
		}

		return nil, fmt.Errorf("Error creating bin: %s, status: %d reason: %q", bin.Name,
			errorResp.Status, errorResp.ErrorMessage)

	}

	response := new(response)
	json.Unmarshal(bodyBytes, &response)
	return getBinFromResponse(response.Data)
}

// ReadBin list details about an existing bin resource
func (client *Client) ReadBin(binID string) (*Bin, error) {
	resource, error := client.readResource("bin", binID, fmt.Sprintf("/bins/%s", binID))
	if error != nil {
		return nil, error
	}

	bin, error := getBinFromResponse(resource.Data)
	return bin, error
}

// ListBins lists all bins
func (client *Client) ListBins() ([]*Bin, error) {
	resource, error := client.readResource("[]bin", "", "/bins")
	if error != nil {
		return nil, error
	}

	bins, error := getBinsFromResponse(resource.Data)
	return bins, error
}

func (bin *Bin) String() string {
	value, err := json.Marshal(bin)
	if err != nil {
		return ""
	}

	return string(value)
}

func getBinsFromResponse(response interface{}) ([]*Bin, error) {
	var bins []*Bin
	err := decode(&bins, response)
	return bins, err
}

func getBinFromResponse(response interface{}) (*Bin, error) {
	bin := new(Bin)
	err := decode(bin, response)
	return bin, err
}
