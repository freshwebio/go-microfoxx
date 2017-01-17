package client_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/freshwebio/go-microfoxx/client"
	"github.com/freshwebio/go-microfoxx/types"
	. "gopkg.in/check.v1"
)

type QueriesSuite struct {
	client Client
}

type queriesTestClient struct {
	dummySessionClient
}

func newQueriesTestHttpClient() WebClient {
	tc := &queriesTestClient{}
	return tc
}

// Deals with preparing a response for cursor requests.
func (c *queriesTestClient) Do(req *http.Request) (resp *http.Response, err error) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/_db//microfoxx")
		switch path {
		case "/insert":
			c.insert(w, r)
		case "/update":
			c.update(w, r)
		case "/remove":
			c.remove(w, r)
		}
	}))
	defer server.Close()
	// Make the needed alterations to the request
	newReq, _ := http.NewRequest(req.Method, server.URL+req.URL.Path, req.Body)
	newReq.Header.Set("Content-Type", req.Header.Get("Content-Type"))
	newReq.URL.RawQuery = req.URL.RawQuery
	resp, err = http.DefaultClient.Do(newReq)
	return resp, err
}

func (c *queriesTestClient) insert(w http.ResponseWriter, req *http.Request) {
	// Try to retrieve the modifying query parameters from the request.
	var params types.ModifyingQueryParams
	json.NewDecoder(req.Body).Decode(&params)
	if params.Query == "FOR i in 1..100 INSERT { value: i, type: @type } IN test" {
		if _, exists := params.BindVars["type"]; exists {
			if params.WriteCollection == "test" {
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				respData := make(map[string]interface{})
				respData["docs"] = make([]map[string]interface{}, 0)
				respData["events"] = make([]map[string]interface{}, 0)
				b := new(bytes.Buffer)
				json.NewEncoder(b).Encode(respData)
				w.Write(b.Bytes())
			} else {
				w.WriteHeader(http.StatusPreconditionFailed)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.Write([]byte("{\"exception\":\"Error 2016: The collection to be written to was not defined as the write collection\"}"))
			}
		} else {
			w.WriteHeader(http.StatusPreconditionFailed)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte("{\"exception\":\"Error 2016: One or more of the bind variables specified in the query are missing\"}"))
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte("{\"exception\":\"Error 2016: The provided AQL query is not of the expected form\"}"))
	}
}

func (c *queriesTestClient) update(w http.ResponseWriter, req *http.Request) {
	var params types.ModifyingQueryParams
	json.NewDecoder(req.Body).Decode(&params)
	if params.Query == "FOR t IN test FILTER t.type=@type UPDATE t WITH { status: 'inactive' } IN test" {
		if _, exists := params.BindVars["type"]; exists {
			if params.WriteCollection == "test" {
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				respData := make(map[string]interface{})
				respData["docs"] = make([]map[string]interface{}, 0)
				respData["events"] = make([]map[string]interface{}, 0)
				b := new(bytes.Buffer)
				json.NewEncoder(b).Encode(respData)
				w.Write(b.Bytes())
			} else {
				w.WriteHeader(http.StatusPreconditionFailed)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.Write([]byte("{\"exception\":\"Error 2016: The collection to be written to was not defined as the write collection\"}"))
			}
		} else {
			w.WriteHeader(http.StatusPreconditionFailed)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte("{\"exception\":\"Error 2016: One or more of the bind variables specified in the query are missing\"}"))
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte("{\"exception\":\"Error 2016: The provided AQL query is not of the expected form\"}"))
	}
}

