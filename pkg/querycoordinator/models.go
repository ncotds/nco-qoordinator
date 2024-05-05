package querycoordinator

// Credentials represents a user credentials
type Credentials struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

// Query represents a SQL query
type Query struct {
	SQL string `json:"sql"`
}

// QueryResult represents a sql query result:
//   - RowSet useful for queries that fetch data (usually 'select')
//   - AffectedRows is the number of rows affected by queries that write data (e.g. 'insert')
type QueryResult struct {
	RowSet       []QueryResultRow `json:"rowset"`
	AffectedRows int              `json:"affected_rows"`
	Error        error            `json:"error"`
}

// QueryResultRow represents a single table row as a <column DSName>: <column value> map
type QueryResultRow map[string]any
