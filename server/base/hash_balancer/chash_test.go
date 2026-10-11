package hash_balancer

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
)

func BenchmarkName(b *testing.B) {
	for i := 0; i < b.N; i++ {

	}
}

func TestA(t *testing.T) {
	h1 := NewHash[int]()
	h2 := NewHash[int]()
	h1.Add("1", 1, 32)
	h1.Add("2", 2, 32)
	h1.Add("3", 3, 32)
	h1.Add("4", 4, 32)
	h1.Build()

	h2.Add("1", 1, 32)
	h2.Add("2", 2, 32)
	h2.Add("3", 3, 32)
	h2.Build()

	var (
		totalCount int
		sameCount  int
	)
	for i := 0; i < 1000*10000; i++ {
		totalCount++
		v := rand.Int63()
		v1 := h1.Get(strconv.FormatInt(v, 10))
		v2 := h2.Get(strconv.FormatInt(v, 10))
		if v1 == v2 {
			sameCount++
		}
	}
	fmt.Println(float64(sameCount)/float64(totalCount), sameCount, totalCount)
}
