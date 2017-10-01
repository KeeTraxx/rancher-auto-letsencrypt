package letsencrypt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"golang.org/x/crypto/acme"
)

type Directory struct {
	url     string
	contact string
}

func NewDirectory(url string, contactEmail string) (*Directory, error) {

	// TODO: Improve validation for URL
	if len(url) < 6 {
		return nil, fmt.Errorf("URL %v invalid", url)
	}

	// TODO: Improve validation for email
	if len(contactEmail) < 4 {
		return nil, fmt.Errorf("contactEmail %v invalid", url)
	}

	return &Directory{
		url:     url,
		contact: contactEmail,
	}, nil
}

func (d *Directory) GetCert(domains ...string) (pemkey []byte, pemcert []byte, err error) {
	log.Printf("Generating registration key...\n")
	regkey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		return nil, nil, err
	}

	acmeClient := &acme.Client{
		Key:          regkey,
		DirectoryURL: d.url,
	}

	log.Printf("Using directory: %v\n", acmeClient.DirectoryURL)
	mail := fmt.Sprintf("mailto:%v", d.contact)
	log.Println(mail)
	uc := acme.Account{
		Contact: []string{mail},
	}

	log.Printf("Registering new ACME account: %v\n", uc.Contact)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if _, err := acmeClient.Register(ctx, &uc, acme.AcceptTOS); err != nil {
		return nil, nil, err
	}

	req := &x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: domains[0]},
		DNSNames: domains,
	}

	log.Printf("Signing key for: %+v\n", req.Subject.CommonName)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	log.Printf("Certificate Signing Request (CSR) for: %+v\n", req.Subject)

	csr, err := x509.CreateCertificateRequest(rand.Reader, req, key)
	if err != nil {
		return nil, nil, err
	}

	for _, domain := range domains {
		log.Printf("Authz for: %v\n", domain)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		if err := authz(ctx, acmeClient, domain); err != nil {
			cancel()
			return nil, nil, err
		}
		cancel()
	}

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	cert, curl, err := acmeClient.CreateCert(ctx, csr, 365*12*time.Hour, true)
	if err != nil {
		return nil, nil, err
	}

	log.Printf("cert url: %s", curl)

	pemkeyblock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	pemkey = pem.EncodeToMemory(pemkeyblock)

	for _, b := range cert {
		b = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: b})
		pemcert = append(pemcert, b...)
	}

	return pemkey, pemcert, nil
}

func authz(ctx context.Context, client *acme.Client, domain string) error {
	z, err := client.Authorize(ctx, domain)
	if err != nil {
		return err
	}
	if z.Status == acme.StatusValid {
		return nil
	}
	var chal *acme.Challenge
	for _, c := range z.Challenges {
		chal = c
		if c.Type == "http-01" {
			break
		}
	}
	if chal == nil {
		return errors.New("no supported challenge found")
	}

	// respond to http-01 challenge
	ln, err := net.Listen("tcp", ":5002")
	if err != nil {
		return fmt.Errorf("listen %s: %v", ":5002", err)
	}
	defer ln.Close()

	val, err := client.HTTP01ChallengeResponse(chal.Token)
	if err != nil {
		return err
	}
	path := client.HTTP01ChallengePath(chal.Token)
	go http.Serve(ln, http01Handler(path, val))

	if _, err := client.Accept(ctx, chal); err != nil {
		return fmt.Errorf("accept challenge: %v", err)
	}
	_, err = client.WaitAuthorization(ctx, z.URI)
	return err
}

func http01Handler(path, value string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			log.Printf("unknown request path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write([]byte(value))
	})
}
