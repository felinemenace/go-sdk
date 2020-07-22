// Copyright (c) 2016 - 2020 Sqreen. All Rights Reserved.
// Please refer to our terms for more information:
// https://www.sqreen.io/terms.html

package client

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResponseErrorChecking(t *testing.T) {
	t.Run("unexpected response status code", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusBadGateway,
		}
		err := checkResponse(resp)
		require.Equal(t, APIError{Response: resp}, err)
		require.NotEmpty(t, err.Error())
	})

	t.Run("token authentication error", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusUnauthorized,
		}
		err := checkResponse(resp)
		require.Equal(t, AuthTokenError{Response: resp}, err)
		require.NotEmpty(t, err.Error())
	})

	t.Run("invalid signal error", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusUnprocessableEntity,
		}
		err := checkResponse(resp)
		require.Equal(t, InvalidSignalError{Response: resp}, err)
		require.NotEmpty(t, err.Error())
	})

	t.Run("ok status code range", func(t *testing.T) {
		for _, status := range []int{http.StatusOK, http.StatusAccepted, 233, 250, 299} {
			status := status
			t.Run(fmt.Sprintf("%d", status), func(t *testing.T) {
				resp := &http.Response{
					StatusCode: status,
				}
				err := checkResponse(resp)
				require.NoError(t, err)
			})
		}
	})
}

func TestClient(t *testing.T) {
	t.Run("NewClient", func(t *testing.T) {
		t.Run("default http client", func(t *testing.T) {
			require.NotNil(t, NewClient(nil, ""))
		})

		t.Run("specific http client", func(t *testing.T) {
			client := &http.Client{
				Timeout: 10 * time.Second,
			}
			require.NotNil(t, NewClient(client, ""))
		})
	})

	t.Run("newRequest", func(t *testing.T) {
		c := NewClient(nil, "")

		for _, tc := range []struct {
			name         string
			endpoint     string
			method       string
			body         interface{}
			wantError    bool
			expectedBody string
		}{
			{
				name:      "get /endpoint without body",
				endpoint:  "endpoint",
				method:    http.MethodGet,
				body:      nil,
				wantError: false,
			},
			{
				name:      "post /endpoint without body",
				endpoint:  "endpoint",
				method:    http.MethodPost,
				body:      nil,
				wantError: false,
			},
			{
				name:      "bad method",
				endpoint:  "endpoint",
				method:    ";",
				body:      nil,
				wantError: true,
			},
			{
				name:      "bad endpoint",
				endpoint:  ":endpoint",
				method:    "GET",
				body:      nil,
				wantError: true,
			},
			{
				name:      "bad endpoint",
				endpoint:  ":endpoint",
				method:    "GET",
				body:      nil,
				wantError: true,
			},
			{
				name:         "post /version/endpoint with body",
				endpoint:     "version/endpoint",
				method:       http.MethodPost,
				body:         []string{"a", "b", "c"},
				expectedBody: "[\"a\",\"b\",\"c\"]\n",
			},
			{
				name:         "post /version/endpoint with body",
				endpoint:     "version/endpoint",
				method:       http.MethodPost,
				body:         "no html & éscaping <",
				expectedBody: "\"no html & éscaping <\"\n",
			},
			{
				name:      "post /endpoint with body marshaling error",
				endpoint:  "version/endpoint",
				method:    http.MethodPost,
				body:      jsonMarshalError{},
				wantError: true,
			},
		} {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				req, err := c.newRequest(tc.method, tc.endpoint, tc.body)

				if tc.wantError {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				require.Equal(t, tc.method, req.Method)
				require.Equal(t, DefaultBaseURL+tc.endpoint, req.URL.String())
				if tc.expectedBody != "" {
					body, err := ioutil.ReadAll(req.Body)
					require.NoError(t, err)
					require.Equal(t, tc.expectedBody, string(body))
				}
			})
		}
	})

	t.Run("do", func(t *testing.T) {
		for _, tc := range []struct {
			name             string
			reqBody          interface{}
			expectedReqBody  string
			respBody         string
			expectedRespBody interface{}
			wantError        bool
			status           int
		}{
			{
				name: "no request nor response bodies",
			},
			{
				name:            "request body without response body",
				reqBody:         "string",
				expectedReqBody: "\"string\"\n",
			},
			{
				name:             "request and response body",
				reqBody:          "request",
				expectedReqBody:  "\"request\"\n",
				respBody:         "\"response\"\n",
				expectedRespBody: "response",
			},
			{
				name:             "no request body and response body",
				respBody:         "\"response\"\n",
				expectedRespBody: "response",
			},
			{
				name:      "bad response json",
				respBody:  "\"oops",
				wantError: true,
			},
			{
				name:      "error status code",
				status:    http.StatusUnprocessableEntity,
				wantError: true,
			},
			{
				name:   "ok status code",
				status: 200,
			},
		} {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if tc.status != 0 {
						w.WriteHeader(tc.status)
					}
					if tc.expectedReqBody != "" {
						require.Equal(t, "application/json", r.Header.Get("Content-Type"))
						body, err := ioutil.ReadAll(r.Body)
						require.NoError(t, err)
						require.Equal(t, tc.expectedReqBody, string(body))
					}
					_, _ = w.Write([]byte(tc.respBody))
				}))
				defer srv.Close()

				baseURL, err := url.Parse(srv.URL)
				require.NoError(t, err)

				c := NewClient(srv.Client(), "")
				c.BaseURL = baseURL

				req, err := c.newRequest("GET", "endpoint", tc.reqBody)
				require.NoError(t, err)

				var respBody interface{}
				err = c.do(context.Background(), req, &respBody)
				if tc.wantError {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.Equal(t, tc.expectedRespBody, respBody)
			})
		}
	})

	t.Run("do with context", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			time.Sleep(1 * time.Second)
		}))
		defer srv.Close()

		baseURL, err := url.Parse(srv.URL)
		require.NoError(t, err)

		c := NewClient(srv.Client(), "")
		c.BaseURL = baseURL

		req, err := c.newRequest("PUT", "endpoint", nil)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond)
		defer cancel()
		err = c.do(ctx, req, nil)
		require.Error(t, err)
		require.True(t, errors.Is(err, context.DeadlineExceeded))
	})

	t.Run("do with context", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		defer srv.Close()

		baseURL, err := url.Parse(srv.URL)
		require.NoError(t, err)

		c := NewClient(srv.Client(), "")
		c.BaseURL = baseURL

		req, err := c.newRequest("PUT", "endpoint", nil)
		require.NoError(t, err)

		err = c.do(nil, req, nil)
		require.Error(t, err)
	})
}

type jsonMarshalError struct{}

func (jsonMarshalError) UnmarshalJSON([]byte) error   { return errors.New("oops") }
func (jsonMarshalError) MarshalJSON() ([]byte, error) { return nil, errors.New("oops") }
