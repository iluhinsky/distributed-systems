package counter

import (
	"context"

	"github.com/iluhinsky/distributed-systems/network"
	"github.com/iluhinsky/distributed-systems/util"
)

type replicatedCounts map[int64]int64

type EpidemicCounter struct {
	rid      int64
	counts   replicatedCounts
	replicas map[int64]network.Link
}

func NewEpidemicCounter(rid int64) *EpidemicCounter {
	return &EpidemicCounter{
		rid:      rid,
		counts:   replicatedCounts{},
		replicas: map[int64]network.Link{},
	}
}

func (c *EpidemicCounter) Inc() bool {
	c.counts[c.rid] = c.counts[c.rid] + 1
	return true
}

func (c *EpidemicCounter) Read() int64 {
	var sum = int64(0)
	for _, v := range c.counts {
		sum += v
	}
	return sum
}

func (c *EpidemicCounter) Introduce(rid int64, link network.Link) {
	if link != nil {
		c.replicas[rid] = link
	}
}

func (c *EpidemicCounter) Receive(rid int64, msg interface{}) interface{} {
	if update, ok := msg.(replicatedCounts); ok {
		for k, v := range update {
			c.counts[k] = util.Max(c.counts[k], v)
		}
	}
	return nil
}

func (c *EpidemicCounter) Periodically() {
	for _, r := range c.replicas {
		r.Send(context.Background(), c.counts)
	}
}
