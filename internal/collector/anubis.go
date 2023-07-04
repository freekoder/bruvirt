package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

func CollectAnubis(ctx context.Context, wg *sync.WaitGroup, resultChan chan BlockResult, domain string) {
	defer wg.Done()

	startedAt := time.Now()

	serviceQueryUrl := fmt.Sprintf("https://jonlu.ca/anubis/subdomains/%s", domain)
	content, err := runTimes(ctx, doServiceRequest, serviceQueryUrl, 5)
	if err != nil {
		resultChan <- BlockResult{
			Name:       "anubis",
			Domain:     domain,
			StartedAt:  startedAt,
			EndedAt:    time.Now(),
			Error:      err,
			Subdomains: make([]SubdomainRecord, 0),
		}
		return
	}

	var serviceResponse []string
	err = json.Unmarshal(content, &serviceResponse)
	if err != nil {
		resultChan <- BlockResult{
			Name:       "anubis",
			Domain:     domain,
			StartedAt:  startedAt,
			EndedAt:    time.Now(),
			Error:      err,
			Subdomains: make([]SubdomainRecord, 0),
		}
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
	endedAt := time.Now()

	resultChan <- BlockResult{
		Name:       "anubis",
		Domain:     domain,
		StartedAt:  startedAt,
		EndedAt:    endedAt,
		Subdomains: subdomains,
	}
}
