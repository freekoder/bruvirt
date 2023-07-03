package collector

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type SubdomainRecord struct {
	Subdomain string
}

type BlockResult struct {
	Name       string
	Domain     string
	StartedAt  time.Time
	EndedAt    time.Time
	Subdomains []SubdomainRecord
}

func (r BlockResult) Duration() time.Duration {
	return r.EndedAt.Sub(r.StartedAt)
}

func CollectSubdomains(domain string) []SubdomainRecord {
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wg.Add(2)
	resultChan := make(chan BlockResult, 2)

	go CollectAnubis(ctx, &wg, resultChan, domain)
	go CollectAlienVault(ctx, &wg, resultChan, domain)
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	results := make([]SubdomainRecord, 0)
	for result := range resultChan {
		fmt.Printf("%s - %s - %d\n", result.Name, result.Duration(), len(result.Subdomains))
		results = append(results, result.Subdomains...)
	}
	return results
}
