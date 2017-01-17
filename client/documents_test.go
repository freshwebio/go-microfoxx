package client_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"time"

	. "github.com/freshwebio/go-microfoxx/client"
	"github.com/freshwebio/go-microfoxx/types"
	. "gopkg.in/check.v1"
)

type DocumentsSuite struct {
	client Client
}

type documentsTestClient struct {
	dummySessionClient
}

type documentTestModel struct {
	Id     string `json:"_id"`
	Key    string `json:"_key"`
	Rating string `json:"rating"`
	Height string `json:"height"`
}

type eventTestModel struct {
	Type    string        `json:"type"`
	Op      string        `json:"op"`
	Data    string        `json:"data"`
	Created time.Duration `json:"created"`
}

func newDocumentsTestHttpClient() WebClient {
	tc := &documentsTestClient{}
	return tc
}

// Deals with preparing a response for document requests.
func (c *documentsTestClient) Do(req *http.Request) (resp *http.Response, err error) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(req.URL.Path, "/_db//microfoxx")
		docRegExp := regexp.MustCompile("^/(\\w+)(\\?(.*))?$")
		countRegExp := regexp.MustCompile("^/(\\w+)/count(\\?(.*))?$")
		docKeyRegExp := regexp.MustCompile("^/(\\w+)/(\\w+)(\\?(.*))?$")
		if countRegExp.MatchString(path) && r.Method == "GET" {
			parts := countRegExp.FindStringSubmatch(path)
			c.countDocs(w, r, parts[1])
		} else if docKeyRegExp.MatchString(path) && r.Method == "DELETE" {
			parts := docKeyRegExp.FindStringSubmatch(path)
			c.removeDoc(w, r, parts[1], parts[2])
		} else if docKeyRegExp.MatchString(path) && r.Method == "GET" {
			parts := docKeyRegExp.FindStringSubmatch(path)
			c.getDoc(w, r, parts[1], parts[2])
		} else if docKeyRegExp.MatchString(path) && r.Method == "PUT" {
			parts := docKeyRegExp.FindStringSubmatch(path)
			c.updateDoc(w, r, parts[1], parts[2])
		} else if docRegExp.MatchString(path) && r.Method == "GET" {
			parts := docRegExp.FindStringSubmatch(path)
			c.getDocs(w, r, parts[1])
		} else if docRegExp.MatchString(path) && r.Method == "POST" {
			parts := docRegExp.FindStringSubmatch(path)
			c.createDoc(w, r, parts[1])
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

func (c *documentsTestClient) getDocs(w http.ResponseWriter, req *http.Request, coll string) {
	// First of all ensure that the provided collection is test.
	if coll != "test" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte("{\"exception\":\"Error 2016: that collection doesn't exist\"}"))
		return
	}
	// Now try to retrieve the expected field parameter provided by the test.
	rating := req.URL.Query().Get("rating")
	if rating == "high" {
		w.WriteHeader(http.StatusOK)
		documents := []*documentTestModel{
			{
				Id:     "test/ab54fgd3",
				Key:    "ab54fgd3",
				Rating: "5",
				Height: "180cm",
			},
			{
				Id:     "test/bd53fgd3",
				Key:    "bd53fgd3",
				Rating: "4",
				Height: "185cm",
			},
			{
				Id:     "test/cq53egd3",
				Key:    "cq53egd3",
				Rating: "2",
				Height: "172cm",
			},
			{
				Id:     "test/aq53agd3",
				Key:    "aq53agd3",
				Rating: "5",
				Height: "177cm",
			},
			{
				Id:     "test/xq533gd3",
				Key:    "xq533gd3",
				Rating: "4",
				Height: "167cm",
			},
		}
		respBytes, _ := json.Marshal(documents)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(respBytes)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte("{\"exception\":\"Error 2016: No documents found\"}"))
	}
}

func (c *documentsTestClient) countDocs(w http.ResponseWriter, req *http.Request, coll string) {
	if coll == "test" {
		if req.URL.Query().Get("rating") != "" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte("{\"count\":65}"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte("{\"count\":34}"))
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte("{\"exception\":\"Error 2016: That collection doesn't exist\"}"))
	}
}

func (c *documentsTestClient) createDoc(w http.ResponseWriter, req *http.Request, coll string) {
	if coll == "test" {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		respData := make(map[string]interface{})
		respData["doc"] = documentTestModel{}
		respData["event"] = eventTestModel{}
		b, _ := json.Marshal(respData)
		w.Write(b)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte("{\"exception\":\"Error 2016: that collection doesn't exist\"}"))
	}
}

