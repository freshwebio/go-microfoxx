package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/freshwebio/go-microfoxx/types"
)

const (
	createCollEndpoint = "/collection"
)

// CollectionClient provides the basis for a client that handles
// collection functionality for the underlying data store service.
type CollectionClient interface {
	CreateColl(string) *types.CreationResult
}

// CreateColl deals with creating a new collection in the ArangoDB
// data store with the provided name.
func (c *clientImpl) CreateColl(name string) *types.CreationResult {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(struct {
		Name string `json:"name"`
	}{Name: name})
	if err != nil {
		return &types.CreationResult{Err: err}
	}
	req := c.prepareRequest("POST", createCollEndpoint, nil, b)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.CreationResult{Err: err}
	}
	// Now deal with attempting to retrieve the response information from the server.
	var creationResult types.CreationResult
	// On a 201 or 200 response then simply parse the meta data from the resposne.
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		creationResult.StatusCode = resp.StatusCode
		var intermediary = struct {
			ID      string `json:"_id"`
			Message string `json:"message"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&intermediary)
		if err != nil {
			creationResult.Err = err
		} else {
			creationResult.CreatedIDs = []string{intermediary.ID}
			creationResult.Message = intermediary.Message
		}
	} else {
		// Otherwise on any other response, parse the response error message and status code
		// to pass into our creationInfo object.
		creationResult.StatusCode = resp.StatusCode
		msg, err := prepareExceptionResponse(resp)
		creationResult.Message = msg
		creationResult.Err = err
	}
	return &creationResult
}
