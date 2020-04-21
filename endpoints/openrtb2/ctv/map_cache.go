package ctv

import "fmt"

type MapKeyType = string

type MapCache struct {
	ICache
	cache map[MapKeyType]bool
}

func NewMapCache() *MapCache {
	return &MapCache{
		cache: make(map[MapKeyType]bool),
	}
}

func (c *MapCache) getKey(xImp, xIndex, yImp, yIndex int) MapKeyType {
	return fmt.Sprintf("%d:%d:%d:%d", xImp, xIndex, yImp, yIndex)
}

func (c *MapCache) Set(xImp, xIndex, yImp, yIndex int, value bool) {
	c.cache[c.getKey(xImp, xIndex, yImp, yIndex)] = value
}

func (c *MapCache) Get(xImp, xIndex, yImp, yIndex int) (bool, bool) {
	v, ok := c.cache[c.getKey(xImp, xIndex, yImp, yIndex)]
	return v, ok
}
