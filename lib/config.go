package lib

// Config is a struct to hold program configuration
type Config struct {
	Project       string `json:"project"`
	Service       string `json:"service"`
	HCRequestPath string `json:"healthcheck_request_path"`
	DNSEntry      string `json:"dns_entry"`
}
