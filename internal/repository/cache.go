package repository

import (
	"Level0/internal/model"
	"context"
	"log"
	"sync"
)

type Node struct {
	key   string
	value model.Order
	next  *Node
	prev  *Node
}

type LRUcache struct {
	mu       sync.Mutex
	db       *Storage
	capacity int
	cache    map[string]*Node
	head     *Node
	tail     *Node
}

func NewLRUCache(cap int, db *Storage) *LRUcache {
	if cap <= 0 {
		panic("Кэш не может быть меньше нуля")
	}
	head := new(Node)
	tail := new(Node)
	head.next = tail
	tail.prev = head
	cache := &LRUcache{
		capacity: cap,
		cache:    make(map[string]*Node),
		head:     head,
		tail:     tail,
		db:       db,
		mu:       sync.Mutex{},
	}

	orders, err := db.SelectOrders(context.Background(), cap)
	log.Printf("Инициализация кеша актуальными данными (кол-во: %d)", len(orders))
	if err != nil {
		log.Printf("Ошибка в инициализации кеша: %s", err.Error())
		return nil
	}

	for _, order := range orders {

		cache.push(order)
	}
	return cache
}

func (c *LRUcache) push(v model.Order) error {
	if node, ok := c.cache[v.OrderUID]; ok {
		node.value = v
		c.moveToFront(node)
		return nil
	}
	if len(c.cache) == c.capacity {
		back := c.tail.prev
		c.remove(back)
		delete(c.cache, back.key)
	}
	node := &Node{key: v.OrderUID, value: v}
	c.cache[v.OrderUID] = node
	c.pushToFront(node)
	return nil
}

func (c *LRUcache) Get(ctx context.Context, uid string) (*model.Order, error) {
	c.mu.Lock()

	if v, ok := c.cache[uid]; ok {
		c.moveToFront(v)
		cached := v.value
		c.mu.Unlock()
		return &cached, nil
	}

	c.mu.Unlock()

	order, err := c.db.GetOrder(ctx, uid)
	if err != nil {
		return nil, err
	}

	if order == nil {
		return nil, nil
	}

	c.mu.Lock()
	c.push(*order)
	c.mu.Unlock()
	return order, nil
}

func (c *LRUcache) moveToFront(n *Node) {
	if n == nil || n == c.head || n == c.tail {
		return
	}
	c.remove(n)
	c.pushToFront(n)
}

func (c *LRUcache) pushToFront(n *Node) {
	if n == nil || n == c.head || n == c.tail {
		return
	}
	n.prev = c.head
	n.next = c.head.next
	c.head.next.prev = n
	c.head.next = n
}

func (c *LRUcache) remove(n *Node) {
	if n == nil || n == c.head || n == c.tail {
		return
	}
	n.next.prev = n.prev
	n.prev.next = n.next
	n.next = nil
	n.prev = nil

}
