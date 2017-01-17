package types

import "io"

// Graph provides the type for a definition of a graph
// to be sent to or stored in the ArangoDB data store exposed
// through the juntos foxx service.
type Graph struct {
	Name      string      `json:"name"`
	Relations []*Relation `json:"relations"`
	// Orphans are collections in the graph that have no relations.
	Orphans []string `json:"orphans"`
}

// Relation provides the data structure for a relation between n vertex collections
// to m vertex collections.
// For instance a company and a customer could buy electronics or groceries
// so there could be an edge relation has_bought that could be from [company, customer] to
// [groceries, electronics].
type Relation struct {
	Name string   `json:"name"`
	From []string `json:"from"`
	To   []string `json:"to"`
}

// CreationResult provides the response data relevant for attempted creation operations.
// Not for document
type CreationResult struct {
	Err        error
	CreatedIDs []string
	Message    string
	StatusCode int
}

// UpdateResult provides the response data relevant for attempted update operations.
// This is for events on collections, graphs and relations.
type UpdateResult struct {
	Err        error
	StatusCode int
	Message    string
	UpdatedIDs map[string][]string
}

// DeleteResult provides the response data relevant for attempted deletion operations.
type DeleteResult struct {
	Err          error
	StatusCode   int
	Message      string
	DeletedItems io.Reader
}

// DocumentOpResult provides the response data relevant for an attempted operation
// on a document in a collection.
type DocumentOpResult struct {
	Err        error
	StatusCode int
	Message    string
	Event      io.Reader
	Document   io.Reader
}

// DocumentsOpResult provides the response data relevant for an attempted operation
// on multiple documents in a collection through a modification AQL query.
type DocumentsOpResult struct {
	Err        error
	StatusCode int
	Message    string
	Events     io.Reader
	Documents  io.Reader
}

// DocumentsResult provides the response data used when running queries to retrieve multiple documents.
type DocumentsResult struct {
	Err        error
	StatusCode int
	Message    string
	Documents  io.Reader
}

// DocumentCountResult provides the data structure which is use to encapsulate the http response
// when making request to retrieve document count in collections.
type DocumentCountResult struct {
	Err        error
	StatusCode int
	Message    string
	Count      int
}

// DocumentResult provides the response data used when retrieving a single document by it's unique
// identifier.
type DocumentResult struct {
	Err        error
	StatusCode int
	Message    string
	Document   io.Reader
}

// DocumentRetrievalParams are the parameters to be used to prepare a request to retrieve
// a document from the data store.
type DocumentRetrievalParams struct {
	Fields      map[string]string
	SortFields  []string
	SortOrder   string
	LimitOffset int
	LimitCount  int
}

// CursorQueryParams are the parameters to be used when making a request to start a new cursor
// for a provided AQL query to retrieve results in batches.
type CursorQueryParams struct {
	Query     string                 `json:"query"`
	BindVars  map[string]interface{} `json:"bindVars"`
	BatchSize int                    `json:"batchSize"`
	Count     bool                   `json:"count"`
}

// CursorQueryResult provides the response result for cursor queries.
type CursorQueryResult struct {
	Err        error
	StatusCode int
	Message    string
	Documents  io.Reader
	Cursor     string
	HasMore    bool
}

// ModifyingQueryParams are the parameters to be used when making a request
// to the modification query endpoints to carry out INSERT, UPDATE or REMOVE operations
// through AQL queries.
type ModifyingQueryParams struct {
	WriteCollection string
	ReadCollections []string
	Query           string
	BindVars        map[string]interface{}
}

// ConnectionParams are the parameters used when invoking sessions for clients.
type ConnectionParams struct {
	Host     string
	Port     string
	Database string
	Scheme   string
	Username string
	Password string
}

// SessionInfo is the data structure holding session information provided when logging into the service.
type SessionInfo struct {
	SID string `json:"sid"`
	UID string `json:"uid"`
}

// IndexListResult provides the data structure to be used
// to store a list of indexes for a collection returned from the foxx service.
type IndexListResult struct {
	Err        error
	StatusCode int
	Message    string
	Indexes    []*Index
}

// IndexParams is the data structure for holding
// the parameters to be used to create a new index.
type IndexParams struct {
	Collection string   `json:"collection"`
	Type       string   `json:"type"`
	Fields     []string `json:"fields"`
	Sparse     bool     `json:"sparse"`
	Unique     bool     `json:"unique"`
}

// Index provides the data structure for retrieving
// indexes from the arango data store.
type Index struct {
	ID                  string   `json:"id"`
	Type                string   `json:"type"`
	Fields              []string `json:"fields"`
	SelectivityEstimate int      `json:"selectivityEstimate"`
	Unique              bool     `json:"unique"`
	Sparse              bool     `json:"sparse"`
}

// IndexOpResult holds the result details for an index
// operation on the foxx service.
type IndexOpResult struct {
	Err        error
	StatusCode int
	Message    string
}
