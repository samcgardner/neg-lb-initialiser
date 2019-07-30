package lib

import (
	"fmt"
	"log"

	compute "google.golang.org/api/compute/v0.beta"
)

func EnsureCert(c Config, s *compute.Service) (*compute.SslCertificate, error) {
	res, err := s.SslCertificates.List(c.Project).Do()
	if err != nil {
		fmt.Printf("Error reading certificates %v\n", err)
		return nil, err
	}

	for _, item := range res.Items {
		if item.Name == c.Service+"-cert" {
			return item, nil
		}
	}

	log.Printf("Managed SSL certificate not present, creating")

	cert := compute.SslCertificate{
		Name: fmt.Sprintf("%s-cert", c.Service),
		Managed: &compute.SslCertificateManagedSslCertificate{
			Domains: []string{c.DNSEntry},
		},
		Type: "MANAGED",
	}

	_, err = s.SslCertificates.Insert(c.Project, &cert).Do()
	if err != nil {
		log.Printf("Error creating managed SSL certificate %v\n", err)
	}

	return &cert, nil
}
