package udp

import (
	"net"
	"sync"
	"time"
)

type Cache struct {
	sync.RWMutex
	internalMap map[string]*Item
	holdingTime time.Duration
}

type Item struct {
	value  interface{}
	access time.Time
}

func NewCache(expireAfter time.Duration) *Cache {
	cache := &Cache{
		internalMap: make(map[string]*Item),
		holdingTime: expireAfter,
	}
	cache.expirationCheck()
	return cache
}

func (U *Cache) Put(ip *net.UDPAddr, host interface{}) {
	U.Lock()
	defer U.Unlock()
	U.internalMap[ip.String()] = &Item{
		value:  host,
		access: time.Now(),
	}
}

func (U *Cache) Remove(ip *net.UDPAddr) {
	U.Lock()
	defer U.Unlock()
	delete(U.internalMap, ip.String())
}

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
