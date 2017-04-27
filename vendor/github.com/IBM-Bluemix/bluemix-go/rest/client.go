package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/IBM-Bluemix/bluemix-go/bmxerror"
)

//ErrEmptyResponseBody ...
var ErrEmptyResponseBody = errors.New("empty response body")
var defaultMaxRetries = 3
var timeLagBtwReq = 10 * time.Second

// Client is a REST client. It's recommend that a client be created with the
// NewClient() method.
type Client struct {
	// The HTTP client to be used. Default is HTTP's defaultClient.
	HTTPClient *http.Client
	// Defaualt header for all outgoing HTTP requests.
	DefaultHeader http.Header
	//Maximum number of retries
	MaxRetries int

	//Retry delay
	RetryDelay time.Duration

	Debug bool
}

// NewClient creates a new REST client.
func NewClient() *Client {
	return &Client{
		HTTPClient: http.DefaultClient,
		MaxRetries: defaultMaxRetries,
		RetryDelay: timeLagBtwReq,
	}
}

// Do sends an request and returns an HTTP response. The resp.Body will be
// consumed and closed in the method.
//
// For 2XX response, it will be JSON decoded into the value pointed to by
// respv.
//
// For non-2XX response, an attempt will be made to unmarshal the response
// into the value pointed to by errV. If unmarshal failed, an error with status code
// and response text is returned.
func (c *Client) Do(r *Request, respV interface{}, errV interface{}) (*http.Response, error) {
	for i := 0; ; i++ {
		req, err := c.makeRequest(r)
		if err != nil {
			return nil, err
		}

		client := c.HTTPClient
		if client == nil {
			client = http.DefaultClient
		}

		var resp *http.Response
		if c.Debug && client.Transport != nil {
			resp, err = client.Transport.RoundTrip(req)
		} else {
			resp, err = client.Do(req)
		}

		if err != nil {
			return resp, err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			remain := c.MaxRetries - i
			if remain == 0 || (resp.StatusCode > 299 && resp.StatusCode < 500) {

				raw, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return resp, fmt.Errorf("Error reading response: %v", err)
				}

				if len(raw) > 0 && errV != nil {
					if json.Unmarshal(raw, errV) == nil {
						return resp, nil
					}
				}

				return resp, bmxerror.NewRequestFailure("", string(raw), resp.StatusCode)
			}
			time.Sleep(c.RetryDelay)
		} else {
			if respV != nil {
				switch respV.(type) {
				case io.Writer:
					_, err = io.Copy(respV.(io.Writer), resp.Body)
				default:
					err = json.NewDecoder(resp.Body).Decode(respV)
					if err == io.EOF {
						err = ErrEmptyResponseBody
					}
				}
			}

			return resp, err
		}
	}

}

func (c *Client) makeRequest(r *Request) (*http.Request, error) {
	req, err := r.Build()
	if err != nil {
		return nil, err
	}

	c.applyDefaultHeader(req)

	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) applyDefaultHeader(req *http.Request) {
	for k, vs := range c.DefaultHeader {
		if req.Header.Get(k) != "" {
			continue
		}
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
}
