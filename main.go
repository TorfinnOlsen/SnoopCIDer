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
	frozenCount := 0
	adsWithin6Months := 0

	sixMonthsAgo := time.Now().AddDate(0, -6, 0)

	uniquePublisherIDs := make(map[string]struct{})

	for _, entry := range data {
		if _, exists := uniquePublisherIDs[entry.Publisher.ID]; !exists {
			uniquePublisherIDs[entry.Publisher.ID] = struct{}{}

			for _, publisherAddr := range entry.Publisher.Addrs {
				uniqueAddrs[publisherAddr] = struct{}{}
				if strings.Contains(publisherAddr, "tcp") && !strings.Contains(publisherAddr, "https") && !strings.Contains(publisherAddr, "wss") {
					uniqueTCPAddrs[publisherAddr] = struct{}{}
				}
				if strings.Contains(publisherAddr, "ws") && !strings.Contains(publisherAddr, "https") {
					uniqueWSAddrs[publisherAddr] = struct{}{}
				}
				if strings.Contains(publisherAddr, "https") || strings.Contains(publisherAddr, "wss") {
					for _, providerAddr := range entry.AddrInfo.Addrs {
						addrPair := fmt.Sprintf("%s : %s", publisherAddr, providerAddr)
						uniqueHTTPSAddrs[addrPair] = struct{}{}
					}
				}
			}
		}

		if entry.FrozenAt != nil {
			frozenCount++
		}

		adTime, err := time.Parse(time.RFC3339, entry.LastAdvertisementTime)
		if err == nil && adTime.After(sixMonthsAgo) {
			adsWithin6Months++
		}
	}

	// Printing results
	fmt.Println("Number of unique publisher Addrs:", len(uniqueAddrs))
	fmt.Println("Number of unique TCP publisher Addrs:", len(uniqueTCPAddrs))
	fmt.Println("Number of unique WS publisher Addrs:", len(uniqueWSAddrs))
	fmt.Println("Number of unique HTTPS publisher Addrs:", len(uniqueHTTPSAddrs))

	fmt.Println("HTTPS publisher addresses and their provider addresses:")
	for addrPair := range uniqueHTTPSAddrs {
		fmt.Printf("\t%s\n", addrPair)
	}

	fmt.Println("Number of unique publisher IDs:", len(uniquePublisherIDs))
	fmt.Println("Number of FrozenAt entries:", frozenCount)
	fmt.Println("Number of advertisements within the last 6 months:", adsWithin6Months)

}