func (c *queriesTestClient) remove(w http.ResponseWriter, req *http.Request) {
	var params types.ModifyingQueryParams
	json.NewDecoder(req.Body).Decode(&params)
	if params.Query == "FOR t IN test FILTER t.type=@type REMOVE { _key: t._key } IN test" {
		if _, exists := params.BindVars["type"]; exists {
			if params.WriteCollection == "test" {
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				respData := make(map[string]interface{})
				respData["docs"] = make([]map[string]interface{}, 0)
				respData["events"] = make([]map[string]interface{}, 0)
				b := new(bytes.Buffer)
				json.NewEncoder(b).Encode(respData)
				w.Write(b.Bytes())
			} else {
				w.WriteHeader(http.StatusPreconditionFailed)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.Write([]byte("{\"exception\":\"Error 2016: The collection to be written to was not defined as the write collection\"}"))
			}
		} else {
			w.WriteHeader(http.StatusPreconditionFailed)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte("{\"exception\":\"Error 2016: One or more of the bind variables specified in the query are missing\"}"))
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte("{\"exception\":\"Error 2016: The provided AQL query is not of the expected form\"}"))
	}
}

var _ = Suite(&QueriesSuite{})

func (s *QueriesSuite) SetUpSuite(c *C) {
	// Simply provide an empty set of connection parameters as our test HTTP client
	// doesn't care about the url, just the request body for testing the cursor functionality.
	cli, err := NewClient(&types.ConnectionParams{}, newQueriesTestHttpClient())
	if err != nil {
		c.Error("Failed to setup our client for testing.")
	}
	s.client = cli
}

func (s *QueriesSuite) TestInsertQuery(c *C) {
	// First of all try to write an invalid query.
	params := &types.ModifyingQueryParams{
		Query: "For i BIN 1..100 INSERT {value:'dfgd'} IN test",
	}
	res := s.client.InsertQuery(params)
	c.Assert(res.Err, Equals, ErrBadRequest)
	c.Assert(res.Message, Equals, "Error 2016: The provided AQL query is not of the expected form")
	c.Assert(res.StatusCode, Equals, http.StatusBadRequest)
	c.Assert(res.Documents, Equals, nil)
	c.Assert(res.Events, Equals, nil)
	// Now try to write a valid query with missing bind variables.
	params.Query = "FOR i in 1..100 INSERT { value: i, type: @type } IN test"
	res = s.client.InsertQuery(params)
	c.Assert(res.Err, Equals, ErrGeneral)
	c.Assert(res.Message, Equals, "Error 2016: One or more of the bind variables specified in the query are missing")
	c.Assert(res.StatusCode, Equals, http.StatusPreconditionFailed)
	c.Assert(res.Documents, Equals, nil)
	c.Assert(res.Events, Equals, nil)
	// Now try to write specifying the wrong write collection.
	params.BindVars = map[string]interface{}{
		"type": "test",
	}
	params.WriteCollection = "user"
	res = s.client.InsertQuery(params)
	c.Assert(res.Err, Equals, ErrGeneral)
	c.Assert(res.Message, Equals, "Error 2016: The collection to be written to was not defined as the write collection")
	c.Assert(res.StatusCode, Equals, http.StatusPreconditionFailed)
	c.Assert(res.Documents, Equals, nil)
	c.Assert(res.Events, Equals, nil)
	// Now try to write with a valid request.
	params.WriteCollection = "test"
	res = s.client.InsertQuery(params)
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.Message, Equals, "")
	// Now ensure we can successfully decode the list of documents and events.
	var docs []map[string]string
	err := json.NewDecoder(res.Documents).Decode(&docs)
	if err != nil {
		c.Error("Failed to decode documents")
	}
	var evts []map[string]string
	err = json.NewDecoder(res.Events).Decode(&evts)
	if err != nil {
		c.Error("Failed to decode events")
	}
}

