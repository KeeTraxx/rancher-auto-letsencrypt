package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/keetraxx/rancher-autocrypt/letsencrypt"
	"github.com/keetraxx/rancher-autocrypt/rancher"
)

func main() {
	log.SetOutput(os.Stdout)

	log.SetPrefix("[rancher-autocrypt] ")
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	directory, err := letsencrypt.NewDirectory(
		getEnv("LETSENCRYPT_DIRECTORY_URL", "https://acme-v01.api.letsencrypt.org/directory"),
		getEnv("LETSENCRYPT_CONTACT_EMAIL", "kt@compile.ch"),
	)

	if err != nil {
		log.Panicln(err)
	}

	cattleAuth, exists := os.LookupEnv("CATTLE_AGENT_INSTANCE_AUTH")

	if !exists {
		log.Println("Environment variable CATTLE_AGENT_INSTANCE_AUTH not found!")
		log.Panicln("Please make sure to set labels io.rancher.container.create_agent = true and io.rancher.container.agent.role = environment")
	}

	cattleURL, exists := os.LookupEnv("CATTLE_URL")

	if !exists {
		log.Println("Environment variable CATTLE_URL not found!")
		log.Panicln("Please make sure to set labels io.rancher.container.create_agent = true and io.rancher.container.agent.role = environment")
	}

	r := rancher.NewRancherEnvironment(cattleURL, cattleAuth)

	for {

		loadbalancers, err := r.GetLoadbalancers()
		services, err := r.GetRelevantServices()

		if err != nil {
			panic(err)
		}

		if len(services) == 0 {
			log.Println("No services with label ch.compile.letsencrypt found... Doing nothing...")
		}

		lbUpdate := false

		for _, service := range services {
			log.Printf("Service %v/%v has label ch.compile.letsencrypt\n", service.StackID, service.Name)

			var hostnames []string
			for _, portrule := range service.LbConfig.PortRules {
				if portrule.Hostname != "" {
					hostnames = append(hostnames, portrule.Hostname)
				}
			}

			log.Printf("Checking certificate for hostnames: %+v\n", hostnames)

			cert, _ := r.GetCertificate(hostnames[0])

			if cert != nil {
				block, _ := pem.Decode([]byte(cert.Cert))
				crt, err := x509.ParseCertificate(block.Bytes)

				if err != nil {
					panic(err)
				}

				log.Printf("Certificate %s expires on: %+v\n", crt.Subject.CommonName, crt.NotAfter)

				// Refresh Certificate 14 days before expiration
				if time.Now().Add(time.Hour * 24 * 14).After(crt.NotAfter) {
					log.Printf("Renewal for %s needed!\n", crt.Subject.CommonName)
					newkey, newcert, err := directory.GetCert(hostnames...)
					if err != nil {
						log.Println(err)
						continue
					}

					saveCert(hostnames[0], newkey, newcert)

					cert.Cert = string(newcert)
					cert.Key = string(newkey)
					cert.Description = fmt.Sprintf("Updated by rancher-autocrypt on %s", time.Now())

					err = r.UpsertCertificate(cert)

					if err != nil {
						log.Println(err)
					}

				} else {
					log.Println("Renewal not needed...")
				}
			} else {
				log.Println("new cert")

				newkey, newcert, err := directory.GetCert(hostnames...)
				if err != nil {
					log.Println(err)
					continue
				}

				saveCert(hostnames[0], newkey, newcert)

				cert := &rancher.Certificate{
					Cert:        string(newcert),
					Description: fmt.Sprintf("Updated by rancher-autocrypt on %s", time.Now()),
					Key:         string(newkey),
					Name:        hostnames[0],
				}

				err = r.UpsertCertificate(cert)

				cert, _ = r.GetCertificate(hostnames[0])

				lbUpdate = true

				for _, l := range loadbalancers {
					l.LbConfig.CertificateIDs = append(l.LbConfig.CertificateIDs, *cert.ID)
				}

				if err != nil {
					log.Println(err)
				}
			}

			if lbUpdate {
				for _, l := range loadbalancers {
					log.Printf("Updating loadbalancer %s configuration\n", l.Name)
					err = r.UpdateLoadbalancer(l)
					if err != nil {
						log.Println(err)
					}
				}

			}

		}

		log.Println("[rancher-autocrypt] Sleeping for 24 hours...")
		time.Sleep(time.Hour * 24)

	}
}

func saveCert(commonName string, key []byte, cert []byte) {
	dir := getEnv("CERTIFICATE_PATH", "/var/rancher-autocrypt")
	os.MkdirAll(dir, 0755)

	file, err := os.OpenFile(fmt.Sprintf("%s/%s", dir, commonName+".key"), os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		log.Println(err)
	}

	file.Write(key)

	log.Printf("Saving key for %s in %s", commonName, file.Name())

	file, err = os.OpenFile(fmt.Sprintf("%s/%s", dir, commonName+".crt"), os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		log.Println(err)
	}

	log.Printf("Saving certificate for %s in %s", commonName, file.Name())

	file.Write(cert)

}

func getEnv(key string, fallback string) string {
	if res, exists := os.LookupEnv(key); exists {
		return res
	}
	return fallback
}
