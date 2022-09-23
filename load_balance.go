package net

type LoadBalanceStrategy func([]*EventLoop) *EventLoop

func RoundRobin() LoadBalanceStrategy {
	var nextLoopIndex int
	return func(loops []*EventLoop) *EventLoop {
		l := loops[nextLoopIndex]
		nextLoopIndex = (nextLoopIndex + 1) % len(loops)
		return l
	}
}
