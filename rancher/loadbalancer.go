package rancher

type LoadbalancerservicesResponse struct {
	Type         string `json:"type"`
	ResourceType string `json:"resourceType"`
	Links        struct {
		Self string `json:"self"`
	} `json:"links"`
	Actions struct {
	} `json:"actions"`
	Data []*Loadbalancer `json:"data"`
}

type Loadbalancer struct {
	ID   string `json:"id"`
	Type string `json:"type"`

	BaseType  string `json:"baseType"`
	Name      string `json:"name"`
	State     string `json:"state"`
	AccountID string `json:"accountId"`

	LbConfig struct {
		Type                 string        `json:"type"`
		CertificateIDs       []string      `json:"certificateIds"`
		DefaultCertificateID string        `json:"defaultCertificateId"`
		PortRules            []interface{} `json:"portRules"`
	} `json:"lbConfig"`
}
