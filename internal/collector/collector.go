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

func CollectSubdomains(domain string) []SubdomainRecord {
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	wg.Add(2)
	resultChan := make(chan []SubdomainRecord, 2)

	go CollectAnubis(ctx, &wg, resultChan, domain)
	go CollectAlienVault(ctx, &wg, resultChan, domain)
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	results := make([]SubdomainRecord, 0)
	for result := range resultChan {
		fmt.Printf("\nBLOCK:\n %+v\n", result)
		results = append(results, result...)
	}
	return results
}