func (s *QueriesSuite) TestUpdateQuery(c *C) {
	// First of all try to write an invalid query.
	params := &types.ModifyingQueryParams{
		Query: "For t BIN test FILTER type=@ttype UPDATE WITH {value:'dfgd'} IN test",
	}
	res := s.client.UpdateQuery(params)
	c.Assert(res.Err, Equals, ErrBadRequest)
	c.Assert(res.Message, Equals, "Error 2016: The provided AQL query is not of the expected form")
	c.Assert(res.StatusCode, Equals, http.StatusBadRequest)
	c.Assert(res.Documents, Equals, nil)
	c.Assert(res.Events, Equals, nil)
	// Now try to write a valid query with missing bind variables.
	params.Query = "FOR t IN test FILTER t.type=@type UPDATE t WITH { status: 'inactive' } IN test"
	res = s.client.UpdateQuery(params)
	c.Assert(res.Err, Equals, ErrGeneral)
	c.Assert(res.Message, Equals, "Error 2016: One or more of the bind variables specified in the query are missing")
	c.Assert(res.StatusCode, Equals, http.StatusPreconditionFailed)
	c.Assert(res.Documents, Equals, nil)
	c.Assert(res.Events, Equals, nil)
	// Now try to write specifying the wrong write collection.
	params.BindVars = map[string]interface{}{
		"type": "test",
	}
	params.WriteCollection = "user"
	res = s.client.UpdateQuery(params)
	c.Assert(res.Err, Equals, ErrGeneral)
	c.Assert(res.Message, Equals, "Error 2016: The collection to be written to was not defined as the write collection")
	c.Assert(res.StatusCode, Equals, http.StatusPreconditionFailed)
	c.Assert(res.Documents, Equals, nil)
	c.Assert(res.Events, Equals, nil)
	// Now try to write with a valid request.
	params.WriteCollection = "test"
	res = s.client.UpdateQuery(params)
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.Message, Equals, "")
	// Now ensure we can successfully decode the list of documents and events.
	var docs []map[string]string
	err := json.NewDecoder(res.Documents).Decode(&docs)
	if err != nil {
		c.Error("Failed to decode documents")
	}
	var evts []map[string]string
	err = json.NewDecoder(res.Events).Decode(&evts)
	if err != nil {
		c.Error("Failed to decode events")
	}
}

func (s *QueriesSuite) TestRemoveQuery(c *C) {
	// First of all try to write an invalid query.
	params := &types.ModifyingQueryParams{
		Query: "For t BIN test FILTER type=@ttype REMOVE {_key: t._key} IN test",
	}
	res := s.client.RemoveQuery(params)
	c.Assert(res.Err, Equals, ErrBadRequest)
	c.Assert(res.Message, Equals, "Error 2016: The provided AQL query is not of the expected form")
	c.Assert(res.StatusCode, Equals, http.StatusBadRequest)
	c.Assert(res.Documents, Equals, nil)
	c.Assert(res.Events, Equals, nil)
	// Now try to write a valid query with missing bind variables.
	params.Query = "FOR t IN test FILTER t.type=@type REMOVE { _key: t._key } IN test"
	res = s.client.RemoveQuery(params)
	c.Assert(res.Err, Equals, ErrGeneral)
	c.Assert(res.Message, Equals, "Error 2016: One or more of the bind variables specified in the query are missing")
	c.Assert(res.StatusCode, Equals, http.StatusPreconditionFailed)
	c.Assert(res.Documents, Equals, nil)
	c.Assert(res.Events, Equals, nil)
	// Now try to write specifying the wrong write collection.
	params.BindVars = map[string]interface{}{
		"type": "test",
	}
	params.WriteCollection = "user"
	res = s.client.RemoveQuery(params)
	c.Assert(res.Err, Equals, ErrGeneral)
	c.Assert(res.Message, Equals, "Error 2016: The collection to be written to was not defined as the write collection")
	c.Assert(res.StatusCode, Equals, http.StatusPreconditionFailed)
	c.Assert(res.Documents, Equals, nil)
	c.Assert(res.Events, Equals, nil)
	// Now try to write with a valid request.
	params.WriteCollection = "test"
	res = s.client.RemoveQuery(params)
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.Message, Equals, "")
	// Now ensure we can successfully decode the list of documents and events.
	var docs []map[string]string
	err := json.NewDecoder(res.Documents).Decode(&docs)
	if err != nil {
		c.Error("Failed to decode documents")
	}
	var evts []map[string]string
	err = json.NewDecoder(res.Events).Decode(&evts)
	if err != nil {
		c.Error("Failed to decode events")
	}
}
