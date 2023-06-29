package collector

type SubdomainRecord struct {
	Subdomain string
}

func CollectSubdomains(domain string) []SubdomainRecord {
	return CollectAlienVault(domain)
}
