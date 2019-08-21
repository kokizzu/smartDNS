// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package ping

import (
	"container/list"
	"strings"

	"github.com/import-yuefeng/smartDNS/core/common"
	"github.com/import-yuefeng/smartDNS/core/outbound"
	"github.com/import-yuefeng/smartDNS/core/outbound/clients"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/sparrc/go-ping"
)

type Pinger struct {
	fastMap *common.FastMap
	bundle  *outbound.Bundle
}

type BundleMsg struct {
	msg        *dns.Msg
	bundleName string
}

func NewDetector(msg *dns.Msg, fastMap *common.FastMap, DNSBunch map[string][]*common.DNSUpstream) *Pinger {
	bundle := new(outbound.Bundle)

	for name, value := range DNSBunch {
		bundle.ClientBundle[name] = clients.NewClientBundle(msg, value, "127.0.0.1", 0, nil, name, nil)
	}
	return &Pinger{fastMap, bundle}
}

func (data *Pinger) Sort(fastTable *list.List) (fastMap *common.FastMap) {
	return fastTable.Front().Value.(*common.FastMap)
}

func (data *Pinger) Detect() (fastTable *list.List) {
	fastTable = list.New()
	dnsBundle := data.bundle
	domainToIP := make(map[string]string)
	var ch chan *BundleMsg

	for name, o := range dnsBundle.ClientBundle {
		go func(c *clients.RemoteClientBundle, ch chan *BundleMsg, bundleName string) {
			result := c.Exchange(true)
			ch <- &BundleMsg{result.ResponseMessage, bundleName}
			return
		}(o, ch, name)
	}
	for i := 0; i < len(dnsBundle.ClientBundle); i++ {
		c := <-ch
		if c != nil {
			dnsRespon := strings.Fields(c.msg.Answer[0].String())
			domainToIP[c.bundleName] = dnsRespon[4]
		}
	}
	var bundle chan string
	for name, ip := range domainToIP {
		go func(ip string, name string, bundle chan string) {
			pinger, err := ping.NewPinger(ip)
			if err != nil {
				log.Error(err)
				panic(err)
			}
			pinger.Count = 1
			pinger.Run()
			stat := pinger.Statistics()
			log.Info(stat)
			bundle <- name
			return
		}(ip, name, bundle)
	}

	for i := 0; i < len(dnsBundle.ClientBundle); i++ {
		if c := <-ch; c != nil {
			fastMap := new(common.FastMap)
			fastMap.Domain, fastMap.DnsBundle = data.fastMap.Domain, c.bundleName
			fastTable.PushBack(fastMap)
		}
		if i >= int(float64(len(dnsBundle.ClientBundle))*0.6) {
			break
		}
	}
	return fastTable
}
