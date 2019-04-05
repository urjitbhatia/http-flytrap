package internal

import "log"

type storage interface {
	append(key string, value interface{})
	exists(key string) bool
	load(key string) []interface{}
	foreach(func(key string, value []interface{}) bool)
	delete(key string) bool
}

type memStore struct {
	data map[string][]interface{}
}

func newMemStore() storage {
	return &memStore{data: make(map[string][]interface{})}
}

func (ms *memStore) append(key string, value interface{}) {
	vals, ok := ms.data[key]
	if !ok {
		vals = []interface{}{}
	}
	vals = append(vals, value)
	ms.data[key] = vals
}

func (ms *memStore) exists(key string) bool {
	_, ok := ms.data[key]
	return ok
}

func (ms *memStore) load(key string) []interface{} {
	vals, _ := ms.data[key]
	return vals
}

func (ms *memStore) delete(key string) bool {
	_, ok := ms.data[key]
	if ok {
		log.Printf("Store deleting key: %s", key)
		delete(ms.data, key)
	}
	return ok
}

func (ms *memStore) foreach(f func(key string, values []interface{}) bool) {
	for k, v := range ms.data {
		if !f(k, v) {
			return
		}
	}
}
