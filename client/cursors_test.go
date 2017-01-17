package client_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"

	. "github.com/freshwebio/go-microfoxx/client"
	"github.com/freshwebio/go-microfoxx/types"
	. "gopkg.in/check.v1"
)

type CursorsSuite struct {
	client Client
}

type testCursor struct {
	Documents []map[string]interface{}
	BatchSize int
	Current   int
}

type cursorsTestClient struct {
	cursors   map[string]*testCursor
	documents []map[string]interface{}
	dummySessionClient
}

func newCursorsTestHttpClient() WebClient {
	tc := &cursorsTestClient{}
	tc.cursors = make(map[string]*testCursor)
	tc.documents = make([]map[string]interface{}, 0)
	for i := 1; i <= 50; i++ {
		var status string
		if i%2 == 0 {
			status = "enabled"
		} else {
			status = "disabled"
		}
		tc.documents = append(tc.documents, map[string]interface{}{
			"name":   "testname" + strconv.Itoa(i),
			"status": status,
		})
	}
	return tc
}

// Deals with preparing a response for cursor requests.
func (c *cursorsTestClient) Do(req *http.Request) (resp *http.Response, err error) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r := regexp.MustCompile(".*/cursor/(\\w+)")
		if strings.HasSuffix(req.URL.Path, "/cursor") {
			c.newCursor(w, req)
		} else if r.MatchString(req.URL.Path) {
			parts := r.FindStringSubmatch(req.URL.Path)
			c.nextBatch(w, req, parts[1])
		}
	}))
	defer server.Close()
	resp, err = http.Post(server.URL+req.URL.Path, req.Header.Get("Content-Type"), req.Body)
	return resp, err
}

func (c *cursorsTestClient) getDocsWithPropertyVal(property string, val string) []map[string]interface{} {
	docs := make([]map[string]interface{}, 0)
	for _, doc := range c.documents {
		if doc[property] == val {
			docs = append(docs, doc)
		}
	}
	return docs
}

func (c *cursorsTestClient) newCursor(w http.ResponseWriter, req *http.Request) {
	cursorParams := types.CursorQueryParams{}
	json.NewDecoder(req.Body).Decode(&cursorParams)
	_, collExists := cursorParams.BindVars["@coll"]
	status, statusExists := cursorParams.BindVars["status"]
	name, nameExists := cursorParams.BindVars["name"]
	if collExists && (statusExists || nameExists) {
		r := regexp.MustCompile("FOR item in @@coll FILTER item.\\w+ == @(\\w+)")
		// For the purpose of testing determine whether the query is valid based on whether
		// it validates with our simple regular expression pattern.
		if r.MatchString(cursorParams.Query) {
			// Get the name of the parameter to filter by.
			parts := r.FindStringSubmatch(cursorParams.Query)
			property := parts[1]
			var val string
			if nameExists {
				val = name.(string)
			} else if statusExists {
				val = status.(string)
			}
			results := c.getDocsWithPropertyVal(property, val)
			cursor := testCursor{}
			cursor.Documents = results
			// Only create a new cursor where the batch size is set to something other than 0.
			var respMap map[string]interface{}
			if cursorParams.BatchSize > 0 && cursorParams.BatchSize < len(results) {
				nextID := strconv.Itoa(len(c.cursors))
				cursor.Current = 0
				cursor.BatchSize = cursorParams.BatchSize
				c.cursors[nextID] = &cursor
				batch := make([]map[string]interface{}, 0)
				var i int
				for i = 0; i < cursor.Current+cursor.BatchSize; i++ {
					batch = append(batch, results[i])
				}
				cursor.Current = 0
				respMap = make(map[string]interface{})
				respMap["results"] = batch
				respMap["hasMore"] = true
				respMap["cursor"] = nextID
				if cursorParams.Count {
					respMap["count"] = len(cursor.Documents)
				}
			} else {
				// Simply return a request with the results of the query.
				respMap = make(map[string]interface{})
				respMap["results"] = cursor.Documents
				respMap["hasMore"] = false
				if cursorParams.Count {
					respMap["count"] = len(cursor.Documents)
				}
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			// Now prepare the response body to be written.
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(respMap)
			w.Write(b.Bytes())
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{\"exception\":\"Error 2016: Invalid AQL query\"}"))
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"exception\":\"Error 2016: Bind parameters in query need to be provided\"}"))
	}
}

