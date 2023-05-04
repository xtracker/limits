package util

import (
	"math/rand"
	"testing"
)

type IntVal int

func (i IntVal) Less(other Comparable) bool {
	return i > (other.(IntVal))
}
func TestSkipList(t *testing.T) {
	//rand.Seed(time.Now().UnixNano())
	s := NewSkipList[IntVal](5)

	for i := 0; i < 20; i++ {
		e := IntVal(rand.Intn(100))
		t.Logf("begin insert %v", e)
		s.Offer(e)
		//t.Logf("insert result : %v %v len=%v", o, ok, s.Len())
	}

	t.Logf("final result is: %v", s)
	//s.removeNode(s.find(IntVal(4)))
	t.Logf("find result is: %v", s)
	s.removeNode(s.head.next[0])
	t.Fatalf("remove head:%v", s)
}