func (c *documentsTestClient) removeDoc(w http.ResponseWriter, req *http.Request, coll string, key string) {
	if coll == "test" {
		if key == "gt543d" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			respData := make(map[string]interface{})
			respData["doc"] = documentTestModel{}
			respData["event"] = eventTestModel{}
			b, _ := json.Marshal(respData)
			w.Write(b)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte("{\"exception\":\"Error 2016: We couldn't find the resource you intend to delete\"}"))
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte("{\"exception\":\"Error 2016: that collection doesn't exist\"}"))
	}
}

func (c *documentsTestClient) getDoc(w http.ResponseWriter, req *http.Request, coll string, key string) {
	if coll == "test" {
		if key == "ab321e" {
			doc := documentTestModel{}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(doc)
			w.Write(b.Bytes())
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte("{\"exception\":\"Error 2016: We couldn't find the document you specified\"}"))
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte("{\"exception\":\"Error 2016: that collection doesn't exist\"}"))
	}
}

func (c *documentsTestClient) updateDoc(w http.ResponseWriter, req *http.Request, coll string, key string) {
	if coll == "test" {
		if key == "ab321e" {
			// First of al read the document to be updated from the request.
			respBody := make(map[string]interface{})
			doc := documentTestModel{}
			json.NewDecoder(req.Body).Decode(&doc)
			respBody["doc"] = doc
			respBody["event"] = &eventTestModel{}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(respBody)
			w.Write(b.Bytes())
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte("{\"exception\":\"Error 2016: We couldn't find the document you intend to update\"}"))
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte("{\"exception\":\"Error 2016: that collection doesn't exist\"}"))
	}
}

var _ = Suite(&DocumentsSuite{})

func (s *DocumentsSuite) SetUpSuite(c *C) {
	// Simply provide an empty set of connection parameters as our test HTTP client
	// doesn't care about the url, just the request body for testing the cursor functionality.
	cli, err := NewClient(&types.ConnectionParams{}, newDocumentsTestHttpClient())
	if err != nil {
		c.Error("Failed to setup our client for testing.")
	}
	s.client = cli
}

func (s *DocumentsSuite) TestGetDocs(c *C) {
	// Set up our parameters to retrieve a
	// set of documents from our test store.
	params := &types.DocumentRetrievalParams{
		Fields: map[string]string{
			"rating": "high",
		},
		SortFields:  []string{"height"},
		SortOrder:   "DESC",
		LimitOffset: 10,
		LimitCount:  5,
	}
	res := s.client.GetDocs("test", params)
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.StatusCode, Equals, http.StatusOK)
	c.Assert(res.Message, Equals, "")
	// Now try to parse the result.
	var docs []documentTestModel
	err := json.NewDecoder(res.Documents).Decode(&docs)
	if err != nil {
		c.Error("Failed to decode response documents")
	}
	c.Assert(len(docs), Equals, 5)
	// Now try with a query that should
	// give us an empty set of documents.
	params.Fields = map[string]string{
		"rating": "mad",
	}
	res = s.client.GetDocs("test", params)
	c.Assert(res.StatusCode, Equals, http.StatusNotFound)
	c.Assert(res.Err, Equals, ErrNotFound)
	c.Assert(res.Message, Equals, "Error 2016: No documents found")
	c.Assert(res.Documents, Equals, nil)
	// Now try with a query to a collection that doesn't exist.
	res = s.client.GetDocs("cars", params)
	c.Assert(res.StatusCode, Equals, http.StatusBadRequest)
	c.Assert(res.Err, Equals, ErrBadRequest)
	c.Assert(res.Message, Equals, "Error 2016: that collection doesn't exist")
	c.Assert(res.Documents, Equals, nil)
}

func (s *DocumentsSuite) TestCreateDocs(c *C) {
	doc := documentTestModel{}
	doc.Rating = "high"
	doc.Height = "184cm"
	res := s.client.CreateDoc("test", doc)
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.StatusCode, Equals, http.StatusCreated)
	// Make sure we can extract the event.
	var event eventTestModel
	err := json.NewDecoder(res.Event).Decode(&event)
	if err != nil {
		c.Error("Failed to decode the event")
	}
	// Make sure we can also extract the document which now contains
	// the unique ID and any relationship information.
	var respDoc documentTestModel
	err = json.NewDecoder(res.Document).Decode(&respDoc)
	if err != nil {
		c.Error("Failed to decode the document")
	}
	// Now ensure we get the correct error response when trying to create
	// a document in a collection that doesn't exist.
	res = s.client.CreateDoc("cars", doc)
	c.Assert(res.Err, Equals, ErrBadRequest)
	c.Assert(res.Message, Equals, "Error 2016: that collection doesn't exist")
	c.Assert(res.Event, Equals, nil)
	c.Assert(res.Document, Equals, nil)
}

