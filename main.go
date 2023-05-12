package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/jamesog/iptoasn"
	"github.com/joho/godotenv"
)

type NetInfoResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Listening bool     `json:"listening"`
		Listeners []string `json:"listeners"`
		NPeers    string   `json:"n_peers"`
		Peers     []struct {
			NodeInfo struct {
				ProtocolVersion struct {
					P2P   string `json:"p2p"`
					Block string `json:"block"`
					App   string `json:"app"`
				} `json:"protocol_version"`
				ID         string `json:"id"`
				ListenAddr string `json:"listen_addr"`
				Network    string `json:"network"`
				Version    string `json:"version"`
				Channels   string `json:"channels"`
				Moniker    string `json:"moniker"`
				Other      struct {
					TxIndex    string `json:"tx_index"`
					RPCAddress string `json:"rpc_address"`
				} `json:"other"`
			} `json:"node_info"`
			IsOutbound       bool `json:"is_outbound"`
			ConnectionStatus struct {
				Duration    string `json:"Duration"`
				SendMonitor struct {
					Start    time.Time `json:"Start"`
					Bytes    string    `json:"Bytes"`
					Samples  string    `json:"Samples"`
					InstRate string    `json:"InstRate"`
					CurRate  string    `json:"CurRate"`
					AvgRate  string    `json:"AvgRate"`
					PeakRate string    `json:"PeakRate"`
					BytesRem string    `json:"BytesRem"`
					Duration string    `json:"Duration"`
					Idle     string    `json:"Idle"`
					TimeRem  string    `json:"TimeRem"`
					Progress int       `json:"Progress"`
					Active   bool      `json:"Active"`
				} `json:"SendMonitor"`
				RecvMonitor struct {
					Start    time.Time `json:"Start"`
					Bytes    string    `json:"Bytes"`
					Samples  string    `json:"Samples"`
					InstRate string    `json:"InstRate"`
					CurRate  string    `json:"CurRate"`
					AvgRate  string    `json:"AvgRate"`
					PeakRate string    `json:"PeakRate"`
					BytesRem string    `json:"BytesRem"`
					Duration string    `json:"Duration"`
					Idle     string    `json:"Idle"`
					TimeRem  string    `json:"TimeRem"`
					Progress int       `json:"Progress"`
					Active   bool      `json:"Active"`
				} `json:"RecvMonitor"`
				Channels []struct {
					ID                int    `json:"ID"`
					SendQueueCapacity string `json:"SendQueueCapacity"`
					SendQueueSize     string `json:"SendQueueSize"`
					Priority          string `json:"Priority"`
					RecentlySent      string `json:"RecentlySent"`
				} `json:"Channels"`
			} `json:"connection_status"`
			RemoteIP string `json:"remote_ip"`
		} `json:"peers"`
	} `json:"result"`
}

type ASNInfo struct {
	ASN    uint32
	ASName string
	Count  int
}

func getASName(ipList []iptoasn.IP, asn uint32) string {
	for _, ip := range ipList {
		if ip.ASNum == asn {
			return ip.ASName
		}
	}
	return "Unknown"
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file, using default RPC URL")
	}

	rpcURL := os.Getenv("RPC_URL") + "/net_info"
	if rpcURL == "" {
		rpcURL = "http://localhost:26657/net_info"
	}

	response, err := QueryRPC(rpcURL)
	if err != nil {
		fmt.Println("Error querying RPC endpoint:", err)
		return
	}

	var ipList []iptoasn.IP
	fmt.Println("Number of Peers:", response.Result.NPeers)
	for _, peer := range response.Result.Peers {
		asn, err := queryASN(peer.RemoteIP)
		if err != nil {
			fmt.Println(err)
		} else {
			ipList = append(ipList, asn)
		}
	}
	ipASNCount := make(map[uint32]int)
	for _, ip := range ipList {
		ipASNCount[ip.ASNum]++
	}

	asnInfoList := make([]ASNInfo, 0, len(ipASNCount))

	for asn, count := range ipASNCount {
		asnInfo := ASNInfo{
			ASN:    asn,
			ASName: getASName(ipList, asn),
			Count:  count,
		}
		asnInfoList = append(asnInfoList, asnInfo)
	}

	sort.Slice(asnInfoList, func(i, j int) bool {
		return asnInfoList[i].Count > asnInfoList[j].Count
	})

	for _, asnInfo := range asnInfoList {
		fmt.Printf("ASN: %d, ASName: %s, Count: %d\n", asnInfo.ASN, asnInfo.ASName, asnInfo.Count)
	}
	GenerateBarChart(asnInfoList)
}

func queryASN(peer_ip string) (iptoasn.IP, error) {
	asn, err := iptoasn.LookupIP(peer_ip)
	return asn, err
}

func QueryRPC(url string) (*NetInfoResponse, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result NetInfoResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func GenerateBarChart(asnInfoList []ASNInfo) {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	var ASNCount []float64
	var ASName []string

	for _, asn := range asnInfoList {
		ASNCount = append(ASNCount, float64(asn.Count))
		ASName = append(ASName, safesubstr(asn.ASName))
		fmt.Printf("ASN: %d, ASName: %s, Count: %d\n", asn.ASN, asn.ASName, asn.Count)
	}

	bc := widgets.NewBarChart()
	bc.Data = ASNCount
	bc.Labels = ASName
	bc.Title = "Peers per ASN Chart"
	bc.SetRect(5, 25, 175, 50)
	bc.BarWidth = 10

	ui.Render(bc)

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		}
	}
}

func safesubstr(s string) string {
	if len(s) > 10 {
		return s[:7] + "..."
	}
	return s
}
