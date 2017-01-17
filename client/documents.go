package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/freshwebio/go-microfoxx/types"
)

// DocClient provides client functionality around handling
// documents.
type DocClient interface {
	CreateDoc(string, interface{}) *types.DocumentOpResult
	GetDocs(string, *types.DocumentRetrievalParams) *types.DocumentsResult
	GetDocCount(string, *types.DocumentRetrievalParams) *types.DocumentCountResult
	RemoveDoc(string, string) *types.DocumentOpResult
	GetDoc(string, string) *types.DocumentResult
	UpdateDoc(string, string, interface{}) *types.DocumentOpResult
}

// CreateDoc deals with creating a new document in the provided collection.
func (c *clientImpl) CreateDoc(coll string, doc interface{}) *types.DocumentOpResult {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(doc)
	if err != nil {
		return &types.DocumentOpResult{Err: err}
	}
	req := c.prepareRequest("POST", "/"+coll, nil, b)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.DocumentOpResult{Err: err}
	}
	// Now deal with retrieving the response information returned from the service.
	var docOpInfo types.DocumentOpResult
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		docOpInfo.StatusCode = resp.StatusCode
		var intermediary = struct {
			Document map[string]interface{} `json:"doc"`
			Event    map[string]interface{} `json:"event"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&intermediary)
		if err != nil {
			docOpInfo.Err = err
		} else {
			// Now encode the decoded abstract maps to a buffer of bytes representing
			// the JSON representation to allow the user to decode the data to an application specific
			// struct type for both documents and events.
			bd := new(bytes.Buffer)
			be := new(bytes.Buffer)
			err = json.NewEncoder(bd).Encode(intermediary.Document)
			if err != nil {
				return &types.DocumentOpResult{Err: err}
			}
			err = json.NewEncoder(be).Encode(intermediary.Event)
			if err != nil {
				return &types.DocumentOpResult{Err: err}
			}
			docOpInfo.Document = bd
			docOpInfo.Event = be
		}
	} else {
		docOpInfo.StatusCode = resp.StatusCode
		msg, err := prepareExceptionResponse(resp)
		docOpInfo.Message = msg
		docOpInfo.Err = err
	}
	return &docOpInfo
}

// GetDocs deals with preparing and executing a request to a microfoxx service
// implementation
func (c *clientImpl) GetDocs(coll string, params *types.DocumentRetrievalParams) *types.DocumentsResult {
	// First build the query parameters from the document retrieval parameters.
	// Build the field parameters.
	qParams := make(url.Values)
	for field, val := range params.Fields {
		// We simply assume the param name is a safe string value
		// as shouldn't ever be anything other alphanumeric characters,
		// hyphens and underscores.
		qParams.Add(field, url.QueryEscape(val))
	}
	// Now the sort fields and orders.
	sortFieldCount := len(params.SortFields)
	if sortFieldCount > 0 {
		sortValues := ""
		for i, sortField := range params.SortFields {
			sortValues += sortField
			if i < sortFieldCount-1 {
				sortValues += ","
			}
		}
		// Now append the sort order if it is set.
		if params.SortOrder != "" {
			sortValues += "::" + params.SortOrder
		}
		qParams.Add("sort", sortValues)
	}
	// Finaly if a limit is provided then add that to the query string.
	if params.LimitCount > 0 {
		qParams.Add("limit", strconv.Itoa(params.LimitOffset)+","+strconv.Itoa(params.LimitCount))
	}
	var docRes types.DocumentsResult
	req := c.prepareRequest("GET", "/"+coll, qParams, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		docRes.Err = err
		return &docRes
	}
	docRes.StatusCode = resp.StatusCode
	if resp.StatusCode == http.StatusOK {
		var intermediary []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&intermediary)
		if err != nil {
			docRes.Err = err
		} else {
			bd := new(bytes.Buffer)
			err = json.NewEncoder(bd).Encode(intermediary)
			if err != nil {
				return &types.DocumentsResult{Err: err}
			}
			docRes.Documents = bd
		}
	} else {
		msg, err := prepareExceptionResponse(resp)
		docRes.Message = msg
		docRes.Err = err
	}
	return &docRes
}

// GetDocCount deals with retrieving the amount of documents in a provided collection
// or the amount of documents filtered by properties provided as query string parameters.
func (c *clientImpl) GetDocCount(coll string, params *types.DocumentRetrievalParams) *types.DocumentCountResult {
	// Now build a query string from the fields to be applied as AQL filters.
	// Only handle fields no need for sort or limit params when retreiving counts.
	qParams := make(url.Values)
	fieldCount := len(params.Fields)
	if fieldCount > 0 {
		for k, v := range params.Fields {
			qParams.Add(k, url.QueryEscape(v))
		}
	}
	req := c.prepareRequest("GET", "/"+coll+"/count", qParams, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Return -1 when an error occurs to indicate that
		// something when wrong or that the provided collection doesn't exist.
		return &types.DocumentCountResult{
			Count: -1,
			Err:   err,
		}
	}
	var respItems map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&respItems)
	if err != nil {
		return &types.DocumentCountResult{
			Count: -1,
			Err:   err,
		}
	}
	if resp.StatusCode == http.StatusOK {
		count, ok := respItems["count"].(float64)
		if ok {
			return &types.DocumentCountResult{
				Count: int(count),
			}
		}
		return &types.DocumentCountResult{
			Count: -1,
			Err:   ErrGeneral,
		}
	}
	message, err := prepareExceptionResponseFromMap(resp.StatusCode, respItems)
	return &types.DocumentCountResult{
		Count:      -1,
		Err:        err,
		Message:    message,
		StatusCode: resp.StatusCode,
	}
}

// RemoveDoc deals with removing the document with the provided key
// from the specified collection in the underlying data store of the microfoxx app instance.
func (c *clientImpl) RemoveDoc(coll string, key string) *types.DocumentOpResult {
	req := c.prepareRequest("DELETE", "/"+coll+"/"+key, nil, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.DocumentOpResult{
			Err: err,
		}
	}
	// Now try to decode the response body.
	var respData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return &types.DocumentOpResult{
			Err: err,
		}
	}
	if resp.StatusCode == http.StatusOK {
		// Now place our event and document into two seperate io.Readers
		// to be decoded to application-specific models.
		intermediary := respData["doc"].(map[string]interface{})
		docRes := types.DocumentOpResult{}
		bd := new(bytes.Buffer)
		err = json.NewEncoder(bd).Encode(intermediary)
		if err != nil {
			return &types.DocumentOpResult{Err: err}
		}
		docRes.Document = bd
		evtIntermediary := respData["event"].(map[string]interface{})
		be := new(bytes.Buffer)
		err = json.NewEncoder(be).Encode(evtIntermediary)
		if err != nil {
			return &types.DocumentOpResult{Err: err}
		}
		docRes.Event = be
		docRes.StatusCode = resp.StatusCode
		return &docRes
	}
	message, err := prepareExceptionResponseFromMap(resp.StatusCode, respData)
	return &types.DocumentOpResult{
		Err:        err,
		Message:    message,
		StatusCode: resp.StatusCode,
	}
}

// GetDoc deals with retrieving a single document from the specified collection
// with the provided key.
func (c *clientImpl) GetDoc(coll string, key string) *types.DocumentResult {
	req := c.prepareRequest("GET", "/"+coll+"/"+key, nil, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.DocumentResult{
			Err: err,
		}
	}
	if resp.StatusCode == http.StatusOK {
		return &types.DocumentResult{
			Document:   resp.Body,
			StatusCode: resp.StatusCode,
		}
	}
	msg, err := prepareExceptionResponse(resp)
	return &types.DocumentResult{
		Err:        err,
		Message:    msg,
		StatusCode: resp.StatusCode,
	}
}

func (c *clientImpl) UpdateDoc(coll string, key string, doc interface{}) *types.DocumentOpResult {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(doc)
	if err != nil {
		return &types.DocumentOpResult{Err: err}
	}
	req := c.prepareRequest("PUT", "/"+coll+"/"+key, nil, b)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &types.DocumentOpResult{
			Err: err,
		}
	}
	if resp.StatusCode == http.StatusOK {
		docOpInfo := types.DocumentOpResult{}
		// First try to decode the response body as a map
		// to retrieve the document and event.
		docOpInfo.StatusCode = resp.StatusCode
		var intermediary = struct {
			Document map[string]interface{} `json:"doc"`
			Event    map[string]interface{} `json:"event"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&intermediary)
		if err != nil {
			docOpInfo.Err = err
		} else {
			// Now encode the decoded abstract maps to a buffer of bytes representing
			// the JSON representation to allow the user to decode the data to an application specific
			// struct type for both documents and events.
			bd := new(bytes.Buffer)
			be := new(bytes.Buffer)
			err = json.NewEncoder(bd).Encode(intermediary.Document)
			if err != nil {
				return &types.DocumentOpResult{Err: err}
			}
			err = json.NewEncoder(be).Encode(intermediary.Event)
			if err != nil {
				return &types.DocumentOpResult{Err: err}
			}
			docOpInfo.Document = bd
			docOpInfo.Event = be
		}
		return &docOpInfo
	}
	msg, err := prepareExceptionResponse(resp)
	return &types.DocumentOpResult{
		Err:        err,
		Message:    msg,
		StatusCode: resp.StatusCode,
	}
}
