package cache

import (
	"container/list"
	"sync"
	"time"

	"github.com/import-yuefeng/smartDNS/core/common"
	"github.com/miekg/dns"
)

// Elem hold an answer and additional section that returned from the cache.
// The signature is put in answer, extra is empty there. This wastes some memory.
type elem struct {
	// time added + TTL, after this the elem is invalid
	expiration time.Time
	msg        *dns.Msg
	fastMap    *common.FastMap
}

// Cache is a cache that holds on the a number of RRs or DNS messages. The cache
// eviction is randomized.
type Cache struct {
	sync.RWMutex

	capacity    int
	domain      map[string]*elem
	head        *list.List
	cacheUpdate *CacheUpdate
	dnsBunch    *[]string
}

type CacheManager struct {
	c        *Cache
	interval string
	detector string
}

type Task struct {
	msg *dns.Msg
}

type CacheUpdate struct {
	TaskChan chan *Task
}
