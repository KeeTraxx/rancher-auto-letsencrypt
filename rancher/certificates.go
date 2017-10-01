package rancher

type CertificatesResponse struct {
	Type         string `json:"type"`
	ResourceType string `json:"resourceType"`
	Links        struct {
		Self string `json:"self"`
	} `json:"links"`
	CreateTypes struct {
	} `json:"createTypes"`
	Actions struct {
	} `json:"actions"`
	Data      []Certificate `json:"data"`
	SortLinks struct {
		AccountID   string `json:"accountId"`
		Cert        string `json:"cert"`
		CertChain   string `json:"certChain"`
		Created     string `json:"created"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Key         string `json:"key"`
		Kind        string `json:"kind"`
		Name        string `json:"name"`
		RemoveTime  string `json:"removeTime"`
		Removed     string `json:"removed"`
		State       string `json:"state"`
		UUID        string `json:"uuid"`
	} `json:"sortLinks"`
	Pagination struct {
		First    interface{} `json:"first"`
		Previous interface{} `json:"previous"`
		Next     interface{} `json:"next"`
		Limit    int         `json:"limit"`
		Total    interface{} `json:"total"`
		Partial  bool        `json:"partial"`
	} `json:"pagination"`
	Sort    interface{} `json:"sort"`
	Filters struct {
		AccountID   interface{} `json:"accountId"`
		Cert        interface{} `json:"cert"`
		CertChain   interface{} `json:"certChain"`
		Created     interface{} `json:"created"`
		Description interface{} `json:"description"`
		ID          interface{} `json:"id"`
		Key         interface{} `json:"key"`
		Kind        interface{} `json:"kind"`
		Name        []struct {
			Value    string `json:"value"`
			Modifier string `json:"modifier"`
		} `json:"name"`
		RemoveTime interface{} `json:"removeTime"`
		Removed    interface{} `json:"removed"`
		State      interface{} `json:"state"`
		UUID       interface{} `json:"uuid"`
	} `json:"filters"`
	CreateDefaults struct {
	} `json:"createDefaults"`
}

type Certificate struct {
	ID          *string     `json:"id,omitempty"`
	Name        string      `json:"name"`
	Cert        string      `json:"cert"`
	CertChain   string      `json:"certChain"`
	Description interface{} `json:"description"`
	Key         interface{} `json:"key"`
}
