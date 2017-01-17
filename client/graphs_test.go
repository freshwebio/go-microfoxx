package client_test

import (
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

type GraphSuite struct {
	client Client
}

type graphTestClient struct {
	dummySessionClient
	graphs         map[string]*types.Graph
	graphNameIDMap map[string]string
}

func newGraphTestHttpClient() WebClient {
	tc := &graphTestClient{}
	tc.graphs = make(map[string]*types.Graph)
	tc.graphNameIDMap = make(map[string]string)
	return tc
}

// Deals with preparing a response based on whether or not a graph already exists.
func (c *graphTestClient) Do(req *http.Request) (resp *http.Response, err error) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r := regexp.MustCompile(".*/graph/(\\w+)/relation")
		if strings.HasSuffix(req.URL.Path, "/graph") {
			c.createGraph(w, req)
		} else if r.MatchString(req.URL.Path) {
			parts := r.FindStringSubmatch(req.URL.Path)
			c.createRelation(w, req, parts[1])
		}
	}))
	defer server.Close()
	resp, err = http.Post(server.URL+req.URL.Path, req.Header.Get("Content-Type"), req.Body)
	return resp, err
}

func (c *graphTestClient) createGraph(w http.ResponseWriter, req *http.Request) {
	graph := types.Graph{}
	json.NewDecoder(req.Body).Decode(&graph)
	if _, exists := c.graphs[graph.Name]; !exists {
		nextID := strconv.Itoa(len(c.graphs))
		c.graphNameIDMap[graph.Name] = nextID
		c.graphs[graph.Name] = &graph
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"message\":\"Successfully created the " + graph.Name + " graph and all it's relations.\"}"))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"exception\":\"Error 2016: Duplicate graph entry\"}"))
	}
}

func (c *graphTestClient) createRelation(w http.ResponseWriter, req *http.Request, graphName string) {
	relation := types.Relation{}
	json.NewDecoder(req.Body).Decode(&relation)
	if graph, exists := c.graphs[graphName]; exists {
		// Ensure the provided relation doesn't already exist
		// for the retrieved graph.
		if !relationExists(relation, graph.Relations) {
			// Add our new relation to the graph.
			graph.Relations = append(graph.Relations, &relation)
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{\"message\":\"Successfully added the new relation " + relation.Name + " to the " + graphName + " graph.\"}"))
		} else {
			// Now in the case the relation already exists responsd with an error.
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{\"exception\":\"Error 2016: Duplicate graph relation\"}"))
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"exception\":\"Error 2016: Graph does not exist\"}"))
	}
}

func relationExists(relation types.Relation, relations []*types.Relation) bool {
	i := 0
	found := false
	for i < len(relations) && !found {
		if relations[i].Name == relation.Name {
			found = true
		} else {
			i++
		}
	}
	return found
}

var _ = Suite(&GraphSuite{})

func (s *GraphSuite) SetUpSuite(c *C) {
	// Simply provide an empty set of connection parameters as our test HTTP client
	// doesn't care about the url, just the request body for testing the single
	// graph focused method.
	cli, err := NewClient(&types.ConnectionParams{}, newGraphTestHttpClient())
	if err != nil {
		c.Error("Failed to setup our client for testing.")
	}
	s.client = cli
}

func (s *GraphSuite) TestCreateGraph(c *C) {
	res := s.client.CreateGraph(&types.Graph{
		Name: "test",
		Relations: []*types.Relation{
			&types.Relation{
				Name: "has",
				From: []string{"user"},
				To:   []string{"role"},
			},
		},
	})
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.StatusCode, Equals, http.StatusCreated)
	c.Assert(res.Message, Equals, "Successfully created the test graph and all it's relations.")
	// Now ensure that we can't recreate the test graph again.
	res = s.client.CreateGraph(&types.Graph{
		Name:      "test",
		Relations: []*types.Relation{},
	})
	c.Assert(res.Err, Not(Equals), nil)
	c.Assert(res.StatusCode, Equals, http.StatusBadRequest)
	c.Assert(res.Message, Equals, "Error 2016: Duplicate graph entry")
}

func (s *GraphSuite) TestCreateRelation(c *C) {
	res := s.client.CreateRelation("test", &types.Relation{
		Name: "isableto",
		From: []string{"user"},
		To:   []string{"permission"},
	})
	c.Assert(res.Err, Equals, nil)
	c.Assert(res.StatusCode, Equals, http.StatusCreated)
	c.Assert(res.Message, Equals, "Successfully added the new relation isableto to the test graph.")
	// Now ensure we can't recreate the same relation on the same graph.
	res = s.client.CreateRelation("test", &types.Relation{
		Name: "isableto",
		From: []string{"user"},
		To:   []string{"permission"},
	})
	c.Assert(res.Err, Not(Equals), nil)
	c.Assert(res.StatusCode, Equals, http.StatusBadRequest)
	c.Assert(res.Message, Equals, "Error 2016: Duplicate graph relation")
}
