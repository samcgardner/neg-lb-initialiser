package lib

import (
	"errors"
	"fmt"
	"log"

	compute "google.golang.org/api/compute/v0.beta"
)

func EnsureURLMap(c Config, s *compute.Service, be *compute.BackendService) (*compute.UrlMap, error) {
	res, err := s.UrlMaps.List(c.Project).Do()
	if err != nil {
		fmt.Printf("Error reading URL maps %v\n", err)
		return nil, err
	}

	for _, item := range res.Items {
		if item.Name == c.Service+"-um" {
			return item, nil
		}
	}

	log.Printf("URL map not present, creating")
	log.Printf("At present, the API for this is broken. You can use the following console command instead")
	fmt.Printf(`gcloud compute url-maps create %s-um \
--default-service %s-be \
--project=%s`, c.Service, c.Service, c.Project)
	fmt.Printf("\n")

	return nil, errors.New("Giving up as it is not possible to create a URL map via the API at present")
}

func EnsureTargetProxy(c Config, s *compute.Service, cert *compute.SslCertificate, um *compute.UrlMap) (*compute.TargetHttpsProxy, error) {
	res, err := s.TargetHttpsProxies.List(c.Project).Do()
	if err != nil {
		fmt.Printf("Error reading target proxies %v\n", err)
		return nil, err
	}

	for _, item := range res.Items {
		if item.Name == c.Service+"-tp" {
			return item, nil
		}
	}

	log.Printf("Target proxy not present, creating")

	tp := compute.TargetHttpsProxy{
		Name:            fmt.Sprintf("%s-tp", c.Service),
		SslCertificates: []string{cert.SelfLink},
		UrlMap:          um.SelfLink,
	}

	_, err = s.TargetHttpsProxies.Insert(c.Project, &tp).Do()
	if err != nil {
		log.Printf("Error creating target proxy %v\n", err)
		return nil, err
	}

	return &tp, nil
}

func EnsureGlobalIP(c Config, s *compute.Service) (*compute.Address, error) {
	res, err := s.GlobalAddresses.List(c.Project).Do()
	if err != nil {
		fmt.Printf("Error reading global addresses %v\n", err)
		return nil, err
	}

	for _, item := range res.Items {
		if item.Name == c.Service+"-ip" {
			return item, nil
		}
	}

	log.Printf("Global address not present, creating")

	addr := compute.Address{
		Name:      fmt.Sprintf("%s-ip", c.Service),
		IpVersion: "IPV4",
	}

	_, err = s.GlobalAddresses.Insert(c.Project, &addr).Do()
	if err != nil {
		log.Printf("Error creating global address %v\n", err)
		return nil, err
	}

	return &addr, nil
}

func EnsureForwardingRule(c Config, s *compute.Service, ip *compute.Address, tp *compute.TargetHttpsProxy) (*compute.ForwardingRule, error) {
	res, err := s.GlobalForwardingRules.List(c.Project).Do()
	if err != nil {
		fmt.Printf("Error reading forwarding rules %v\n", err)
		return nil, err
	}

	for _, item := range res.Items {
		if item.Name == c.Service+"-fw" {
			return item, nil
		}
	}

	log.Printf("Forwarding rule not present, creating")

	fw := compute.ForwardingRule{
		Name:        fmt.Sprintf("%s-fw", c.Service),
		IPAddress:   ip.Address,
		IPProtocol:  "TCP",
		NetworkTier: "PREMIUM",
		PortRange:   "443",
		Target:      tp.SelfLink,
	}

	_, err = s.GlobalForwardingRules.Insert(c.Project, &fw).Do()
	if err != nil {
		log.Printf("Error creating forwarding rule %v\n", err)
		return nil, err
	}

	return &fw, nil
}
