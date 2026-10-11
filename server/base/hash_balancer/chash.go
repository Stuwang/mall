package hash_balancer

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Node[T any] struct {
	v   T
	key string
	c   int
}

type RingItem struct {
	Index int
	Hash  uint32
}

type CHash[T any] struct {
	Nodes []Node[T]
	Ring  []RingItem
}

func (h *CHash[T]) Add(key string, v T, c int) {
	h.Nodes = append(h.Nodes, Node[T]{
		v:   v,
		key: key,
		c:   c,
	})
}

func (h *CHash[T]) Build() {
	for index, n := range h.Nodes {
		for i := 0; i < n.c; i++ {
			var key = strconv.Itoa(i) + n.key
			v := h.hashKey(key)
			h.Ring = append(h.Ring, RingItem{
				Index: index,
				Hash:  v,
			})
		}
	}
	sort.Slice(h.Ring, func(i, j int) bool {
		return h.Ring[i].Hash < h.Ring[j].Hash
	})
}

func NewHash[T any]() *CHash[T] {
	ch := &CHash[T]{
		Nodes: nil,
		Ring:  nil,
	}
	return ch
}

func (h *CHash[T]) hashKey(key string) uint32 {
	crc := crc32.ChecksumIEEE([]byte(key))
	return crc
}

func (h *CHash[T]) Get(key string) T {
	v := h.hashKey(key)
	index := sort.Search(len(h.Ring), func(i int) bool {
		return v <= h.Ring[i].Hash
	})
	if index == len(h.Ring) {
		index = 0
	}
	return h.Nodes[h.Ring[index].Index].v
}

func (h *CHash[T]) Hash(key string) (res Node[T]) {
	v := h.hashKey(key)
	index := sort.Search(len(h.Ring), func(i int) bool {
		return v <= h.Ring[i].Hash
	})
	if index == len(h.Ring) {
		index = 0
	}
	return h.Nodes[h.Ring[index].Index]
}
