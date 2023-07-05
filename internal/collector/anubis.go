package collector

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func CollectAnubis(ctx context.Context, wg *sync.WaitGroup, resultChan chan BlockResult, domain string) {
	defer wg.Done()

	startedAt := time.Now()

	serviceQueryUrl := fmt.Sprintf("https://jonlu.ca/anubis/subdomains/%s", domain)
	var serviceResponse []string
	err := doRequestTimes(ctx, serviceQueryUrl, &serviceResponse, 5)
	if err != nil {
		resultChan <- MakeErrorBlockResult("anubis", domain, startedAt, err)
		return
	}

	subdomainsSet := make(map[string]bool)
	for _, record := range serviceResponse {
		subdomainsSet[record] = true
	}

	subdomains := make([]SubdomainRecord, 0)
	for subdomain := range subdomainsSet {
		subdomains = append(subdomains, SubdomainRecord{Subdomain: subdomain})
	}

	resultChan <- MakeSuccessBlockResult("anubis", domain, startedAt, subdomains)
}
