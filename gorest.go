package gorest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"io"
)

// This is a basic REST client.  I would have preferred to simply import
//    github.com/go-resty/resty.v0 but the SevOne REST API does not accept
//    auth tokens in the form of an "Authorization: Bearer" header; it expects
//    them in the X-Auth-Token header.
// Hopefully, this will eventually be fixed.

type Client struct {
	// HTTP client
	client *http.Client

	// Base URL for API requests.
	BaseURL *url.URL

	// Additional headers
	Headers map[string]string
}

// This type will allow us to create type functions for this library
type Response http.Response

// This will decode the return JSON into whatever you provide as a container
func (this *Response) Decode(target interface{}) (error) {
	err := json.NewDecoder(this.Body).Decode(target)
	err2 := this.Body.Close()
	if(err == nil) {
		return err2
	}
	return err
}

func New(api_url string, extra_headers map[string]string) *Client {
	// Ensure the URL ends with a slash
	if(api_url[len(api_url) - 1] != '/') {
		api_url += "/"
	}

	base_url, _ := url.Parse(api_url)
	client := &Client{
		client : http.DefaultClient,
		BaseURL : base_url,
		Headers : make(map[string]string),
	}

	for header, value := range extra_headers {
		client.Headers[header] = value
	}

	return client
}

// Creates an API request.  A relative URL can be provided ("path"), which will
//    be resolved to the BaseURL of the Client.  Relative URLs should always be
//    specified WITHOUT a preceding slash.
func (this *Client) Request(method string, path string, body io.Reader) (*http.Response, error) {
	// Ensure the relative URL doesn't start with a slash
	if(path[0] == '/') {
		path = path[1:]
	}

	// Parse the URL
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	// Build the URL
	full_url := this.BaseURL.ResolveReference(rel)

	// Make the request
	req, err := http.NewRequest(method, full_url.String(), body)
	if err != nil {
		return nil, err
	}

	// Headers
	for header, value := range this.Headers {
		req.Header.Add(header, value)
	}

	// Do the request
	return this.client.Do(req)
}

// This reads from your source container and provides a Reader for the request
func NewJSONReader(source interface{}) (io.Reader, error) {
	JSONBytes, err := json.Marshal(source)
	if(err != nil) {
		return nil, nil
	}
	JSONReader := bytes.NewReader(JSONBytes)
	return JSONReader, nil
}

// Begin request functions

func (this *Client) Get(path string) (*Response, error) {
	httpresp, err := this.Request("GET", path, nil)
	if(err != nil) {
		return nil, err
	}
	resp := Response(*httpresp)
	if(err != nil) {
		return nil, err
	}
	return &resp, nil
}

func (this *Client) Delete(path string) (*Response, error) {
	httpresp, err := this.Request("DELETE", path, nil)
	if(err != nil) {
		return nil, err
	}
	resp := Response(*httpresp)
	if(err != nil) {
		return nil, err
	}
	return &resp, nil
}

func (this *Client) Post(path string, data interface{}) (*Response, error) {
	var JSONReader io.Reader
	var err error

	// If "data" is a reader, we'll assume it already contains valid JSON;
	//    otherwise, we'll hand off to a function to return JSON from anything
	//    Marshal-able
	switch data := data.(type) {
	case io.Reader:
		JSONReader = data
	default:
		JSONReader, err = NewJSONReader(data)
		if(err != nil) {
			return nil, err
		}
	}
	httpresp, err := this.Request("POST", path, JSONReader)
	if(err != nil) {
		return nil, err
	}
	resp := Response(*httpresp)
	return &resp, nil
}

func (this *Client) Put(path string, data interface{}) (*Response, error) {
	// If it's a reader, we'll assume we're already passing good JSON
	// Otherwise we'll hand off to a function to return JSON from about anything
	var JSONReader io.Reader
	var err error
	switch data := data.(type) {
	case io.Reader:
		JSONReader = data
	default:
		JSONReader, err = NewJSONReader(data)
		if(err != nil) {
			return nil, err
		}
	}
	httpresp, err := this.Request("PUT", path, JSONReader)
	if(err != nil) {
		return nil, err
	}
	resp := Response(*httpresp)
	return &resp, nil
}

