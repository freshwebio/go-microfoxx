package client_test

import (
	"io"
	"net/http"
	"net/http/httptest"
)

type dummySessionClient struct{}

// Deals with returning dummy session data for the purpose of testing.
func (c *dummySessionClient) Post(url string, bodyType string, body io.Reader) (resp *http.Response, err error) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"sid\":\"12345\", \"uid\":\"6789\"}"))
	}))
	defer server.Close()
	resp, err = http.Post(server.URL, bodyType, body)
	return resp, err
}
