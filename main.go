package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Data struct {
	AddrInfo              AddrInfo               `json:"AddrInfo"`
	LastAdvertisement     map[string]string      `json:"LastAdvertisement"`
	LastAdvertisementTime string                 `json:"LastAdvertisementTime"`
	Publisher             AddrInfo               `json:"Publisher"`
	ExtendedProviders     map[string]interface{} `json:"ExtendedProviders"`
	FrozenAt              *string                `json:"FrozenAt"`
}

type AddrInfo struct {
	ID    string   `json:"ID"`
	Addrs []string `json:"Addrs"`
}

func main() {
	url := "https://cid.contact/providers"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading data:", err)
		return
	}

	var data []Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	uniqueAddrs := make(map[string]struct{})
	uniqueTCPAddrs := make(map[string]struct{})
	uniqueWSAddrs := make(map[string]struct{})
	uniqueHTTPSAddrs := make(map[string]struct{})
	uniqueIDs := make(map[string]struct{})
	frozenCount := 0
	adsWithin6Months := 0

	sixMonthsAgo := time.Now().AddDate(0, -6, 0)

	for _, entry := range data {
		for _, addr := range entry.AddrInfo.Addrs {
			uniqueAddrs[addr] = struct{}{}
			if strings.Contains(addr, "tcp") && !strings.Contains(addr, "https") && !strings.Contains(addr, "wss") {
				uniqueTCPAddrs[addr] = struct{}{}
			}
			if strings.Contains(addr, "ws") && !strings.Contains(addr, "https") {
				uniqueWSAddrs[addr] = struct{}{}
			}
			if strings.Contains(addr, "https") || strings.Contains(addr, "wss") {
				uniqueHTTPSAddrs[addr] = struct{}{}
			}
		}
		uniqueIDs[entry.AddrInfo.ID] = struct{}{}

		if entry.FrozenAt != nil {
			frozenCount++
		}

		adTime, err := time.Parse(time.RFC3339, entry.LastAdvertisementTime)
		if err == nil && adTime.After(sixMonthsAgo) {
			adsWithin6Months++
		}
	}

	fmt.Println("Number of unique provider Addrs:", len(uniqueAddrs))
	fmt.Println("Number of unique TCP provider Addrs:", len(uniqueTCPAddrs))
	fmt.Println("Number of unique WS provider Addrs:", len(uniqueWSAddrs))
	fmt.Println("Number of unique HTTPS provider Addrs:", len(uniqueHTTPSAddrs))
	fmt.Println("Number of unique provider IDs:", len(uniqueIDs))
	fmt.Println("Number of FrozenAt entries:", frozenCount)
	fmt.Println("Number of advertisements within the last 6 months:", adsWithin6Months)
}
