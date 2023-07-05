package collector

import (
	"context"
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
	var serviceResponse AlienVaultResponse
	err := doRequestTimes(ctx, serviceQueryUrl, &serviceResponse, 5)
	if err != nil {
		resultChan <- MakeErrorBlockResult("alienvault", domain, startedAt, err)
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

	resultChan <- MakeSuccessBlockResult("alienvault", domain, startedAt, subdomains)
}
