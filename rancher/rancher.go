package rancher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type RancherEnvironment struct {
	endpoint      string
	authorization string
	client        *http.Client
}

func NewRancherEnvironment(endpoint string, authorization string) *RancherEnvironment {
	client := &http.Client{}
	return &RancherEnvironment{
		endpoint,
		authorization,
		client,
	}
}

func (rancher *RancherEnvironment) GetRelevantServices() (relevantServices []Service, err error) {
	url := fmt.Sprintf("%v/services?system=false&kind=service", rancher.endpoint)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", rancher.authorization)

	var serviceResponse ServiceResponse

	res, err := rancher.client.Do(req)
	body, err := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &serviceResponse)

	if err != nil {
		return nil, err
	}

	for _, service := range serviceResponse.Data {
		_, exists := service.LaunchConfig.Labels["ch.compile.letsencrypt"]

		if exists && service.State == "active" {
			relevantServices = append(relevantServices, service)
		}
	}

	return relevantServices, nil
}

func (rancher *RancherEnvironment) GetLoadbalancers() (relevantServices []*Loadbalancer, err error) {
	// http: //rancher.tran-engineering.ch:8080/v2-beta/loadbalancerservices
	url := fmt.Sprintf("%v/loadbalancerservices", rancher.endpoint)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", rancher.authorization)

	var loadbalancerservicesResponse LoadbalancerservicesResponse

	res, err := rancher.client.Do(req)
	body, err := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &loadbalancerservicesResponse)

	if err != nil {
		return nil, err
	}

	return loadbalancerservicesResponse.Data, nil
}

func (rancher *RancherEnvironment) GetCertificate(hostname string) (*Certificate, error) {
	url := fmt.Sprintf("%s/certificates?system=false&name=%s", rancher.endpoint, hostname)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", rancher.authorization)

	var certificatesResponse CertificatesResponse

	res, err := rancher.client.Do(req)
	body, err := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &certificatesResponse)

	if err != nil {
		return nil, err
	}

	if len(certificatesResponse.Data) > 0 {
		return &certificatesResponse.Data[0], nil
	}

	return nil, nil

}

func (rancher *RancherEnvironment) UpsertCertificate(cert *Certificate) error {
	var url string
	var method string

	project, err := rancher.GetProject()

	if err != nil {
		return err
	}

	if cert.ID != nil {
		url = fmt.Sprintf("%s/projects/%s/certificates/%s", rancher.endpoint, project.ID, *cert.ID)
		method = "PUT"
	} else {
		url = fmt.Sprintf("%s/projects/%s/certificates", rancher.endpoint, project.ID)
		method = "POST"
	}

	jsondata, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsondata))

	req.Header.Add("Authorization", rancher.authorization)
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return err
	}

	log.Println("Updating certificate on rancher...")
	//log.Printf("%+v", req)

	resp, err := rancher.client.Do(req)

	if resp.StatusCode != 200 {
		return fmt.Errorf("Server returned %d %s", resp.StatusCode, resp.Status)
	}

	return nil

}

func (rancher *RancherEnvironment) UpdateLoadbalancer(lb *Loadbalancer) error {
	url := fmt.Sprintf("%s/loadbalancerservices/%s", rancher.endpoint, lb.ID)

	jsondata, err := json.Marshal(lb)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsondata))

	req.Header.Add("Authorization", rancher.authorization)
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return err
	}

	log.Printf("%+v\n\n", req)

	resp, err := rancher.client.Do(req)

	if err != nil {
		return err
	}

	log.Printf("%+v\n\n", resp)

	return nil

}

func (rancher *RancherEnvironment) GetProject() (*Project, error) {
	url := fmt.Sprintf("%s/projects", rancher.endpoint)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", rancher.authorization)

	var projectsResponse ProjectsResponse

	res, err := rancher.client.Do(req)
	body, err := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &projectsResponse)

	if err != nil {
		return nil, err
	}

	if len(projectsResponse.Data) > 0 {
		return &projectsResponse.Data[0], nil
	}

	return nil, nil
}
