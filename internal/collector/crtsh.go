package collector

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type CertRecord struct {
	CommonName string `json:"common_name"`
	NameValue  string `json:"name_value"`
}

func CollectCrtSh(ctx context.Context, wg *sync.WaitGroup, resultChan chan BlockResult, domain string) {
	defer wg.Done()

	startedAt := time.Now()

	serviceQueryUrl := fmt.Sprintf("https://crt.sh/?q=.%s&output=json", domain)
	var serviceResponse []CertRecord
	err := doRequestTimes(ctx, serviceQueryUrl, &serviceResponse, 5)
	if err != nil {
		resultChan <- MakeErrorBlockResult("crtsh", domain, startedAt, err)
		return
	}

	subdomainsSet := make(map[string]bool)
	for _, record := range serviceResponse {
		subdomain := record.CommonName
		if strings.HasPrefix(subdomain, "*.") {
			subdomain = string([]byte(subdomain[2:]))
		}
		subdomain = strings.ToLower(subdomain)
		if subdomain == domain || strings.HasSuffix(subdomain, "."+domain) {
			subdomainsSet[subdomain] = true
		}
		altNamesRecord := record.NameValue
		altNames := strings.Split(altNamesRecord, "\n")
		for _, altName := range altNames {
			if strings.HasPrefix(altName, "*.") {
				altName = string([]byte(altName[2:]))
			}
			altName = strings.ToLower(altName)
			if altName == domain || strings.HasSuffix(altName, "."+domain) {
				subdomainsSet[altName] = true
			}
		}
	}

	subdomains := make([]SubdomainRecord, 0)
	for subdomain := range subdomainsSet {
		subdomains = append(subdomains, SubdomainRecord{Subdomain: subdomain})
	}

	resultChan <- MakeSuccessBlockResult("crtsh", domain, startedAt, subdomains)
}
