package transport

import (
	"container/ring"
	"fmt"
	//"time"
)

func Distance(src *ring.Ring, dst *ring.Ring) int {
	count := 0
	cur := src
	for {
		if cur == dst {
			return count
		}
		cur = cur.Next()
		count++
	}
	return count
}
func Ringtest() {
	//coffee := []string{"1", "2", "3"}

	// create a ring and populate it with some values
	r := ring.New(512)
	first := r
	second := r
	third := r
	for i := 0; i < r.Len(); i++ {
		r.Value = i
		r = r.Next()
	}

	// print all values of the ring, easy done with ring.Do()
	//	r.Do(func(x interface{}) {
	//		fmt.Println(x)
	//	})

	// .. or each one by one.
	//for _ = range time.Tick(time.Second * 1) {
	first = first.Move(-20)
	fmt.Println("first:", first.Value, first)
	second = second.Next()
	fmt.Println("second:", second.Value, second)
	third = third.Next()
	fmt.Println("third:", third.Value)
	fmt.Println("distance:", Distance(second, first))
	//}
}
