package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
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

func CollectAlienVault(ctx context.Context, wg *sync.WaitGroup, resultChan chan BlockResult, domain string) {
	defer wg.Done()

	startedAt := time.Now()

	serviceQueryUrl := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain)
	content, err := runTimes(ctx, doServiceRequest, serviceQueryUrl, 5)
	if err != nil {
		return
	}

	var serviceResponse AlienVaultResponse
	err = json.Unmarshal(content, &serviceResponse)
	if err != nil {
		return
	}

	subdomainsSet := make(map[string]bool)
	for _, record := range serviceResponse.PassiveDNS {
		subdomainsSet[record.Hostname] = true
	}

	subdomains := make([]SubdomainRecord, 0)
	for subdomain := range subdomainsSet {
		subdomains = append(subdomains, SubdomainRecord{Subdomain: subdomain})
	}
	endedAt := time.Now()

	resultChan <- BlockResult{
		Name:       "alienvault",
		Domain:     domain,
		StartedAt:  startedAt,
		EndedAt:    endedAt,
		Subdomains: subdomains,
	}
}
