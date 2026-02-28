package timecache

import (
	"sync"
	"time"
)

type Cache[T any] struct {
	m        *sync.Map
	timeout  int
	timeouts *sync.Map
}

func (c *Cache[T]) newTimeout() int {
	return int(time.Now().Unix()) + c.timeout
}

func New[T any](timeout int) *Cache[T] {
	return &Cache[T]{
		m:        &sync.Map{},
		timeout:  timeout,
		timeouts: &sync.Map{},
	}
}

func (c *Cache[T]) Get(objID string, onExists func(string, T), onMiss func(string)) {

	obj, exists := c.m.Load(objID)
	if exists {
		onExists(objID, obj.(T))
		c.timeouts.Store(objID, c.newTimeout())
		return
	}
	onMiss(objID)
}

func (c *Cache[T]) Store(objID string, obj T) {
	c.m.Store(objID, obj)

	timeout := int(time.Now().Unix()) + c.timeout
	c.timeouts.Store(objID, timeout)
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
			c.timeouts.Range(func(key, value any) bool {

				k, ok := key.(string)
				if !ok {
					c.m.Delete(key)
					return true
				}

				expirationTime, ok := value.(int)
				if !ok {
					c.m.Delete(key)
					return true
				}

				if now > expirationTime {
					c.m.Delete(k)
				}

				return true

			})
		}
	}()

}
