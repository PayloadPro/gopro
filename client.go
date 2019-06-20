package payloadpro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"io/ioutil"
	"log"
	"sync"

	"github.com/hashicorp/go-cleanhttp"
)

// APIURL is the default payload pro api uri
const APIURL = "http://api.payload.pro"

// ClientAPI interface for mocking data in unit tests
type ClientAPI interface {
	CreateBin(bin *Bin) (*Bin, error)
	ListBins() ([]*Bin, error)
	//	ListRequests(binID string) ([]*Request, error)
	ReadBin(binID string) (*Bin, error)
	//	ReadRequest(requestID string) (*Request, error)
}

// Client provides access to create, read, update and delete runscope resources
type Client struct {
	APIURL      string
	AccessToken string
	HTTP        *http.Client
	sync.Mutex
}

type response struct {
	Meta  metaResponse  `json:"meta"`
	Data  interface{}   `json:"data"`
	Error errorResponse `json:"error"`
}

type errorResponse struct {
	Status       int    `json:"status"`
	ErrorMessage string `json:"error"`
}

type metaResponse struct {
	Status string `json:"status"`
}

// NewClient creates a new client instance
func NewClient(apiURL string, accessToken string) *Client {
	client := Client{
		APIURL:      apiURL,
		AccessToken: accessToken,
		HTTP:        cleanhttp.DefaultClient(),
	}

	return &client
}

// NewClientAPI Interface initialization
func NewClientAPI(apiURL string, accessToken string) ClientAPI {
	return &Client{
		APIURL:      apiURL,
		AccessToken: accessToken,
		HTTP:        cleanhttp.DefaultClient(),
	}
}

func (client *Client) createResource(
	resource interface{}, resourceType string, resourceName string, endpoint string) (*response, error) {
	log.Printf("[DEBUG] creating %s %s", resourceType, resourceName)

	bytes, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] 	request: POST %s %s", endpoint, string(bytes))

	req, err := client.newRequest("POST", endpoint, bytes)
	if err != nil {
		return nil, err
	}

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
			return nil, fmt.Errorf("Error creating %s: %s", resourceType, resourceName)
		}

		return nil, fmt.Errorf("Error creating %s: %s, status: %d reason: %q", resourceType,
			resourceName, errorResp.Status, errorResp.ErrorMessage)
	}

	response := new(response)
	json.Unmarshal(bodyBytes, &response)
	return response, nil

}

func (client *Client) readResource(resourceType string, resourceName string, endpoint string) (*response, error) {
	log.Printf("[DEBUG] reading %s %s", resourceType, resourceName)
	response := new(response)

	req, err := client.newRequest("GET", endpoint, nil)
	if err != nil {
		return response, err
	}

	log.Printf("[DEBUG] 	request: GET %s", endpoint)
	resp, err := client.HTTP.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	log.Printf("[DEBUG] 	response: %d %s", resp.StatusCode, bodyString)

	if resp.StatusCode >= 300 {
		errorResp := new(errorResponse)
		if err = json.Unmarshal(bodyBytes, &errorResp); err != nil {
			return response, fmt.Errorf("Status: %s Error reading %s: %s",
				resp.Status, resourceType, resourceName)
		}

		return response, fmt.Errorf("Status: %s Error reading %s: %s, reason: %q",
			resp.Status, resourceType, resourceName, errorResp.ErrorMessage)
	}

	json.Unmarshal(bodyBytes, &response)
	return response, nil
}

func (client *Client) newRequest(method string, endpoint string, body []byte) (*http.Request, error) {

	var urlStr string
	urlStr = client.APIURL + endpoint
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("Error during parsing request URL: %s", err)
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("Error during creation of request: %s", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.AccessToken))
	req.Header.Add("Accept", "application/json")

	if method != "GET" {
		req.Header.Add("Content-Type", "application/json")
	}

	return req, nil
}
