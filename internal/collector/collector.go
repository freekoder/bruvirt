package collector

import (
	"context"
	"encoding/json"
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

func (r BlockResult) ErrorString() string {
	if r.Error != nil {
		return r.Error.Error()
	} else {
		return ""
	}
}

type SourceList []string

type CollectorStat struct {
	Name     string
	Duration time.Duration
	Count    int
	Error    string
}

type Result struct {
	Domain     string
	Stats      []CollectorStat
	Subdomains map[string]SourceList
}

func (r *Result) AddBucket(bucket BlockResult) {
	r.Stats = append(r.Stats, CollectorStat{
		Name:     bucket.Name,
		Duration: bucket.Duration(),
		Count:    len(bucket.Subdomains),
		Error:    bucket.ErrorString(),
	})
	for _, subdomain := range bucket.Subdomains {
		r.Add(subdomain.Subdomain, bucket.Name)
	}
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

	collectorResult := Result{Domain: domain, Stats: make([]CollectorStat, 0), Subdomains: make(map[string]SourceList)}
	for result := range resultChan {
		//fmt.Printf("%s - %s - %d - (%v)\n", result.Name, result.Duration(), len(result.Subdomains), result.Error)
		collectorResult.AddBucket(result)
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

func doRequestTimes(ctx context.Context, url string, response interface{}, times int) error {
	var lastError error
	for i := 0; i < times; i++ {
		reqContent, err := doServiceRequest(ctx, url)
		if err != nil {
			lastError = err
			if err == context.DeadlineExceeded || err == context.Canceled {
				break
			}
			time.Sleep(500 * time.Millisecond)
			continue
		} else {
			err = json.Unmarshal(reqContent, &response)
			if err != nil {
				lastError = err
				time.Sleep(500 * time.Millisecond)
				continue
			}
			break
		}
	}
	return lastError
}

func MakeErrorBlockResult(name string, domain string, startedAt time.Time, err error) BlockResult {
	return BlockResult{
		Name:       name,
		Domain:     domain,
		StartedAt:  startedAt,
		EndedAt:    time.Now(),
		Error:      err,
		Subdomains: make([]SubdomainRecord, 0),
	}
}

func MakeSuccessBlockResult(name string, domain string, startedAt time.Time, subdomains []SubdomainRecord) BlockResult {
	return BlockResult{
		Name:       name,
		Domain:     domain,
		StartedAt:  startedAt,
		EndedAt:    time.Now(),
		Subdomains: subdomains,
	}
}
