package proxmox

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"pve-summary-exporter/pkg/config"
)

type ProxmoxClient struct {
	client *http.Client
}

func NewProxmoxClient() *ProxmoxClient {
	return &ProxmoxClient{
		client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

type Credentials struct {
	CSRFToken string `json:"csrftoken"`
	Ticket    string `json:"ticket"`
	IsError   bool   `json:"-"`
}

func (pc *ProxmoxClient) GetCredentials(host string, ticketCfg config.TicketConfig) (*Credentials, error) {
	apiUrl := fmt.Sprintf("%s/api2/extjs/access/ticket", host)

	data := url.Values{}
	data.Set("username", ticketCfg.Username)
	data.Set("password", ticketCfg.Password)
	data.Set("realm", ticketCfg.Realm)
	if ticketCfg.NewFormat != "" {
		data.Set("new-format", ticketCfg.NewFormat)
	}

	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux 64x; Linux; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")

	resp, err := pc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed: %s", resp.Status)
	}

	var result struct {
		Data struct {
			CSRFPreventionToken string `json:"CSRFPreventionToken"`
			Ticket              string `json:"ticket"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &Credentials{
		CSRFToken: result.Data.CSRFPreventionToken,
		Ticket:    result.Data.Ticket,
	}, nil
}

type ClusterStatusItem struct {
	IP     string `json:"ip"`
	Name   string `json:"name"`
	Online int    `json:"online"`
	Type   string `json:"type"`
	Level  string `json:"level"`
}

type ClusterStatusResponse struct {
	Now  *ClusterStatusItem
	List []ClusterStatusItem
}

func (pc *ProxmoxClient) GetClusterStatus(host, ticket, csrf string) (*ClusterStatusResponse, bool, error) {
	apiUrl := fmt.Sprintf("%s/api2/extjs/cluster/status", host)
	parsedHost, _ := url.Parse(host)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", ticket))
	req.Header.Set("CSRFPreventionToken", csrf)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux 64x; Linux; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")

	resp, err := pc.client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, true, fmt.Errorf("unauthorized")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("request failed: %s", resp.Status)
	}

	var result struct {
		Data []ClusterStatusItem `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, false, err
	}

	response := &ClusterStatusResponse{
		List: result.Data,
	}

	for i := range result.Data {
		if result.Data[i].IP == parsedHost.Hostname() {
			response.Now = &result.Data[i]
			break
		}
	}

	return response, false, nil
}

type RRDDataPoint struct {
	Time      int64   `json:"time"`
	NetOut    float64 `json:"netout"`
	NetIn     float64 `json:"netin"`
	IOWait    float64 `json:"iowait"`
	MaxCPU    float64 `json:"maxcpu"`
	CPU       float64 `json:"cpu"`
	MemTotal  float64 `json:"memtotal"`
	MemUsed   float64 `json:"memused"`
	SwapUsed  float64 `json:"swapused"`
	SwapTotal float64 `json:"swaptotal"`
	RootUsed  float64 `json:"rootused"`
	RootTotal float64 `json:"roottotal"`
	LoadAvg   float64 `json:"loadavg"`
}

func (pc *ProxmoxClient) GetClusterRRDData(host, ticket, csrf, nodeName string) (map[string]float64, bool, error) {
	apiUrl := fmt.Sprintf("%s/api2/json/nodes/%s/rrddata?timeframe=hour&cf=AVERAGE", host, nodeName)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", ticket))
	req.Header.Set("CSRFPreventionToken", csrf)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux 64x; Linux; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")

	resp, err := pc.client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, true, fmt.Errorf("unauthorized")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("request failed: %s", resp.Status)
	}

	var result struct {
		Data []RRDDataPoint `json:"data"`
	}

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, false, err
	}

	if len(result.Data) == 0 {
		return nil, false, nil
	}

	// Filter and find latest data point
	var latest *RRDDataPoint
	for i := range result.Data {
		if latest == nil || result.Data[i].Time > latest.Time {
			latest = &result.Data[i]
		}
	}

	if latest == nil {
		return nil, false, nil
	}

	data := map[string]float64{
		"memtotal":  latest.MemTotal,
		"memused":   latest.MemUsed,
		"memfree":   latest.MemTotal - latest.MemUsed,
		"swaptotal": latest.SwapTotal,
		"swapused":  latest.SwapUsed,
		"swapfree":  latest.SwapTotal - latest.SwapUsed,
		"rootused":  latest.RootUsed,
		"roottotal": latest.RootTotal,
		"rootfree":  latest.RootTotal - latest.RootUsed,
		"loadavg":   latest.LoadAvg,
		"cpu":       latest.CPU,
		"maxcpu":    latest.MaxCPU,
		"iowait":    latest.IOWait,
		"netin":     latest.NetIn,
		"netout":    latest.NetOut,
	}

	return data, false, nil
}

