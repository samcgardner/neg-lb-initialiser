package lib

import (
	"fmt"
	"log"

	compute "google.golang.org/api/compute/v0.beta"
	dns "google.golang.org/api/dns/v1"
)

func EnsureDNSEntry(c Config, s *dns.Service, ip *compute.Address) error {
	res, err := s.ResourceRecordSets.List(c.Project, "dns").Do()
	if err != nil {
		fmt.Printf("Error reading forwarding rules %v\n", err)
		return err
	}

	for _, item := range res.Rrsets {
		if item.Name == c.DNSEntry {
			return nil
		}
	}

	log.Printf("Creating DNS entry")

	change := dns.Change{
		Additions: []*dns.ResourceRecordSet{
			&dns.ResourceRecordSet{
				Name:    c.DNSEntry,
				Rrdatas: []string{ip.Address},
				Ttl:     300,
				Type:    "A",
			},
		},
	}

	_, err = s.Changes.Create(c.Project, "dns", &change).Do()
	if err != nil {
		fmt.Printf("Error creating DNS entry %v\n", err)
		return err
	}

	return nil
}
