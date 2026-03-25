package main

import (
	"sync"
	"time"
)

type Cache[T any] struct {
	m        *sync.Map
	timeout  int
	timeouts *sync.Map
	onDelete func(map[string]T)
}

func (c *Cache[T]) newTimeout() int {
	return int(time.Now().Unix()) + c.timeout
}

func New[T any](timeout int, onDelete func(map[string]T)) *Cache[T] {
	return &Cache[T]{
		m:        &sync.Map{},
		timeout:  timeout,
		timeouts: &sync.Map{},
		onDelete: onDelete,
	}
}

func (c *Cache[T]) Get(objID string, onExists func(string, T), onMiss func(string)) {

	obj, exists := c.m.Load(objID)
	if exists {
		c.timeouts.Store(objID, c.newTimeout())
		onExists(objID, obj.(T))
		return
	}
	onMiss(objID)
}

func (c *Cache[T]) Store(objID string, obj T) {
	timeout := int(time.Now().Unix()) + c.timeout
	c.timeouts.Store(objID, timeout)
	c.m.Store(objID, obj)
}

func (c *Cache[T]) Delete(objID string) {
	c.m.Delete(objID)
	c.timeouts.Delete(objID)
}

func (c *Cache[T]) Start() {

	go func() {

		t := time.Duration(c.timeout * int(time.Second))
		for {
			time.Sleep(t)
			now := int(time.Now().Unix())
			deleted := make(map[string]T)
			c.timeouts.Range(func(key, value any) bool {

				k, ok := key.(string)
				if !ok {
					c.m.Delete(key)
					c.timeouts.Delete(key)
					return true
				}
				expirationTime, ok := value.(int)

				if !ok {
					c.Delete(k)
					return true
				}

				obj, ok := c.m.Load(k)

				if !ok {
					c.Delete(k)
					return true
				}

				ob, ok := obj.(T)

				if !ok {
					c.Delete(k)
					return true
				}

				if expirationTime < now {
					deleted[k] = ob
				}

				return true

			})

			c.onDelete(deleted)
			for k := range deleted {
				c.Delete(k)
			}
		}
	}()

}
