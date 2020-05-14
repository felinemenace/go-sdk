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

	"github.com/sqreen/go-sdk/signal/client"
	"github.com/sqreen/go-sdk/signal/client/api"
	"github.com/stretchr/testify/require"
)

func TestSignalService(t *testing.T) {
	t.Run("SendSignal", func(t *testing.T) {
		mux := http.NewServeMux()
		tsrv := httptest.NewServer(mux)
		c := client.NewClient(tsrv.Client(), "")
		baseURL, err := url.Parse(tsrv.URL)
		require.NoError(t, err)
		c.BaseURL = baseURL

		signal := &api.Signal{
			SignalPayload: &api.SignalPayload{
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
			var sent *api.Signal
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
		c := client.NewClient(tsrv.Client(), "")
		baseURL, err := url.Parse(tsrv.URL)
		require.NoError(t, err)
		c.BaseURL = baseURL

		trace := &api.Trace{
			Data: []*api.Signal{
				{
					SignalPayload: &api.SignalPayload{
						Schema:  "my schema 1",
						Payload: "hello signal 1",
					},
					Type:   "my type 1",
					Name:   "my signal 1",
					Source: "agent",
				},
				{
					SignalPayload: &api.SignalPayload{
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
			var sent *api.Trace
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
			err = c.SignalService().SendTrace(context.Background(), &api.Trace{
				Data: []*api.Signal{},
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
		c := client.NewClient(tsrv.Client(), "")
		baseURL, err := url.Parse(tsrv.URL)
		require.NoError(t, err)
		c.BaseURL = baseURL

		trace := &api.Trace{
			Signal: api.Signal{
				Name: "my trace",
				SignalPayload: &api.SignalPayload{
					Schema:  "my trace schema",
					Payload: "hello trace",
				},
			},
			Data: []*api.Signal{
				{
					Name: "my signal 1",
					SignalPayload: &api.SignalPayload{
						Schema:  "my signal schema 1",
						Payload: "hello signal 1",
					},
				},
				{
					Name: "my signal 2",
					SignalPayload: &api.SignalPayload{
						Schema:  "my signal schema 2",
						Payload: "hello signal 2",
					},
				},
			},
		}

		signal := &api.Signal{
			Name: "my signal 3",
			SignalPayload: &api.SignalPayload{
				Schema:  "my signal schema 3",
				Payload: "hello signal 3",
			},
		}

		batch := api.Batch{
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

			var recvSignal *api.Signal
			err = json.Unmarshal(recvBatch[0], &recvSignal)
			require.NoError(t, err)
			require.Equal(t, signal, recvSignal)

			var recvTrace *api.Trace
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
			err = c.SignalService().SendBatch(context.Background(), api.Batch{})
			require.Error(t, err)
		})

		t.Run("with nil value", func(t *testing.T) {
			err = c.SignalService().SendBatch(context.Background(), nil)
			require.Error(t, err)
		})
	})
}
