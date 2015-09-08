package main

import (
	"bytes"
	"encoding/gob"

	"github.com/bradfitz/gomemcache/memcache"
)

type keyInfo string

var (
	mc = NewCacheStore()

	//overshadows memcache package errors
	ErrCacheMiss   = memcache.ErrCacheMiss
	ErrNoServers   = memcache.ErrNoServers
	ErrServerError = memcache.ErrServerError
)

const (
	CacheProductsKeyInfo    keyInfo = "cached_products"
	CacheDefaultDurationSec int32   = 10800 //3 hours = 3*60*60
)

func NewCacheStore() *memcache.Client {
	return memcache.New(config.MemcacheHostAddr)
}

func WriteToCache(key keyInfo, v interface{}) error {
	obj, err := encode(v)
	if err != nil {
		return err
	}
	item := &memcache.Item{
		Key:   string(key),
		Value: obj,
	}
	return mc.Set(item)
}

func ReadFromCache(key keyInfo, v interface{}) error {
	_ = "breakpoint"
	item, err := mc.Get(string(key))
	if err != nil {
		return err
	}
	err = decode(item.Value, v)
	if err != nil {
		return err
	}
	return nil
}

func encode(obj interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(obj)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decode(data []byte, v interface{}) error {
	decoder := gob.NewDecoder(bytes.NewBuffer(data))
	err := decoder.Decode(v)
	if err != nil {
		return err
	}
	return nil
}