func (s *DocumentsSuite) TestGetDocCount(c *C) {
	res := s.client.GetDocCount("test", &types.DocumentRetrievalParams{})
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.Count, Equals, 34)
	params := &types.DocumentRetrievalParams{
		Fields: map[string]string{
			"rating": "high",
		},
	}
	res = s.client.GetDocCount("test", params)
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.Count, Equals, 65)
	res = s.client.GetDocCount("cars", &types.DocumentRetrievalParams{})
	c.Assert(res.Err, Equals, ErrBadRequest)
	c.Assert(res.Count, Equals, -1)
	c.Assert(res.Message, Equals, "Error 2016: That collection doesn't exist")
}

func (s *DocumentsSuite) TestRemoveDoc(c *C) {
	// First of all try to remove a document that doesn't exist.
	res := s.client.RemoveDoc("test", "g54325fgdf")
	c.Assert(res.Err, Equals, ErrNotFound)
	c.Assert(res.Message, Equals, "Error 2016: We couldn't find the resource you intend to delete")
	c.Assert(res.StatusCode, Equals, http.StatusNotFound)
	// Now try to remove a document from a non-existent collection.
	res = s.client.RemoveDoc("cars", "fe34ff21dfa9b")
	c.Assert(res.Err, Equals, ErrBadRequest)
	c.Assert(res.Message, Equals, "Error 2016: that collection doesn't exist")
	c.Assert(res.StatusCode, Equals, http.StatusBadRequest)
	// Now try to remove a document that does exist from an existing collection.
	res = s.client.RemoveDoc("test", "gt543d")
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.StatusCode, Equals, http.StatusOK)
	// Ensure we can decode the document and event.
	var respDoc documentTestModel
	err := json.NewDecoder(res.Document).Decode(&respDoc)
	if err != nil {
		c.Error("Failed to decode the document")
	}
	var respEvt eventTestModel
	err = json.NewDecoder(res.Event).Decode(&respEvt)
	if err != nil {
		c.Error("Failed to decode the event")
	}
}

func (s *DocumentsSuite) TestGetDoc(c *C) {
	// First of all try to retrieve a document from a collection that doesn't exist.
	res := s.client.GetDoc("cars", "ab321e")
	c.Assert(res.Err, Equals, ErrBadRequest)
	c.Assert(res.Message, Equals, "Error 2016: that collection doesn't exist")
	c.Assert(res.Document, Equals, nil)
	c.Assert(res.StatusCode, Equals, http.StatusBadRequest)
	// Now a non-existent item from a collection that exists.
	res = s.client.GetDoc("test", "bg542eq")
	c.Assert(res.Err, Equals, ErrNotFound)
	c.Assert(res.Message, Equals, "Error 2016: We couldn't find the document you specified")
	c.Assert(res.Document, Equals, nil)
	c.Assert(res.StatusCode, Equals, http.StatusNotFound)
	// Now try to retreive an existing document living in an existing collection.
	res = s.client.GetDoc("test", "ab321e")
	c.Assert(res.StatusCode, Equals, http.StatusOK)
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.Message, Equals, "")
	// Now try to decode the document from the response.
	var doc documentTestModel
	err := json.NewDecoder(res.Document).Decode(&doc)
	if err != nil {
		c.Error("Failed to decode document response")
	}
}

func (s *DocumentsSuite) TestUpdateDoc(c *C) {
	// Try to update a document in a collection that doesn't exist.
	res := s.client.UpdateDoc("cars", "ab321e", documentTestModel{})
	c.Assert(res.Err, Equals, ErrBadRequest)
	c.Assert(res.Message, Equals, "Error 2016: that collection doesn't exist")
	c.Assert(res.Document, Equals, nil)
	c.Assert(res.Event, Equals, nil)
	c.Assert(res.StatusCode, Equals, http.StatusBadRequest)
	// Now try updating a document that doesn't exist in the provided collection.
	res = s.client.UpdateDoc("test", "ab123edf", documentTestModel{})
	c.Assert(res.Err, Equals, ErrNotFound)
	c.Assert(res.Message, Equals, "Error 2016: We couldn't find the document you intend to update")
	c.Assert(res.Document, Equals, nil)
	c.Assert(res.Event, Equals, nil)
	c.Assert(res.StatusCode, Equals, http.StatusNotFound)
	// Finally attempt to update an existing document in an existing collection
	// expecting a result of the same document and an event.
	res = s.client.UpdateDoc("test", "ab321e", documentTestModel{
		Rating: "high",
		Height: "186cm",
	})
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.Message, Equals, "")
	c.Assert(res.StatusCode, Equals, http.StatusOK)
	// Ensure we can decode the document and an accompanying event.
	var doc documentTestModel
	err := json.NewDecoder(res.Document).Decode(&doc)
	if err != nil {
		c.Error("Failed to decode the document")
	}
	c.Assert(doc.Rating, Equals, "high")
	c.Assert(doc.Height, Equals, "186cm")
	var evt eventTestModel
	err = json.NewDecoder(res.Event).Decode(&evt)
	if err != nil {
		c.Error("Failed to decode the event")
	}
}
