package mpeg_dash_tools

import (
	"fmt"
	"net/http"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var _ HttpClient = http.DefaultClient

type RequestTuple struct {
	Method string
	Url    string
}

type MockedHttpClient struct {
	requests  []RequestTuple
	responses []http.Response
}

func (m *MockedHttpClient) Do(req *http.Request) (*http.Response, error) {
	if len(m.responses) == 0 {
		return nil, fmt.Errorf("no more responses in the mocked client")
	}

	if len(m.requests) == 0 {
		return nil, fmt.Errorf("no more requests in the mocked client")
	}

	expectedReq := m.requests[0]
	res := m.responses[0]

	// Consume one request and one response
	m.requests = m.requests[1:]
	m.responses = m.responses[1:]

	if req.URL.String() != expectedReq.Url || req.Method != expectedReq.Method {
		fmt.Printf("req URL: %s\n", req.URL.String())
		return &http.Response{StatusCode: 404, Status: "Not Found (mocked client didn't expect this)"}, nil
	}

	return &res, nil
}

var _ HttpClient = (*MockedHttpClient)(nil)
