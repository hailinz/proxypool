package getter

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"encoding/json"

	"github.com/Sansui233/proxypool/log"
	"github.com/Sansui233/proxypool/pkg/proxy"
	"github.com/Sansui233/proxypool/pkg/tool"
)

// Add key value pair to creatorMap(string → creator) in base.go
func init() {
	Register("subscribe-nocode", NewSubscribeNocode)
}

// SubscribeNocode is A Getter with an additional property
type SubscribeNocode struct {
	Url string
}

// Get() of SubscribeNocode is to implement Getter interface
func (s *SubscribeNocode) Get() proxy.ProxyList {
	resp, err := tool.GetHttpClient().Get(s.Url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var nodesString = string(body)
	nodesString = strings.ReplaceAll(nodesString, "\t", "")

	nodes := strings.Split(nodesString, "\n")

	var proxylist proxy.ProxyList
	tempMap := map[string]byte{}
	tempMapLen := len(tempMap)

	for i, pstr := range nodes {
		if i == 0 || len(pstr) < 2 {
			continue
		}
		tempMap[pstr] = 0
		if len(tempMap) == tempMapLen {
			continue
		}
		tempMapLen++
		pstr = pstr[2:]
		if pp, ok := convert2Proxy(pstr); ok {
			if i == 1 && pp.BaseInfo().Name == "NULL" {
				return proxylist
			}
			proxylist = append(proxylist, pp)
		}
	}
	return proxylist
}

// Get2Chan() of SubscribeNocode is to implement Getter interface. It gets proxies and send proxy to channel one by one
func (s *SubscribeNocode) Get2ChanWG(pc chan proxy.Proxy, wg *sync.WaitGroup) {
	defer wg.Done()
	nodes := s.Get()
	log.Infoln("STATISTIC: SubscribeNocode\tcount=%d\turl=%s\n", len(nodes), s.Url)
	for _, node := range nodes {
		pc <- node
	}
}

func (s *SubscribeNocode) Get2Chan(pc chan proxy.Proxy) {
	nodes := s.Get()
	log.Infoln("STATISTIC: SubscribeNocode\tcount=%d\turl=%s\n", len(nodes), s.Url)
	for _, node := range nodes {
		pc <- node
	}
}

func NewSubscribeNocode(options tool.Options) (getter Getter, err error) {
	urlInterface, found := options["url"]
	if found {
		url, err := AssertTypeStringNotNull(urlInterface)
		if err != nil {
			return nil, err
		}
		return &SubscribeNocode{
			Url: url,
		}, nil
	}
	return nil, ErrorUrlNotFound
}

// Convert json string(clash format) to proxy
func convert2Proxy(pjson string) (proxy.Proxy, bool) {
	var f interface{}
	err := json.Unmarshal([]byte(pjson), &f)
	if err != nil {
		return nil, false
	}
	jsnMap := f.(interface{}).(map[string]interface{})

	switch jsnMap["type"].(string) {
	case "ss":
		var p proxy.Shadowsocks
		err := json.Unmarshal([]byte(pjson), &p)
		if err != nil {
			return nil, false
		}
		return &p, true
	case "ssr":
		var p proxy.ShadowsocksR
		err := json.Unmarshal([]byte(pjson), &p)
		if err != nil {
			return nil, false
		}
		return &p, true
	case "vmess":
		var p proxy.Vmess
		err := json.Unmarshal([]byte(pjson), &p)
		if err != nil {
			return nil, false
		}
		return &p, true
	case "trojan":
		var p proxy.Trojan
		err := json.Unmarshal([]byte(pjson), &p)
		if err != nil {
			return nil, false
		}
		return &p, true
	}
	return nil, false
}
