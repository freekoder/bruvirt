package collector

import (
	"context"
	"fmt"
	"io"
	"net/http"
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
	Error      error
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

func doServiceRequest(ctx context.Context, serviceUrl string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", serviceUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Golang_Spider_Bot/3.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func runTimes(ctx context.Context, requestFunc func(context.Context, string) ([]byte, error), url string, times int) ([]byte, error) {
	var content []byte
	var lastError error
	for i := 0; i < times; i++ {
		reqContent, err := requestFunc(ctx, url)
		if err != nil {
			lastError = err
			if err == context.DeadlineExceeded || err == context.Canceled {
				break
			}
			time.Sleep(500 * time.Millisecond)
			continue
		} else {
			content = reqContent
			break
		}
	}
	return content, lastError
}
