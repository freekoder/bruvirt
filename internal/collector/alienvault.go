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

type AlienVaultRecord struct {
	Address       string `json:"address"`
	First         string `json:"first"`
	Last          string `json:"last"`
	Hostname      string `json:"hostname"`
	RecordType    string `json:"record_type"`
	IndicatorLink string `json:"indicator_link"`
	FlagUrl       string `json:"flag_url"`
	FlagTitle     string `json:"flag_title"`
	AssetType     string `json:"asset_type"`
	ASN           string `json:"asn"`
}

type AlienVaultResponse struct {
	PassiveDNS []AlienVaultRecord `json:"passive_dns"`
}

func CollectAlienVault(ctx context.Context, wg *sync.WaitGroup, resultChan chan []SubdomainRecord, domain string) {
	defer wg.Done()

	subdomains := make([]SubdomainRecord, 0)
	subdomainsSet := make(map[string]bool)
	serviceUrl := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain)

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
	var serviceResponse AlienVaultResponse
	err = json.Unmarshal(content, &serviceResponse)
	if err != nil {
		return
	}
	for _, record := range serviceResponse.PassiveDNS {
		subdomainsSet[record.Hostname] = true
	}
	for subdomain := range subdomainsSet {
		subdomains = append(subdomains, SubdomainRecord{Subdomain: subdomain})
	}
	resultChan <- subdomains
}
