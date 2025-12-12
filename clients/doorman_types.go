package clients

// doormanLoginReq represents the login request payload
type doormanLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// queryPayload represents the query request payload
type queryPayload struct {
	ClusterName  string `json:"clusterName"`
	InstanceName string `json:"instanceName"`
	Schema       string `json:"schema"`
	Query        string `json:"query"`
}

// queryResponse represents the query response
type queryResponse struct {
	Code   int `json:"code"`
	Result struct {
		Headers []string        `json:"headers"`
		Rows    [][]interface{} `json:"rows"`
	} `json:"result"`
}
