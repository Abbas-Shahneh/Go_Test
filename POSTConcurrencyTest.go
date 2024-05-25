package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {

	ipAddresses := []string{
		"192.168.1.1",
		"192.168.1.2",
		"192.168.1.3",
		//"192.168.1.4",
		//"192.168.1.5",
		//"192.168.1.6",
		//"192.168.1.7",
		//"192.168.1.8",
		//"192.168.1.9",
		//"192.168.1.10",
	}

	var wg sync.WaitGroup
	wg.Add(len(ipAddresses))

	start := time.Now()

	for _, ip := range ipAddresses {
		go func(ip string) {
			defer wg.Done()
			url := "http://localhost:8080/scan/" + ip
			resp, err := http.Post(url, "application/json", nil)
			if err != nil {
				log.Printf("Failed to send POST request to %s: %v\n", ip, err)
				return
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusAccepted {
				log.Printf("Received non-accepted response for POST %s: %s\n", ip, resp.Status)
			} else {
				log.Printf("Successfully sent scan request for %s\n", ip)
			}
		}(ip)
	}

	wg.Wait()

	duration := time.Since(start)
	log.Printf("All POST requests completed in %s\n", duration)
}
