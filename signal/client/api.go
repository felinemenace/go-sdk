// Copyright (c) 2016 - 2020 Sqreen. All Rights Reserved.
// Please refer to our terms for more information:
// https://www.sqreen.io/terms.html

package client

type SignalType string

const (
	PointSignalType  SignalType = "point"
	MetricSignalType SignalType = "metric"
)

type Signal struct {
	SignalPayload
	Type     SignalType  `json:"type"`
	Name     string      `json:"signal_name"`
	Source   string      `json:"source,omitempty"`
	Actor    interface{} `json:"actor,omitempty"`
	Context  interface{} `json:"context,omitempty"`
	Trigger  interface{} `json:"trigger,omitempty"`
	Location *Location   `json:"location,omitempty"`
}

type SignalPayload struct {
	Schema  string      `json:"payload_schema"`
	Payload interface{} `json:"payload"`
}

type Location struct {
	StackTrace []StackFrame `json:"stack_trace,omitempty"`
}

type StackFrame struct {
	// TODO
}

// Trace is a set of signals. Common signal fields can be factored in the trace
// root fields.
type Trace struct {
	Signal
	Data []Signal `json:"data"`
}

type Batch []SignalFace

// SignalFace is a simple helper to make sure only a given set of types defined
// in this package can be added to the batch array (private interface method
// can indeed only be implemented in the same package).
type SignalFace interface {
	isSignal()
}

func (*Trace) isSignal()  {}
func (*Signal) isSignal() {}
