package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/freekoder/bruvirt/internal/collector"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("bruvirt v0.0.1")
	domain := "palantir.com"

	collectSubdomains(domain)

	//ips := readIPS(domain)
	//vhosts := readVhosts(domain)

	//for idx, ip := range ips {
	//	fmt.Printf("checking: %d - %s\n", idx, ip)
	//	checkIP(ip, vhosts, domain)
	//}
	os.Exit(0)
}

func collectSubdomains(domain string) error {
	_ = collector.CollectSubdomains(domain)
	//fmt.Printf("subdomains: %v", subdomains)
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

func readIPS(domain string) []string {
	file, err := os.Open(fmt.Sprintf("tests/%s/ips", domain))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = file.Close()
	}()

	ips := make([]string, 0)
	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		ips = append(ips, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return ips
}

func readVhosts(domain string) []string {
	file, err := os.Open(fmt.Sprintf("tests/%s/vhosts", domain))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = file.Close()
	}()

	vhosts := make([]string, 0)
	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		vhosts = append(vhosts, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return vhosts
}
