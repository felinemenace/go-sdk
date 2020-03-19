// Copyright (c) 2016 - 2020 Sqreen. All Rights Reserved.
// Please refer to our terms for more information:
// https://www.sqreen.io/terms.html

package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sqreen/go-standalone-sdk/signal/client"
	"github.com/stretchr/testify/require"
)

func TestSignalService(t *testing.T) {
	t.Run("SendSignal", func(t *testing.T) {
		mux := http.NewServeMux()
		tsrv := httptest.NewServer(mux)
		c := client.NewClient(tsrv.Client())
		baseURL, err := url.Parse(tsrv.URL)
		require.NoError(t, err)
		c.BaseURL = baseURL

		signal := &client.Signal{
			SignalPayload: client.SignalPayload{
				Schema:  "my schema",
				Payload: "hello signal",
			},
			Type:   "my type",
			Name:   "my signal",
			Source: "agent",
		}

		mux.HandleFunc("/", func(_ http.ResponseWriter, r *http.Request) {
			require.Equal(t, "POST", r.Method)
			require.Equal(t, "/signals", r.RequestURI)
			var sent *client.Signal
			err := json.NewDecoder(r.Body).Decode(&sent)
			require.NoError(t, err)
			require.Equal(t, signal, sent)
		})

		t.Run("nominal", func(t *testing.T) {
			err = c.SignalService().SendSignal(context.Background(), signal)
			require.NoError(t, err)
		})

		t.Run("with nil context", func(t *testing.T) {
			err = c.SignalService().SendSignal(nil, signal)
			require.Error(t, err)
		})

		t.Run("with nil value", func(t *testing.T) {
			err = c.SignalService().SendSignal(context.Background(), nil)
			require.Error(t, err)
		})
	})
	t.Run("SendTrace", func(t *testing.T) {
		mux := http.NewServeMux()
		tsrv := httptest.NewServer(mux)
		c := client.NewClient(tsrv.Client())
		baseURL, err := url.Parse(tsrv.URL)
		require.NoError(t, err)
		c.BaseURL = baseURL

		trace := &client.Trace{
			Data: []client.Signal{
				{
					SignalPayload: client.SignalPayload{
						Schema:  "my schema 1",
						Payload: "hello signal 1",
					},
					Type:   "my type 1",
					Name:   "my signal 1",
					Source: "agent",
				},
				{
					SignalPayload: client.SignalPayload{
						Schema:  "my schema 2",
						Payload: "hello signal 2",
					},
					Type:   "my type 2",
					Name:   "my signal 2",
					Source: "agent",
				},
			},
		}

		mux.HandleFunc("/", func(_ http.ResponseWriter, r *http.Request) {
			require.Equal(t, "POST", r.Method)
			require.Equal(t, "/traces", r.RequestURI)
			var sent *client.Trace
			err := json.NewDecoder(r.Body).Decode(&sent)
			require.NoError(t, err)
			require.Equal(t, trace, sent)
		})

		t.Run("nominal", func(t *testing.T) {
			err = c.SignalService().SendTrace(context.Background(), trace)
			require.NoError(t, err)
		})

		t.Run("with nil context", func(t *testing.T) {
			err = c.SignalService().SendTrace(nil, trace)
			require.Error(t, err)
		})

		t.Run("with empty trace data", func(t *testing.T) {
			err = c.SignalService().SendTrace(context.Background(), &client.Trace{
				Data: []client.Signal{},
			})
			require.Error(t, err)
		})

		t.Run("with nil value", func(t *testing.T) {
			err = c.SignalService().SendTrace(context.Background(), nil)
			require.Error(t, err)
		})
	})
	t.Run("SendBatch", func(t *testing.T) {
		mux := http.NewServeMux()
		tsrv := httptest.NewServer(mux)
		c := client.NewClient(tsrv.Client())
		baseURL, err := url.Parse(tsrv.URL)
		require.NoError(t, err)
		c.BaseURL = baseURL

		trace := &client.Trace{
			Signal: client.Signal{
				Name: "my trace",
				SignalPayload: client.SignalPayload{
					Schema:  "my trace schema",
					Payload: "hello trace",
				},
			},
			Data: []client.Signal{
				{
					Name: "my signal 1",
					SignalPayload: client.SignalPayload{
						Schema:  "my signal schema 1",
						Payload: "hello signal 1",
					},
				},
				{
					Name: "my signal 2",
					SignalPayload: client.SignalPayload{
						Schema:  "my signal schema 2",
						Payload: "hello signal 2",
					},
				},
			},
		}

		signal := &client.Signal{
			Name: "my signal 3",
			SignalPayload: client.SignalPayload{
				Schema:  "my signal schema 3",
				Payload: "hello signal 3",
			},
		}

		batch := client.Batch{
			signal,
			trace,
		}

		mux.HandleFunc("/", func(_ http.ResponseWriter, r *http.Request) {
			require.Equal(t, "POST", r.Method)
			require.Equal(t, "/batches", r.RequestURI)

			var recvBatch []json.RawMessage
			err := json.NewDecoder(r.Body).Decode(&recvBatch)
			require.NoError(t, err)
			require.Len(t, recvBatch, len(batch))

			var recvSignal *client.Signal
			err = json.Unmarshal(recvBatch[0], &recvSignal)
			require.NoError(t, err)
			require.Equal(t, signal, recvSignal)

			var recvTrace *client.Trace
			err = json.Unmarshal(recvBatch[1], &recvTrace)
			require.NoError(t, err)
			require.Equal(t, trace, recvTrace)
		})

		t.Run("nominal", func(t *testing.T) {
			err = c.SignalService().SendBatch(context.Background(), batch)
			require.NoError(t, err)
		})

		t.Run("with nil context", func(t *testing.T) {
			err = c.SignalService().SendBatch(nil, batch)
			require.Error(t, err)
		})

		t.Run("with empty batch", func(t *testing.T) {
			err = c.SignalService().SendBatch(context.Background(), client.Batch{})
			require.Error(t, err)
		})

		t.Run("with nil value", func(t *testing.T) {
			err = c.SignalService().SendBatch(context.Background(), nil)
			require.Error(t, err)
		})
	})
}
