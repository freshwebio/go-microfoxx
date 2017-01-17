package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/freshwebio/go-microfoxx/types"
)

const (
	mountEndpoint = "/microfoxx"
	loginEndpoint = "/login"
)

var (
	// ErrBadRequest is the error used when the server returns a 400 response.
	ErrBadRequest = errors.New("The body of the request didn't meet the expected requirements")
	// ErrGeneral is the error to be used for any other 4xx or 5xx response status codes
	// than a 400 bad request response.
	ErrGeneral = errors.New("Something went wrong in the process of the request")
	// ErrNotFound is the error returned in a response when no results could be found
	// for a query to the database service.
	ErrNotFound = errors.New("No results were found for the provided query")
)

// Client provides the base definition for all the functionality provided
// to interact with the Juntos service layer sitting on top of an ArangoDB database.
type Client interface {
	Refresh() error
	GetDocs(coll string, params *types.DocumentRetrievalParams) *types.DocumentsResult
	CreateDoc(coll string, doc interface{}) *types.DocumentOpResult
	GetDocCount(coll string, params *types.DocumentRetrievalParams) *types.DocumentCountResult
	RemoveDoc(coll string, key string) *types.DocumentOpResult
	GetDoc(coll string, key string) *types.DocumentResult
	UpdateDoc(coll string, key string, doc interface{}) *types.DocumentOpResult
	CursorQuery(params *types.CursorQueryParams) *types.CursorQueryResult
	CursorGetNextBatch(cursorID string) *types.CursorQueryResult
	InsertQuery(params *types.ModifyingQueryParams) *types.DocumentsOpResult
	UpdateQuery(params *types.ModifyingQueryParams) *types.DocumentsOpResult
	RemoveQuery(params *types.ModifyingQueryParams) *types.DocumentsOpResult
	CreateColl(name string) *types.CreationResult
	CreateGraph(graph *types.Graph) *types.CreationResult
	CreateRelation(graph string, relation *types.Relation) *types.CreationResult
	GetIndexes(coll string) *types.IndexListResult
	RemoveIndex(handle string) *types.IndexOpResult
	CreateIndex(params *types.IndexParams) *types.IndexOpResult
}

// WebClient provides a basis for the http client functionality
// utilised by a Juntos Arango client implementation to make HTTP requests.
type WebClient interface {
	Do(req *http.Request) (*http.Response, error)
	Post(url string, bodyType string, body io.Reader) (*http.Response, error)
}

type clientImpl struct {
	httpClient       WebClient
	connectionParams *types.ConnectionParams
	endpoint         string
	sessionInfo      *types.SessionInfo
}

// NewClient deals with creating a new client setup with the provided connection
// Result to be used on every request to the Juntos service for the provided database.
// Sessions are kept alive as long as they are being accessed, a session expires 5 minutes after
// the last time the session was accessed.
// It is up to the user to initialise a new session by calling a client's Refresh() method.
func NewClient(cParams *types.ConnectionParams, httpClient ...WebClient) (Client, error) {
	cli := &clientImpl{}
	// In the case httpClient isn't provided then use standard http.Client with a 10 second timeout.
	if len(httpClient) > 0 {
		// Only ever grab the first item as we only care for a single client.
		cli.httpClient = httpClient[0]
	} else {
		cli.httpClient = &http.Client{
			Timeout: time.Second * 10,
		}
	}
	cli.connectionParams = cParams
	// Set the defualt protocol scheme to http and the default host to localhost
	// and the default port to 80.
	if cParams.Scheme == "" {
		cParams.Scheme = "http"
	}
	if cParams.Host == "" {
		cParams.Host = "localhost"
	}
	if cParams.Port == "" {
		cParams.Port = "80"
	}
	cli.endpoint = cParams.Scheme + "://" + cParams.Host + ":" + cParams.Port + "/_db/" + cParams.Database + mountEndpoint
	// Now deal with setting up the session for the client.
	sessionInfo, err := cli.newSession()
	if err == nil {
		cli.sessionInfo = sessionInfo
	}
	return cli, err
}

// Refresh deals with creating a new session and updating the client's current
// session accordingly.
func (c *clientImpl) Refresh() error {
	sessionInfo, err := c.newSession()
	if err == nil {
		c.sessionInfo = sessionInfo
	}
	return err
}

// Deals with creating a new session by logging into the server.
func (c *clientImpl) newSession() (*types.SessionInfo, error) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{Username: c.connectionParams.Username, Password: c.connectionParams.Password})
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Post(c.endpoint+loginEndpoint, "application/json; charset=utf-8", b)
	if err != nil {
		return nil, err
	}
	var sessionInfo types.SessionInfo
	err = json.NewDecoder(resp.Body).Decode(&sessionInfo)
	return &sessionInfo, err
}

// Deals with attaching the session header to authenticate with the server on each request.
func (c *clientImpl) prepareRequest(method string, path string, qParams url.Values, body io.Reader) *http.Request {
	// Extract the query string from the path.
	req, _ := http.NewRequest(method, c.endpoint+path, body)
	if qParams != nil && len(qParams) > 0 {
		req.URL.RawQuery = qParams.Encode()
	}
	req.Header.Add("X-Session-Id", c.sessionInfo.SID)
	return req
}

func prepareExceptionResponse(resp *http.Response) (message string, err error) {
	var intermediary = struct {
		Message    string `json:"exception,omitempty"`
		ErrMessage string `json:"errorMessage,omitempty"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&intermediary)
	if err == nil {
		if intermediary.Message != "" {
			message = intermediary.Message
		} else if intermediary.ErrMessage != "" {
			message = intermediary.ErrMessage
		}
		switch resp.StatusCode {
		case http.StatusBadRequest:
			err = ErrBadRequest
		case http.StatusNotFound:
			err = ErrNotFound
		default:
			err = ErrGeneral
		}
	}
	return message, err
}

func prepareExceptionResponseFromMap(statusCode int, respMap map[string]interface{}) (message string, err error) {
	message = respMap["exception"].(string)
	switch statusCode {
	case http.StatusBadRequest:
		err = ErrBadRequest
	case http.StatusNotFound:
		err = ErrNotFound
	default:
		err = ErrGeneral
	}
	return message, err
}
