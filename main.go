package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"

	lib "github.com/samcgardner/neg-lb-initialiser/lib"
	compute "google.golang.org/api/compute/v0.beta"
	dns "google.golang.org/api/dns/v1"
)

// Config is the struct used to parse a .json file containing script config

func main() {
	c := parseConfig()

	ctx := context.Background()
	s, err := compute.NewService(ctx)
	if err != nil {
		log.Fatalf("Failed to create compute service due to error %v\n", err)
	}

	dnsService, err := dns.NewService(ctx)
	if err != nil {
		log.Fatalf("Failed to create DNS service due to error %v\n", err)
	}

	hc, err := lib.EnsureHealthCheck(c, s)
	if err != nil {
		log.Fatalf("Failed to ensure health check due to error %v\n", err)
	}

	be, err := lib.EnsureBackendServices(c, s, hc)
	if err != nil {
		log.Fatalf("Failed to ensure health check due to error %v\n", err)
	}

	err = lib.RegisterNEGs(c, s, be)
	if err != nil {
		log.Fatalf("Failed to register NEGs due to error %v\n", err)
	}

	cert, err := lib.EnsureCert(c, s)
	if err != nil {
		log.Fatalf("Failed to ensure cert due to error %v\n", err)
	}

	URLMap, err := lib.EnsureURLMap(c, s, be)
	if err != nil {
		log.Fatalf("Failed to ensure URL map due to error %v\n", err)
	}

	tp, err := lib.EnsureTargetProxy(c, s, cert, URLMap)
	if err != nil {
		log.Fatalf("Failed to ensure target proxy due to error %v\n", err)
	}

	ip, err := lib.EnsureGlobalIP(c, s)
	if err != nil {
		log.Fatalf("Failed to ensure global IP due to error %v\n", err)
	}

	_, err = lib.EnsureForwardingRule(c, s, ip, tp)
	if err != nil {
		log.Fatalf("Failed to ensure forwarding rule due to error %v\n", err)
	}

	err = lib.EnsureDNSEntry(c, dnsService, ip)
	if err != nil {
		log.Fatalf("Failed to ensure DNS entry due to error %v\n", err)
	}
}

func parseConfig() lib.Config {
	var c lib.Config
	bytes, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatalf("Failed to read config.json due to err %v", err)
	}
	if err := json.Unmarshal(bytes, &c); err != nil {
		log.Fatalf("Failed to parse config from config.json due to err %v", err)
	}

	if c.Project == "" {
		log.Fatalf("Project cannot be blank")
	}
	if c.Service == "" {
		log.Fatalf("Service cannot be blank")
	}
	if c.HCRequestPath == "" {
		log.Fatalf("Request path cannot be blank")
	}
	if c.DNSEntry == "" {
		log.Fatalf("DNS entry cannot be blank")
	}

	return c
}
