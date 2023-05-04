package util

type Comparable interface {
	Less(other Comparable) bool
}

type Queue[T Comparable] interface {
	Len() int
	Offer(T) (T, bool)
	PeekFirst() (T, bool)
	PollFirst() (T, bool)
	Empty() bool
}

type Deque[T Comparable] interface {
	Queue[T]
	PeekLast() (T, bool)
	PollLast() (T, bool)
}
