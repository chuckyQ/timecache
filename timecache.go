package timecache

import (
	"sync"
	"time"
)

type cache[T any] struct {
	m        *sync.Map
	timeout  int
	timeouts *sync.Map
}

func (c *cache[T]) newTimeout() int {
	return int(time.Now().Unix()) + c.timeout
}

func New[T any](timeout int) *cache[T] {
	return &cache[T]{
		m:        &sync.Map{},
		timeout:  timeout,
		timeouts: &sync.Map{},
	}
}

func (c *cache[T]) Get(objID string, onExists func(string, T), onMiss func(string)) {

	obj, exists := c.m.Load(objID)
	if exists {
		onExists(objID, obj.(T))
		c.timeouts.Store(objID, c.newTimeout())
		return
	}
	onMiss(objID)
}

func (c *cache[T]) Store(objID string, obj T) {
	c.m.Store(objID, obj)

	timeout := int(time.Now().Unix()) + c.timeout
	c.timeouts.Store(objID, timeout)
}

func (c *cache[T]) Delete(objID string) {
	c.m.Delete(objID)
	c.timeouts.Delete(objID)
}

func (c *cache[T]) Start() {

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
