package rancher

type ProjectsResponse struct {
	Type         string `json:"type"`
	ResourceType string `json:"resourceType"`
	Links        struct {
		Self string `json:"self"`
	} `json:"links"`
	Actions struct {
	} `json:"actions"`
	Data []Project `json:"data"`
}

type Project struct {
	ID   string `json:"id"`
	Type string `json:"type"`

	BaseType  string `json:"baseType"`
	Name      string `json:"name"`
	State     string `json:"state"`
	AccountID string `json:"accountId"`
}
