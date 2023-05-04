package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const (
	kMaxHeight = 8
	kBraching  = 2
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type node[T Comparable] struct {
	Val    T
	next   [kMaxHeight]*node[T]
	prev   [kMaxHeight]*node[T]
	height int
}

func newNode[T Comparable](val T, h int) *node[T] {
	return &node[T]{
		Val:    val,
		height: h,
	}
}

func newEmptyNode[T Comparable](h int) *node[T] {
	return &node[T]{
		height: h,
	}
}

type pool[T Comparable] []*node[T]

func (p *pool[T]) get() *node[T] {
	l := len(*p)
	if l == 0 {
		return &node[T]{}
	}

	n := (*p)[l-1]
	*p = (*p)[0 : l-1]
	return n
}

func (p *pool[T]) put(n *node[T]) {
	*p = append(*p, n)
	//n.Val = nil
}

type skiplist[T Comparable] struct {
	head      *node[T]
	tail      *node[T]
	maxHeight int
	height    int
	len       int
	capacity  int
	rnd       *rand.Rand
	pool      pool[T]
	Nil       T
}

func NewPriorityDeque[T Comparable](cap int) Deque[T] {
	return NewSkipList[T](cap)
}

func NewSkipList[T Comparable](cap int) *skiplist[T] {
	head := newEmptyNode[T](kMaxHeight)
	tail := newEmptyNode[T](kMaxHeight)

	for i := 0; i < kMaxHeight; i++ {
		head.next[i] = tail
		tail.prev[i] = head
	}

	return &skiplist[T]{
		head:     head,
		tail:     tail,
		capacity: cap,
		rnd:      rand.New(rand.NewSource(time.Now().UnixNano())),
		pool:     make(pool[T], 0, cap),
	}
}

func (s *skiplist[T]) randomHeight() int {

	h := 1
	for h < kMaxHeight && s.rnd.Int()%kBraching == 0 {
		h++
	}

	return h
}

func (s *skiplist[T]) less(x *node[T], base *node[T]) bool {
	if base == s.tail {
		return true
	} else if base == s.head {
		return false
	}

	return x.Val.Less(base.Val)
}

/*
func (s *skiplist[T]) find(val T) *node[T] {
	fake := &node[T]{Val: val}
	cur := s.head

	for height := s.height - 1; height >= 0; {
		if s.less(fake, cur.next[height]) {
			height--
		} else {
			cur = cur.next[height]
		}
	}

	for cur != s.tail {
		if cur.Val == val {
			return cur
		}
	}

	return nil
}*/

func (s *skiplist[T]) removeNode(n *node[T]) {
	if n == s.head || n == s.tail {
		panic("head and tail are not candidate for removal")
	}

	for i := 0; i < n.height; i++ {
		n.prev[i].next[i] = n.next[i]
		n.next[i].prev[i] = n.prev[i]
	}

	s.len--
	s.pool.put(n)
}

func (s *skiplist[T]) insert(ele T) {
	height := s.randomHeight()
	n := s.pool.get()
	n.height = height
	n.Val = ele

	if height > s.height {
		s.height = height
	}

	for level, cur := height-1, s.head; level >= 0; {
		if s.less(n, cur.next[level]) {
			n.next[level] = cur.next[level]
			n.prev[level] = cur
			cur.next[level] = n
			n.next[level].prev[level] = n
			level--
		} else {
			cur = cur.next[level]
		}
	}

	s.len++
}

func (s *skiplist[T]) Empty() bool {
	return s.head.next[0] == s.tail
}

func (s *skiplist[T]) PeekFirst() (T, bool) {
	if s.Empty() {
		return s.Nil, false
	}

	return s.head.next[0].Val, true
}

func (s *skiplist[T]) PollFirst() (T, bool) {
	if s.Empty() {
		return s.Nil, false
	}

	defer s.removeNode(s.head.next[0])
	return s.PeekFirst()
}

func (s *skiplist[T]) PeekLast() (T, bool) {
	if s.Empty() {
		return s.Nil, false
	}

	return s.tail.prev[0].Val, true
}

func (s *skiplist[T]) PollLast() (T, bool) {
	if s.Empty() {
		return s.Nil, false
	}

	defer s.removeNode(s.tail.prev[0])
	return s.PeekLast()
}

func (s *skiplist[T]) Offer(val T) (T, bool) {
	if s.len < s.capacity {
		s.insert(val)
		return s.Nil, true
	}

	last, _ := s.PeekLast()
	if last.Less(val) {
		return s.Nil, false
	}

	old := s.tail.prev[0].Val
	s.removeNode(s.tail.prev[0])
	s.insert(val)
	return old, true
}

func (s *skiplist[T]) Len() int {
	return s.len
}

func (s *skiplist[T]) String() string {
	var sb strings.Builder
	sb.WriteString("\n")
	for h := s.height; h > 0; h-- {
		for cur := s.head; cur != nil; cur = cur.next[0] {
			if /*len(cur.next)*/ cur.height < h {
				sb.WriteString("->")
			} else {
				sb.WriteString(fmt.Sprintf(" %v", cur.Val))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
