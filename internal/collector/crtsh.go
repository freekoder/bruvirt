package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	subdomains := make([]SubdomainRecord, 0)
	subdomainsSet := make(map[string]bool)
	serviceUrl := fmt.Sprintf("https://crt.sh/?q=.%s&output=json", domain)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", serviceUrl, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "Golang_Spider_Bot/3.0")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var serviceResponse []CertRecord
	err = json.Unmarshal(content, &serviceResponse)
	if err != nil {
		return
	}

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

	for subdomain := range subdomainsSet {
		subdomains = append(subdomains, SubdomainRecord{Subdomain: subdomain})
	}
	endedAt := time.Now()

	resultChan <- BlockResult{
		Name:       "crtsh",
		Domain:     domain,
		StartedAt:  startedAt,
		EndedAt:    endedAt,
		Subdomains: subdomains,
	}
}
