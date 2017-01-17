package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/freshwebio/go-microfoxx/types"
)

const (
	graphEndpoint    = "/graph"
	relationEndpoint = "/relation"
)

// GraphClient provides graph-specific data store
// service functionality.
type GraphClient interface {
	CreateGraph(*types.Graph) *types.CreationResult
}

// CreateGraph deals with creating a new graph and all the relations defined
// in the provided graph definition.
func (c *clientImpl) CreateGraph(graphDef *types.Graph) *types.CreationResult {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(graphDef)
	if err != nil {
		return &types.CreationResult{Err: err}
	}
	req := c.prepareRequest("POST", graphEndpoint, nil, b)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.CreationResult{Err: err}
	}
	// Now deal with retrieving the response information returned from the service.
	var creationResult types.CreationResult
	// On 2xx response code retrieve the response information from the server.
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		creationResult.StatusCode = resp.StatusCode
		var intermediary = struct {
			Message string `json:"message"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&intermediary)
		if err != nil {
			creationResult.Err = err
		} else {
			creationResult.Message = intermediary.Message
		}
	} else {
		// In the case of a different response status code to pass to our
		// CreationResult object.
		creationResult.StatusCode = resp.StatusCode
		msg, err := prepareExceptionResponse(resp)
		creationResult.Message = msg
		creationResult.Err = err
	}
	return &creationResult
}

func (c *clientImpl) CreateRelation(graph string, relation *types.Relation) *types.CreationResult {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(relation)
	if err != nil {
		return &types.CreationResult{Err: err}
	}
	// Now make the request to the foxx service.
	req := c.prepareRequest("POST", graphEndpoint+"/"+graph+relationEndpoint, nil, b)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.CreationResult{Err: err}
	}
	var creationResult types.CreationResult
	creationResult.StatusCode = resp.StatusCode
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		// Now try to parse the response of the succesful attempt to add a new relation
		// to the specified graph.
		var intermediary = struct {
			Message string `json:"message"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&intermediary)
		if err != nil {
			creationResult.Err = err
		} else {
			creationResult.Message = intermediary.Message
		}
	} else {
		msg, err := prepareExceptionResponse(resp)
		creationResult.Message = msg
		creationResult.Err = err
	}
	return &creationResult
}
