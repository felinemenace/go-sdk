// Copyright (c) 2016 - 2020 Sqreen. All Rights Reserved.
// Please refer to our terms for more information:
// https://www.sqreen.io/terms.html

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	DefaultBaseURL = "https://ingestion.sqreen.com/"
)

type Client struct {
	BaseURL *url.URL
	Logger  DebugLogger
	client  *http.Client
	token   string
}

type DebugLogger interface {
	Debugf(format string, v ...interface{})
}

func NewClient(client *http.Client, token string) *Client {
	if client == nil {
		client = &http.Client{}
	}
	baseURL, _ := url.Parse(DefaultBaseURL)
	return &Client{
		client:  client,
		BaseURL: baseURL,
		token:   token,
	}
}

func (c *Client) SignalService() *SignalService {
	return (*SignalService)(c)
}

func (c *Client) newRequest(method, urlStr string, reqBody interface{}) (*http.Request, error) {
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if reqBody != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(reqBody); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, respBody interface{}) error {
	if ctx == nil {
		return errors.New("context must be non-nil")
	}

	req = req.WithContext(ctx)
	req.Header.Set("X-Session-Key", c.token)

	c.debugf("sending request\n%s\n", (*httpRequestStringer)(req))

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		// Drain the body and close it in order to make the underlying connection
		// available again in the pool
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	c.debugf("received response\n%s\n", (*httpResponseStringer)(resp))

	err = checkResponse(resp)
	if err != nil {
		return err
	}

	if respBody != nil {
		decErr := json.NewDecoder(resp.Body).Decode(respBody)
		if decErr != nil && decErr != io.EOF {
			return decErr
		}
	}

	return nil
}

func (c *Client) debugf(fmt string, args ...interface{}) {
	if c.Logger == nil {
		return
	}
	c.Logger.Debugf(fmt, args...)
}

type (
	httpRequestStringer  http.Request
	httpResponseStringer http.Response
)

func (r *httpRequestStringer) String() string {
	dump, _ := httputil.DumpRequestOut((*http.Request)(r), true)
	return string(dump)
}

func (r *httpResponseStringer) String() string {
	dump, _ := httputil.DumpResponse((*http.Response)(r), true)
	return string(dump)
}

// Client error types.
type (
	// APIError is the generic request error returned when the request status
	// code is unknown.
	APIError struct {
		Response *http.Response
	}
	// AuthError is a request error returned when the request could not be
	// authenticated.
	AuthTokenError APIError
	// InvalidSignalError is a request error returned when one or more signal(s)
	// sent are invalid.
	InvalidSignalError APIError
)

func (e APIError) Error() string {
	return fmt.Sprintf("api error: response with status code %s", e.Response.Status)
}

func (e AuthTokenError) Error() string {
	return "api error: access token is missing or invalid"
}

func (e InvalidSignalError) Error() string {
	return "api error: one of the provided signal is invalid"
}

func checkResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := APIError{Response: r}
	switch r.StatusCode {
	case http.StatusUnauthorized:
		return AuthTokenError(errorResponse)
	case http.StatusUnprocessableEntity:
		return InvalidSignalError(errorResponse)
	default:
		return errorResponse
	}
}