func (pc *ProxmoxClient) GetNodeStatus(host, ticket, csrf, nodeName string) (map[string]interface{}, bool, error) {
	apiUrl := fmt.Sprintf("%s/api2/json/nodes/%s/status", host, nodeName)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", ticket))
	req.Header.Set("CSRFPreventionToken", csrf)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux 64x; Linux; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")

	resp, err := pc.client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, true, fmt.Errorf("unauthorized")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("request failed: %s", resp.Status)
	}

	var result struct {
		Data struct {
			CPU     float64 `json:"cpu"`
			CPUInfo struct {
				Model   string `json:"model"`
				CPUs    int    `json:"cpus"`
				Cores   int    `json:"cores"`
				Sockets int    `json:"sockets"`
				UserHZ  int    `json:"user_hz"`
				MHz     string `json:"mhz"`
			} `json:"cpuinfo"`
			LoadAvg []float64 `json:"loadavg"`
			Swap    struct {
				Total float64 `json:"total"`
				Used  float64 `json:"used"`
				Free  float64 `json:"free"`
			} `json:"swap"`
			Memory struct {
				Total float64 `json:"total"`
				Used  float64 `json:"used"`
				Free  float64 `json:"free"`
			} `json:"memory"`
			RootFS struct {
				Total float64 `json:"total"`
				Avail float64 `json:"avail"`
				Used  float64 `json:"used"`
				Free  float64 `json:"free"`
			} `json:"rootfs"`
			PVEVersion string  `json:"pveversion"`
			Wait       float64 `json:"wait"`
			Uptime     int64   `json:"uptime"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, false, err
	}

	data := map[string]interface{}{
		"cpu":         result.Data.CPU,
		"model":       result.Data.CPUInfo.Model,
		"cpus":        result.Data.CPUInfo.CPUs,
		"cores":       result.Data.CPUInfo.Cores,
		"sockets":     result.Data.CPUInfo.Sockets,
		"user_hz":     result.Data.CPUInfo.UserHZ,
		"mhz":         result.Data.CPUInfo.MHz,
		"loadavg1":    0.0,
		"loadavg2":    0.0,
		"loadavg3":    0.0,
		"swaptotal":   result.Data.Swap.Total,
		"swapused":    result.Data.Swap.Used,
		"swapfree":    result.Data.Swap.Free,
		"memorytotal": result.Data.Memory.Total,
		"memoryused":  result.Data.Memory.Used,
		"memoryfree":  result.Data.Memory.Free,
		"rootfstotal": result.Data.RootFS.Total,
		"rootfsavail": result.Data.RootFS.Avail,
		"rootfsused":  result.Data.RootFS.Used,
		"rootfsfree":  result.Data.RootFS.Free,
		"pveversion":  result.Data.PVEVersion,
		"wait":        result.Data.Wait,
		"uptime":      result.Data.Uptime,
	}

	if len(result.Data.LoadAvg) >= 3 {
		data["loadavg1"] = result.Data.LoadAvg[0]
		data["loadavg2"] = result.Data.LoadAvg[1]
		data["loadavg3"] = result.Data.LoadAvg[2]
	}

	return data, false, nil
}
