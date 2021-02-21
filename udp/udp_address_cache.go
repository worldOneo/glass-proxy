package udp

import (
	"net"
	"sync"
	"time"
)

// Cache is a sync.RWMutex
// Cache has an internal map which stores the items.
// The map at least RLocks itself on every operation
// Therefore taking the Lock will block every access.
type Cache struct {
	sync.RWMutex
	internalMap map[string]*Item
	holdingTime time.Duration
}

// Item an item holding a value and its last acces time.
// The access time is used to determin if an item should be removed.
type Item struct {
	value  interface{}
	access time.Time
}

// NewCache creates a new cache.
// Items are removed after the given expiration time.
// The map is cleaned at least every expireAfter time.
// The map might be cleaned without delay.
func NewCache(expireAfter time.Duration) *Cache {
	cache := &Cache{
		internalMap: make(map[string]*Item),
		holdingTime: expireAfter,
	}
	cache.expirationCheck()
	return cache
}

// Put puts the ip and its corresponding item into the cache
func (U *Cache) Put(ip *net.UDPAddr, host interface{}) {
	U.Lock()
	defer U.Unlock()
	U.internalMap[ip.String()] = &Item{
		value:  host,
		access: time.Now(),
	}
}

// Remove forces the ip out of the cache
func (U *Cache) Remove(ip *net.UDPAddr) {
	U.Lock()
	defer U.Unlock()
	delete(U.internalMap, ip.String())
}

// Get returns the stored item coresbonding to a clients IP
func (U *Cache) Get(ip *net.UDPAddr) interface{} {
	U.RLock()
	defer U.RUnlock()
	host := U.internalMap[ip.String()]
	if host == nil {
		return nil
	}
	U.internalMap[ip.String()].access = time.Now()
	return U.internalMap[ip.String()].value
}

func (U *Cache) expirationCheck() {
	U.Lock()
	defer U.Unlock()

	now := time.Now()
	shortest := U.holdingTime
	for key, item := range U.internalMap {
		access := item.access
		diff := now.Sub(access)

		if diff >= U.holdingTime {
			delete(U.internalMap, key)
			continue
		}

		if U.holdingTime-diff < shortest {
			shortest = U.holdingTime - diff
		}
	}

	time.AfterFunc(shortest, func() {
		go U.expirationCheck()
	})
}
