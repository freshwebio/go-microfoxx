package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/freshwebio/go-microfoxx/types"
)

const (
	indexEndpoint = "/index"
)

// IndexClient provides functionality to carry out tasks
// around indices in the underlying data store.
type IndexClient interface {
	GetIndexes(coll string) *types.IndexListResult
	RemoveIndex(handle string) *types.IndexOpResult
	CreateIndex(params *types.IndexParams) *types.IndexOpResult
}

// GetIndexes retrieves the indexes for the provided collection.
func (c *clientImpl) GetIndexes(coll string) *types.IndexListResult {
	req := c.prepareRequest("GET", indexEndpoint+"/"+coll, nil, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.IndexListResult{Err: err}
	}
	var ilRes types.IndexListResult
	ilRes.StatusCode = resp.StatusCode
	if resp.StatusCode == http.StatusOK {
		var indexes []*types.Index
		err = json.NewDecoder(resp.Body).Decode(&indexes)
		if err != nil {
			return &types.IndexListResult{Err: err}
		}
		ilRes.Indexes = indexes
	} else {
		msg, err := prepareExceptionResponse(resp)
		ilRes.Message = msg
		ilRes.Err = err
	}
	return &ilRes
}

// RemoveIndex deals with removing the index with the provided
// handle {collection}/{key} from the data store.
func (c *clientImpl) RemoveIndex(handle string) *types.IndexOpResult {
	req := c.prepareRequest("DELETE", indexEndpoint+"/"+handle, nil, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.IndexOpResult{Err: err}
	}
	var ioRes types.IndexOpResult
	ioRes.StatusCode = resp.StatusCode
	if resp.StatusCode == http.StatusOK {
		var intermediary struct {
			Message string `json:"message"`
		}
		err = json.NewDecoder(resp.Body).Decode(&intermediary)
		if err != nil {
			return &types.IndexOpResult{Err: err}
		}
		ioRes.Message = intermediary.Message
	} else {
		msg, err := prepareExceptionResponse(resp)
		ioRes.Message = msg
		ioRes.Err = err
	}
	return &ioRes
}

// CreateIndex deals with creating a new index in the data store to
// adhere to the provided parameters.
func (c *clientImpl) CreateIndex(params *types.IndexParams) *types.IndexOpResult {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(params)
	if err != nil {
		return &types.IndexOpResult{Err: err}
	}
	req := c.prepareRequest("POST", indexEndpoint, nil, b)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.IndexOpResult{Err: err}
	}
	var ioRes types.IndexOpResult
	ioRes.StatusCode = resp.StatusCode
	if resp.StatusCode != http.StatusCreated {
		msg, err := prepareExceptionResponse(resp)
		ioRes.Message = msg
		ioRes.Err = err
	}
	return &ioRes
}
