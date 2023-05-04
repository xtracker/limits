package limits

import "context"

type Result int

const (
	SUCCESS Result = iota
	DROPPED        // request failed(eg. timeout, net_err)
	IGNORED        //ignore the result
)

type Listener func(context.Context, Result)

type Limiter interface {
	Acquire(context.Context) (Listener, error)
}
