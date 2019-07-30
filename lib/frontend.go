package lib

import (
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

	/* For reasons thoroughly opaque to me, this (correct) API call fails. You can run the following
		command manually as a workaround

		gcloud compute url-maps create $SERVICE-um \                                                                                                                                                                                                                                                                         master
	    --default-service $SERVICE-be \
			--project=$PROJECT \
			--global
	*/
	URLMap := compute.UrlMap{
		Name:           fmt.Sprintf("%s-um", c.Service),
		DefaultService: be.SelfLink,
	}

	_, err = s.UrlMaps.Insert(c.Project, &URLMap).Do()
	if err != nil {
		log.Printf("Error creating URL map %v\n", err)
		return nil, err
	}

	return &URLMap, nil
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
