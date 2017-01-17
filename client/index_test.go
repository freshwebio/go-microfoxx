package client_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"

	. "github.com/freshwebio/go-microfoxx/client"
	"github.com/freshwebio/go-microfoxx/types"
	. "gopkg.in/check.v1"
)

type IndexSuite struct {
	client Client
}

type indexTestClient struct {
	dummySessionClient
}

// Deals with preparing a response for the index test cases.
func (c *indexTestClient) Do(req *http.Request) (resp *http.Response, err error) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/index") && r.Method == "POST" {
			c.createIndex(w, r)
		} else {
			// Now attempt to match regular expressions for deletion
			// and listing indexes.
			deletePattern := regexp.MustCompile(".*/index/(\\w+)/(\\w+)")
			listPattern := regexp.MustCompile(".*/index/(\\w+)")
			if deletePattern.MatchString(r.URL.Path) && r.Method == "DELETE" {
				parts := deletePattern.FindStringSubmatch(r.URL.Path)
				c.removeIndex(w, r, parts[1], parts[2])
			} else if listPattern.MatchString(r.URL.Path) && r.Method == "GET" {
				parts := listPattern.FindStringSubmatch(r.URL.Path)
				c.getIndexes(w, r, parts[1])
			}
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

func (c *indexTestClient) createIndex(w http.ResponseWriter, r *http.Request) {
	// First try to decode the request body and ensure it is a valid set of index parameters.
	var params *types.IndexParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"exception\":\"Error 2016: The request body couldn't be decoded to a set of index parameters\"}"))
	}
	if params.Collection == "test" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("{\"message\":\"Successfully created the index for the specified collection\"}"))
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusPreconditionFailed)
		w.Write([]byte("{\"exception\":\"Error 2016: The provided collection doesn't exist\"}"))
	}
}

func (c *indexTestClient) getIndexes(w http.ResponseWriter, r *http.Request, collection string) {
	if collection == "test" {
		indexes := []*types.Index{
			{
				ID:                  "test/1",
				Type:                "primary",
				Fields:              []string{"_key"},
				SelectivityEstimate: 1,
				Unique:              true,
				Sparse:              true,
			},
			{
				ID:                  "test/2",
				Type:                "hash",
				Fields:              []string{"attr1"},
				SelectivityEstimate: 1,
				Unique:              true,
				Sparse:              true,
			},
		}
		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(indexes)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(b.Bytes())
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusPreconditionFailed)
		w.Write([]byte("{\"exception\":\"Error 2016: The specified collection doesn't exist\"}"))
	}
}

func (c *indexTestClient) removeIndex(w http.ResponseWriter, r *http.Request, collection string, key string) {
	if collection == "test" {
		if key == "3210043921" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{\"message\":\"Successfully removed the index specified\"}"))
		} else {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("{\"exception\":\"Error 2016: The index with the provided handle doesn't exist\"}"))
		}
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusPreconditionFailed)
		w.Write([]byte("{\"exception\":\"Error 2016: The specified collection doesn't exist\"}"))
	}
}

func newIndexTestHttpClient() WebClient {
	return &indexTestClient{}
}

var _ = Suite(&IndexSuite{})

func (s *IndexSuite) SetUpSuite(c *C) {
	// Simply provide an empty set of connection parameters as our test HTTP client
	// doesn't care about the url, just the request body for testing the index functionality.
	cli, err := NewClient(&types.ConnectionParams{}, newIndexTestHttpClient())
	if err != nil {
		c.Error("Failed to setup our client for testing.")
	}
	s.client = cli
}

func (s *IndexSuite) TestGetIndexes(c *C) {
	// First ensure we get the correct error message for the wrong collection.
	res := s.client.GetIndexes("user")
	c.Assert(res.Err, Equals, ErrGeneral)
	c.Assert(res.Message, Equals, "Error 2016: The specified collection doesn't exist")
	c.Assert(res.StatusCode, Equals, http.StatusPreconditionFailed)
	// Now ensure we get the correct result set for an existing collection.
	res = s.client.GetIndexes("test")
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.Message, Equals, "")
	c.Assert(len(res.Indexes), Equals, 2)
	c.Assert(res.StatusCode, Equals, http.StatusOK)
}

func (s *IndexSuite) TestCreateIndex(c *C) {
	// First ensure we can't create a new index for a non-existent collection.
	p := &types.IndexParams{
		Collection: "user",
		Type:       "hash",
		Fields:     []string{},
	}
	res := s.client.CreateIndex(p)
	c.Assert(res.Err, Equals, ErrGeneral)
	c.Assert(res.Message, Equals, "Error 2016: The provided collection doesn't exist")
	c.Assert(res.StatusCode, Equals, http.StatusPreconditionFailed)
	p.Collection = "test"
	res = s.client.CreateIndex(p)
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.Message, Equals, "")
	c.Assert(res.StatusCode, Equals, http.StatusCreated)
}

func (s *IndexSuite) TestRemoveIndex(c *C) {
	// First of all try to remove an index from a collection that doesn't exist.
	res := s.client.RemoveIndex("user/324542")
	c.Assert(res.Err, Equals, ErrGeneral)
	c.Assert(res.Message, Equals, "Error 2016: The specified collection doesn't exist")
	c.Assert(res.StatusCode, Equals, http.StatusPreconditionFailed)
	// Now try to remove a non-existent index.
	res = s.client.RemoveIndex("test/34534235324")
	c.Assert(res.Err, Equals, ErrNotFound)
	c.Assert(res.Message, Equals, "Error 2016: The index with the provided handle doesn't exist")
	c.Assert(res.StatusCode, Equals, http.StatusNotFound)
	// Now try to remove an existing index.
	res = s.client.RemoveIndex("test/3210043921")
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.Message, Equals, "Successfully removed the index specified")
	c.Assert(res.StatusCode, Equals, http.StatusOK)
}
