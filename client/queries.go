package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/freshwebio/go-microfoxx/types"
)

// QueryClient provides the functionality for running
// data modifying queries on the data store service.
type QueryClient interface {
	InsertQuery(params *types.ModifyingQueryParams) *types.DocumentsOpResult
	UpdateQuery(params *types.ModifyingQueryParams) *types.DocumentsOpResult
	RemoveQuery(params *types.ModifyingQueryParams) *types.DocumentsOpResult
}

// InsertQuery sends the provided AQL query and relevant transaction and AQL query bind variables
// settings and the foxx service then executes the query and returns the newly inserted documents
// and all the events for each insert operation.
func (c *clientImpl) InsertQuery(params *types.ModifyingQueryParams) *types.DocumentsOpResult {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(params)
	if err != nil {
		return &types.DocumentsOpResult{
			Err: err,
		}
	}
	req := c.prepareRequest("POST", "/insert", nil, b)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.DocumentsOpResult{
			Err: err,
		}
	}
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		docOpRes := &types.DocumentsOpResult{}
		docOpRes.StatusCode = resp.StatusCode
		var intermediary = struct {
			Events    []map[string]interface{} `json:"events"`
			Documents []map[string]interface{} `json:"docs"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&intermediary)
		if err != nil {
			return &types.DocumentsOpResult{
				Err: err,
			}
		}
		bd := new(bytes.Buffer)
		err = json.NewEncoder(bd).Encode(intermediary.Documents)
		if err != nil {
			return &types.DocumentsOpResult{
				Err: err,
			}
		}
		docOpRes.Documents = bd
		be := new(bytes.Buffer)
		err = json.NewEncoder(be).Encode(intermediary.Events)
		if err != nil {
			return &types.DocumentsOpResult{
				Err: err,
			}
		}
		docOpRes.Events = be
		return docOpRes
	}
	msg, err := prepareExceptionResponse(resp)
	return &types.DocumentsOpResult{
		Err:        err,
		Message:    msg,
		StatusCode: resp.StatusCode,
	}
}

// UpdateQuery deals with sending an AQL query to the foxx service which updates existing
// documents in the Arango data store.
func (c *clientImpl) UpdateQuery(params *types.ModifyingQueryParams) *types.DocumentsOpResult {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(params)
	if err != nil {
		return &types.DocumentsOpResult{
			Err: err,
		}
	}
	req := c.prepareRequest("POST", "/update", nil, b)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.DocumentsOpResult{
			Err: err,
		}
	}
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		docOpRes := &types.DocumentsOpResult{}
		docOpRes.StatusCode = resp.StatusCode
		var intermediary = struct {
			Events    []map[string]interface{} `json:"events"`
			Documents []map[string]interface{} `json:"docs"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&intermediary)
		if err != nil {
			return &types.DocumentsOpResult{
				Err: err,
			}
		}
		bd := new(bytes.Buffer)
		err = json.NewEncoder(bd).Encode(intermediary.Documents)
		if err != nil {
			return &types.DocumentsOpResult{
				Err: err,
			}
		}
		docOpRes.Documents = bd
		be := new(bytes.Buffer)
		err = json.NewEncoder(be).Encode(intermediary.Events)
		if err != nil {
			return &types.DocumentsOpResult{
				Err: err,
			}
		}
		docOpRes.Events = be
		return docOpRes
	}
	msg, err := prepareExceptionResponse(resp)
	return &types.DocumentsOpResult{
		Err:        err,
		Message:    msg,
		StatusCode: resp.StatusCode,
	}
}

// RemoveQuery deals with executing the provided removal AQL query through the
// foxx service endpoint and returns a result with all the removed documents and each removal
// operation event.
func (c *clientImpl) RemoveQuery(params *types.ModifyingQueryParams) *types.DocumentsOpResult {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(params)
	if err != nil {
		return &types.DocumentsOpResult{
			Err: err,
		}
	}
	req := c.prepareRequest("POST", "/remove", nil, b)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.DocumentsOpResult{
			Err: err,
		}
	}
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		docOpRes := &types.DocumentsOpResult{}
		docOpRes.StatusCode = resp.StatusCode
		var intermediary = struct {
			Events    []map[string]interface{} `json:"events"`
			Documents []map[string]interface{} `json:"docs"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&intermediary)
		if err != nil {
			return &types.DocumentsOpResult{
				Err: err,
			}
		}
		bd := new(bytes.Buffer)
		err = json.NewEncoder(bd).Encode(intermediary.Documents)
		if err != nil {
			return &types.DocumentsOpResult{
				Err: err,
			}
		}
		docOpRes.Documents = bd
		be := new(bytes.Buffer)
		err = json.NewEncoder(be).Encode(intermediary.Events)
		if err != nil {
			return &types.DocumentsOpResult{
				Err: err,
			}
		}
		docOpRes.Events = be
		return docOpRes
	}
	msg, err := prepareExceptionResponse(resp)
	return &types.DocumentsOpResult{
		Err:        err,
		Message:    msg,
		StatusCode: resp.StatusCode,
	}
}