func (c *cursorsTestClient) nextBatch(w http.ResponseWriter, req *http.Request, cursorID string) {
	// First of all we need to attempt to retrieve the cursor with the provided ID.
	if cursor, exists := c.cursors[cursorID]; exists {
		i := 0
		batch := make([]map[string]interface{}, 0)
		var max int
		if len(cursor.Documents) < cursor.Current+cursor.BatchSize {
			max = len(cursor.Documents)
		} else {
			max = cursor.Current + cursor.BatchSize
		}
		for i = cursor.Current; i < max; i++ {
			batch = append(batch, cursor.Documents[i])
		}
		cursor.Current = i
		var hasMore bool
		if cursor.Current < len(cursor.Documents) {
			hasMore = true
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		dataToSend := struct {
			HasMore bool                     `json:"hasMore"`
			Results []map[string]interface{} `json:"results"`
		}{
			HasMore: hasMore,
			Results: batch,
		}
		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(dataToSend)
		w.Write(b.Bytes())
		// In the case there are no more results simply remove the cursor.
		if !hasMore {
			delete(c.cursors, cursorID)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"exception\":\"Error 2016: Cursor not found\"}"))
	}
}

var _ = Suite(&CursorsSuite{})

func (s *CursorsSuite) SetUpSuite(c *C) {
	// Simply provide an empty set of connection parameters as our test HTTP client
	// doesn't care about the url, just the request body for testing the cursor functionality.
	cli, err := NewClient(&types.ConnectionParams{}, newCursorsTestHttpClient())
	if err != nil {
		c.Error("Failed to setup our client for testing.")
	}
	s.client = cli
}

func (s *CursorsSuite) TestCursorQuery(c *C) {
	// First of all test that we get the correct response
	// for a cursor query including the count.
	res := s.client.CursorQuery(&types.CursorQueryParams{
		Query: "FOR item in @@coll FILTER item.name == @name",
		BindVars: map[string]interface{}{
			"@coll": "users",
			"name":  "testname1",
		},
		Count: true,
	})
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.StatusCode, Equals, http.StatusOK)
	docs := make([]map[string]interface{}, 0)
	json.NewDecoder(res.Documents).Decode(&docs)
	c.Assert(len(docs), Equals, 1)
	// Now ensure an invalid query gets a bad request response.
	res = s.client.CursorQuery(&types.CursorQueryParams{
		Query: "FORINFFF item in @@coll FILTERO item.name == @name",
		BindVars: map[string]interface{}{
			"@coll": "users",
			"name":  "testname",
		},
		Count: true,
	})
	c.Assert(res.Err, Not(Equals), nil)
	c.Assert(res.StatusCode, Equals, http.StatusBadRequest)
	c.Assert(res.Message, Equals, "Error 2016: Invalid AQL query")
	// Ensure that a query with bind parameters that do not exist gets an invalid response.
	res = s.client.CursorQuery(&types.CursorQueryParams{
		Query:    "FOR item in @@coll FILTER item.name == @name",
		BindVars: map[string]interface{}{},
	})
	c.Assert(res.Err, Not(Equals), nil)
	c.Assert(res.StatusCode, Equals, http.StatusBadRequest)
	c.Assert(res.Message, Equals, "Error 2016: Bind parameters in query need to be provided")
}

func (s *CursorsSuite) TestCursorGetNextBatch(c *C) {
	// First of all create the initial cursor query.
	res := s.client.CursorQuery(&types.CursorQueryParams{
		Query: "FOR item in @@coll FILTER item.status == @status",
		BindVars: map[string]interface{}{
			"@coll":  "users",
			"status": "enabled",
		},
		BatchSize: 5,
	})
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.StatusCode, Equals, http.StatusOK)
	c.Assert(res.HasMore, Equals, true)
	c.Assert(res.Cursor, Not(Equals), "")
	// Now ensure we can get the next batch 5 times as the provided
	// batch size warrants five batches of five with 25 results.
	cid := res.Cursor
	for i := 0; i < 5; i++ {
		batchRes := s.client.CursorGetNextBatch(cid)
		c.Assert(batchRes.Err, Equals, nil)
		c.Assert(batchRes.StatusCode, Equals, http.StatusOK)
		if i < 4 {
			c.Assert(batchRes.HasMore, Equals, true)
		} else {
			c.Assert(batchRes.HasMore, Equals, false)
		}
		var batchDocs []map[string]interface{}
		json.NewDecoder(batchRes.Documents).Decode(&batchDocs)
		c.Assert(len(batchDocs), Equals, 5)
	}
	// Ensure our cursor is no longer available.
	batchRes := s.client.CursorGetNextBatch(cid)
	c.Assert(batchRes.Err, Not(Equals), nil)
	c.Assert(batchRes.Message, Equals, "Error 2016: Cursor not found")
	// Now ensure an invalid response for getting the next batch
	// of a non-existent cursor.
	batchRes = s.client.CursorGetNextBatch("43")
	c.Assert(batchRes.Err, Not(Equals), nil)
	c.Assert(batchRes.Message, Equals, "Error 2016: Cursor not found")
	c.Assert(batchRes.HasMore, Equals, false)
	c.Assert(batchRes.StatusCode, Equals, http.StatusBadRequest)
}
