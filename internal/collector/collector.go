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

type SourceList []string

type Result struct {
	Domain     string
	Subdomains map[string]SourceList
}

func (r *Result) Add(subdomain, source string) {
	if _, ok := r.Subdomains[subdomain]; !ok {
		r.Subdomains[subdomain] = make(SourceList, 0)
	}
	r.Subdomains[subdomain] = append(r.Subdomains[subdomain], source)
}

func CollectSubdomains(domain string) Result {
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wg.Add(3)
	resultChan := make(chan BlockResult, 3)

	go CollectAnubis(ctx, &wg, resultChan, domain)
	go CollectAlienVault(ctx, &wg, resultChan, domain)
	go CollectCrtSh(ctx, &wg, resultChan, domain)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	collectorResult := Result{Domain: domain, Subdomains: make(map[string]SourceList)}
	for result := range resultChan {
		fmt.Printf("%s - %s - %d\n", result.Name, result.Duration(), len(result.Subdomains))
		for _, subdomain := range result.Subdomains {
			collectorResult.Add(subdomain.Subdomain, result.Name)
		}
	}
	return collectorResult
}
