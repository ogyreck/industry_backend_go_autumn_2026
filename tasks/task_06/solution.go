package main

import "container/list"

type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {

	if capacity <= 0 {
		capacity = 0
	}
	return &LRUCache[K, V]{
		capacity: capacity,
		ll:       list.List{},
		items:    make(map[K]*list.Element, capacity),
	}
}

func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {

	var el *list.Element
	var el_res entry[K, V]
	var zero_value V

	el, ok = c.items[key]

	if ok {
		c.ll.MoveToFront(el)
		el_res = el.Value.(entry[K, V])
		return el_res.value, ok
	}

	return zero_value, ok

}
func (c *LRUCache[K, V]) Set(key K, value V) {

	//Добавляем новое значение в начало
	//Все предыдущие значения переходят на новое место (текущее -1).
	//Элемент который становится > capacity - вытесняются

	if c.capacity == 0 {
		return
	}
	v, exists := c.items[key]

	if !exists {
		item := c.ll.PushFront(entry[K, V]{key, value})

		if c.ll.Len() > c.capacity {
			last_el := c.ll.Back()
			key_last_el := last_el.Value.(entry[K, V])

			delete(c.items, key_last_el.key)
			c.ll.Remove(last_el)
		}

		v = item
	} else {
		v.Value = entry[K, V]{key: key, value: value}
	}

	c.items[key] = v

}

type LRU[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
}
