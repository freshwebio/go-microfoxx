package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/freshwebio/go-microfoxx/types"
)

const (
	cursorEndpoint = "/cursor"
)

// CursorClient provides the base for a service
// that deals with the cursor functionality to interact with
// an underlying data store service.
type CursorClient interface {
	CursorQuery(*types.CursorQueryParams) *types.CursorQueryResult
	CursorGetNextBatch(string) *types.CursorQueryResult
}

// CursorQuery sends an AQL query to the ArangoDB service
// to create a new cursor and return the set of results.
// You can specify count if you want to retrieve the total amount
// of results and can supply a batch size to retrieve results in batches.
func (c *clientImpl) CursorQuery(params *types.CursorQueryParams) *types.CursorQueryResult {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(params)
	if err != nil {
		return &types.CursorQueryResult{Err: err}
	}
	req := c.prepareRequest("POST", cursorEndpoint, nil, b)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.CursorQueryResult{Err: err}
	}
	// Now attempt to parse the response provided by the foxx service.
	var cursorQueryRes types.CursorQueryResult
	cursorQueryRes.StatusCode = resp.StatusCode
	if cursorQueryRes.StatusCode == http.StatusOK || cursorQueryRes.StatusCode == http.StatusCreated {
		intermediary := struct {
			Results []map[string]interface{} `json:"results"`
			Cursor  string                   `json:"cursor"`
			HasMore bool                     `json:"hasMore"`
			Count   int                      `json:"count"`
		}{}
		err := json.NewDecoder(resp.Body).Decode(&intermediary)
		if err != nil {
			return &types.CursorQueryResult{Err: err}
		}
		cursorQueryRes.Cursor = intermediary.Cursor
		cursorQueryRes.HasMore = intermediary.HasMore
		docBytes := new(bytes.Buffer)
		err = json.NewEncoder(docBytes).Encode(intermediary.Results)
		if err != nil {
			return &types.CursorQueryResult{Err: err}
		}
		cursorQueryRes.Documents = docBytes
	} else {
		msg, err := prepareExceptionResponse(resp)
		cursorQueryRes.Message = msg
		cursorQueryRes.Err = err
	}
	return &cursorQueryRes
}

// CursorGetNextBatch deals with retrieving the next batch
// for the provided cursor.
func (c *clientImpl) CursorGetNextBatch(cursorID string) *types.CursorQueryResult {
	req := c.prepareRequest("PUT", cursorEndpoint+"/"+cursorID, nil, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.CursorQueryResult{Err: err}
	}
	// Now attempt to parse our next batch of results.
	var cursorQueryRes types.CursorQueryResult
	cursorQueryRes.StatusCode = resp.StatusCode
	if cursorQueryRes.StatusCode == http.StatusOK || cursorQueryRes.StatusCode == http.StatusCreated {
		intermediary := struct {
			Results []map[string]interface{} `json:"results"`
			HasMore bool                     `json:"hasMore"`
		}{}
		err := json.NewDecoder(resp.Body).Decode(&intermediary)
		if err != nil {
			return &types.CursorQueryResult{Err: err}
		}
		cursorQueryRes.HasMore = intermediary.HasMore
		docBytes := new(bytes.Buffer)
		err = json.NewEncoder(docBytes).Encode(intermediary.Results)
		if err != nil {
			return &types.CursorQueryResult{Err: err}
		}
		cursorQueryRes.Documents = docBytes
	} else {
		msg, err := prepareExceptionResponse(resp)
		cursorQueryRes.Message = msg
		cursorQueryRes.Err = err
	}
	return &cursorQueryRes
}
