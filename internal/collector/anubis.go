package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

func CollectAnubis(ctx context.Context, wg *sync.WaitGroup, resultChan chan []SubdomainRecord, domain string) {
	defer wg.Done()

	subdomains := make([]SubdomainRecord, 0)
	subdomainsSet := make(map[string]bool)
	serviceUrl := fmt.Sprintf("https://jonlu.ca/anubis/subdomains/%s", domain)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", serviceUrl, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("User-Agent", "Golang_Spider_Bot/3.0")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var serviceResponse []string
	err = json.Unmarshal(content, &serviceResponse)
	if err != nil {
		return
	}
	for _, record := range serviceResponse {
		subdomainsSet[record] = true
	}
	for subdomain := range subdomainsSet {
		subdomains = append(subdomains, SubdomainRecord{Subdomain: subdomain})
	}
	resultChan <- subdomains
}
