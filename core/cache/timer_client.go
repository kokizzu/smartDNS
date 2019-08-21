// The MIT License (MIT)
// Copyright (c) 2019 import-yuefeng
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package cache

import (
	"fmt"
	"time"

	"github.com/import-yuefeng/smartDNS/core/common"
	"github.com/import-yuefeng/smartDNS/core/detector/ping"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

func (worker *CacheUpdate) AddTask(expiration uint32, msg *dns.Msg, fastMap *common.FastMap, DNSBunch *[]string) {

	newTimer := time.NewTimer(time.Second * time.Duration(expiration))
	go func() {
		<-newTimer.C
		// TaskDetector detector.Detector

		TaskDetector := ping.NewDetector(msg, fastMap, DNSBunch)
		TaskDetector.Detect()
		fmt.Printf("task: %s , time: %d expired\n", msg.Answer, expiration)
		// fmt.Printf("time: %d expired\n", expiration)

	}()

}

func (worker *CacheUpdate) Handle() {
	for {
		select {
		case msg, ok := <-worker.TaskChan:
			if !ok {
				break
			}
			log.Infof("Program %d: %v \n", time.Now().Unix(), msg.msg.Answer)
		}
	}

}

func main() {
	var updateService = new(CacheUpdate)
	updateService.TaskChan = make(chan *Task, 1000)
	DNSBunch := make([]string, 10)

	for i := 0; i < 6; i++ {
		msg := new(dns.Msg)
		fastMap := new(common.FastMap)
		updateService.AddTask(uint32(i), msg, fastMap, &DNSBunch)
	}

	updateService.Handle()
}
