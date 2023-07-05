package main

import (
	"crypto/tls"
	"fmt"
	"github.com/freekoder/bruvirt/internal/collector"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("bruvirt v0.0.2\n")
	domain := "hackerone.com"

	collectSubdomains(domain)

	//for idx, ip := range ips {
	//	fmt.Printf("checking: %d - %s\n", idx, ip)
	//	checkIP(ip, vhosts, domain)
	//}
	os.Exit(0)
}

func collectSubdomains(domain string) error {
	result := collector.CollectSubdomains(domain)
	fmt.Printf("total subdomains: %d\n\n", len(result.Subdomains))
	for subdomain, sources := range result.Subdomains {
		fmt.Printf("%s %s\n", subdomain, sources)
	}
	return nil
}

func checkIP(ip string, vhosts []string, domain string) error {
	err := checkVhost(ip, "")
	if err != nil {
		return err
	}
	err = checkVhost(ip, fmt.Sprintf("notpresentvhost.%s", domain))
	if err != nil {
		return err
	}
	for _, vhost := range vhosts {
		err = checkVhost(ip, vhost)
	}
	return nil
}

func checkVhost(ip string, vhost string) error {
	hostUrl, err := url.Parse(fmt.Sprintf("https://%s", ip))
	if err != nil {
		return err
	}
	req := http.Request{URL: hostUrl}
	if vhost != "" {
		req.Host = vhost
	}
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   3 * time.Second,
	}
	resp, err := client.Do(&req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	content := string(body)
	lines := strings.Split(content, "\n")

	statusCode := resp.StatusCode
	fmt.Printf("response %s-%s:\t\tsc:%d size:%d lines:%d\n", ip, vhost, statusCode, len(body), len(lines))
	return nil
}
