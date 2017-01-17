package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	. "github.com/freshwebio/go-microfoxx/client"
	"github.com/freshwebio/go-microfoxx/types"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type CollectionSuite struct {
	client Client
}

type collectionTestClient struct {
	dummySessionClient
	collections map[string]string
}

// Deals with preparing a response based on whether or not a collection already exists.
func (c *collectionTestClient) Do(req *http.Request) (resp *http.Response, err error) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		collData := struct {
			Name string `json:"name"`
		}{}
		json.NewDecoder(req.Body).Decode(&collData)
		if _, exists := c.collections[collData.Name]; !exists {
			nextID := strconv.Itoa(len(c.collections))
			c.collections[collData.Name] = nextID
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{\"message\":\"Successfully created the " + collData.Name + " collection\",\"_id\":\"" +
				nextID + "\"}"))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{\"exception\":\"Error 2016: Collection already exists\"}"))
		}
	}))
	defer server.Close()
	resp, err = http.Post(server.URL, req.Header.Get("Content-Type"), req.Body)
	return resp, err
}

func newCollectionTestHttpClient() WebClient {
	tc := &collectionTestClient{}
	tc.collections = make(map[string]string)
	return tc
}

var _ = Suite(&CollectionSuite{})

func (s *CollectionSuite) SetUpSuite(c *C) {
	// Simply provide an empty set of connection parameters as our test HTTP client
	// doesn't care about the url, just the request body for testing the single
	// collection focused method.
	cli, err := NewClient(&types.ConnectionParams{}, newCollectionTestHttpClient())
	if err != nil {
		c.Error("Failed to setup our client for testing.")
	}
	s.client = cli
}

func (s *CollectionSuite) TestCreateColl(c *C) {
	// First of all ensure creating a non-existent collection gives
	// us the correct response.
	res := s.client.CreateColl("test")
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.StatusCode, Equals, http.StatusCreated)
	c.Assert(res.Message, Equals, "Successfully created the test collection")
	// Now ensure that trying to create an already existing returns
	// an error and the correct error message.
	res = s.client.CreateColl("test")
	c.Assert(res.Err, Not(Equals), nil)
	c.Assert(res.StatusCode, Equals, http.StatusBadRequest)
	c.Assert(res.Message, Equals, "Error 2016: Collection already exists")
}
