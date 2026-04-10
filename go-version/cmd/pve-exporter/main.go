package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"pve-summary-exporter/pkg/cache"
	"pve-summary-exporter/pkg/config"
	"pve-summary-exporter/pkg/metrics"
	"pve-summary-exporter/pkg/proxmox"
)

var (
	memCache      = cache.NewCache()
	proxmoxClient = proxmox.NewProxmoxClient()
	pveMetrics    = metrics.NewMetrics()
	configPath    = getEnv("CONFIG_PATH", "./config/proxmox-node.yml")
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func callAllRequest(node config.NodeConfig) metrics.ScrapeResult {
	startTime := time.Now()
	log.Printf("Reading Info: identify=%s, host=%s", node.Identify, node.Host)

	keyCache := fmt.Sprintf("%s-auth", node.Identify)
	var creds *proxmox.Credentials

	cachedCreds, found := memCache.Get(keyCache)
	if found {
		creds = cachedCreds.(*proxmox.Credentials)
	} else {
		var err error
		creds, err = proxmoxClient.GetCredentials(node.Host, node.Ticket)
		if err != nil {
			log.Printf("Error getting credentials for %s: %v", node.Identify, err)
			return metrics.ScrapeResult{
				Identify: node.Identify,
				Host:     node.Host,
				Success:  false,
				TimeEnd:  time.Since(startTime).Milliseconds(),
			}
		}
		memCache.Set(keyCache, creds)
	}

	clusterInfo, isNeedReAuth, err := proxmoxClient.GetClusterStatus(node.Host, creds.Ticket, creds.CSRFToken)
	if isNeedReAuth {
		memCache.Del(keyCache)
		return metrics.ScrapeResult{
			Identify: node.Identify,
			Host:     node.Host,
			Success:  false,
			TimeEnd:  time.Since(startTime).Milliseconds(),
		}
	}
	if err != nil {
		log.Printf("Error getting cluster status for %s: %v", node.Identify, err)
		return metrics.ScrapeResult{
			Identify: node.Identify,
			Host:     node.Host,
			Success:  false,
			TimeEnd:  time.Since(startTime).Milliseconds(),
		}
	}

	if clusterInfo.Now == nil {
		log.Printf("Could not find current node in cluster status for %s", node.Identify)
		return metrics.ScrapeResult{
			Identify: node.Identify,
			Host:     node.Host,
			Success:  false,
			TimeEnd:  time.Since(startTime).Milliseconds(),
		}
	}

	nodeName := clusterInfo.Now.Name

	var rrdData map[string]float64
	var nodeStatus map[string]interface{}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		var err error
		rrdData, _, err = proxmoxClient.GetClusterRRDData(node.Host, creds.Ticket, creds.CSRFToken, nodeName)
		if err != nil {
			log.Printf("Error getting RRD data for %s: %v", node.Identify, err)
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		nodeStatus, _, err = proxmoxClient.GetNodeStatus(node.Host, creds.Ticket, creds.CSRFToken, nodeName)
		if err != nil {
			log.Printf("Error getting node status for %s: %v", node.Identify, err)
		}
	}()

	wg.Wait()

	return metrics.ScrapeResult{
		Identify: node.Identify,
		Host:     node.Host,
		Success:  true,
		TimeEnd:  time.Since(startTime).Milliseconds(),
		Node:     clusterInfo,
		RRDData:  rrdData,
		Status:   nodeStatus,
	}
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading config: %v", err), http.StatusInternalServerError)
		return
	}

	var results []metrics.ScrapeResult
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, node := range cfg.Monitoring {
		wg.Add(1)
		go func(n config.NodeConfig) {
			defer wg.Done()
			res := callAllRequest(n)
			mu.Lock()
			results = append(results, res)
			mu.Unlock()
		}(node)
	}

	wg.Wait()

	pveMetrics.BuildMetrics(results)

	promhttp.HandlerFor(pveMetrics.Registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}

func main() {
	http.HandleFunc("/metrics", metricsHandler)

	port := ":8007"
	log.Printf("Server running on port %s\nURL: http://localhost%s/metrics", port, port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
