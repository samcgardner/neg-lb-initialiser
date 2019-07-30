package lib

import (
	"fmt"
	"log"
	"strings"

	compute "google.golang.org/api/compute/v0.beta"
)

// EnsureHealthCheck creates a health check for the service if one does not already exist and then returns
// the service's health check
func EnsureHealthCheck(c Config, s *compute.Service) (*compute.HealthCheck, error) {
	res, err := s.HealthChecks.List(c.Project).Do()
	if err != nil {
		log.Printf("Error reading healthchecks %v\n", err)
		return nil, err
	}

	for _, item := range res.Items {
		if item.Name == c.Service+"-hc" {
			return item, nil
		}
	}

	log.Printf("Health check not present, creating")

	HTTPHc := compute.HTTPHealthCheck{
		RequestPath: c.HCRequestPath,
		Port:        8080,
	}

	hc := compute.HealthCheck{
		Name:            fmt.Sprintf("%s-hc", c.Service),
		Type:            "HTTP",
		HttpHealthCheck: &HTTPHc,
	}

	_, err = s.HealthChecks.Insert(c.Project, &hc).Do()
	if err != nil {
		log.Printf("Error creating healthcheck %v\n", err)
		return nil, err
	}

	return &hc, nil
}

func EnsureBackendServices(c Config, s *compute.Service, hc *compute.HealthCheck) (*compute.BackendService, error) {
	res, err := s.BackendServices.List(c.Project).Do()
	if err != nil {
		log.Printf("Error reading backend services %v\n", err)
		return nil, err
	}

	for _, item := range res.Items {
		if item.Name == c.Service+"-be" {
			return item, nil
		}
	}

	log.Printf("Backend service not present, creating")

	be := compute.BackendService{
		HealthChecks: []string{hc.SelfLink},
		Name:         fmt.Sprintf("%s-be", c.Service),
		Port:         8080,
	}

	_, err = s.BackendServices.Insert(c.Project, &be).Do()
	if err != nil {
		log.Printf("Error creating backend service %v\n", err)
		return nil, err
	}

	return &be, nil
}

func RegisterNEGs(c Config, s *compute.Service, be *compute.BackendService) error {
	res, err := s.NetworkEndpointGroups.AggregatedList(c.Project).Do()
	if err != nil {
		log.Printf("Error reading NEGs %v\n", err)
		return err
	}

	log.Printf("Registering NEGs")

	var NEGs []*compute.NetworkEndpointGroup
	for _, agg := range res.Items {
		for _, neg := range agg.NetworkEndpointGroups {
			NEGs = append(NEGs, neg)
		}
	}

	var backends []*compute.Backend
	testString := fmt.Sprintf("%s-80", c.Service)
	for _, neg := range NEGs {
		if !strings.Contains(neg.Name, testString) {
			continue
		}
		backend := fromNEG(neg)
		backends = append(backends, backend)
	}

	be.Backends = backends
	_, err = s.BackendServices.Update(c.Project, be.Name, be).Do()
	if err != nil {
		log.Printf("Error registering NEGs %v\n", err)
		return err
	}

	return nil
}

func fromNEG(neg *compute.NetworkEndpointGroup) *compute.Backend {
	return &compute.Backend{
		BalancingMode: "RATE",
		Group:         neg.SelfLink,
		MaxRate:       9999,
	}
}
