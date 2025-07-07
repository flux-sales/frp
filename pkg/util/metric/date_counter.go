package metric

import (
	"sync"
	"time"
)

type DateCounter interface {
	TodayCount() int64
	GetLastDaysCount(days int64) []int64
	Inc(int64)
	Dec(int64)
	Snapshot() DateCounter
	Clear()
}

func NewDateCounter(reserveDays int64) DateCounter {
	if reserveDays <= 0 {
		reserveDays = 1
	}
	return &StandardDateCounter{
		reserveDays:    reserveDays,
		counts:         make([]int64, reserveDays),
		lastUpdateDate: dayStart(time.Now()),
	}
}

type StandardDateCounter struct {
	reserveDays    int64
	counts         []int64
	lastUpdateDate time.Time
	mu             sync.Mutex
}

func (c *StandardDateCounter) TodayCount() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rotate()
	return c.counts[0]
}

func (c *StandardDateCounter) GetLastDaysCount(days int64) []int64 {
	if days > c.reserveDays {
		days = c.reserveDays
	}
	result := make([]int64, days)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.rotate()

	copy(result, c.counts[:days])
	return result
}

func (c *StandardDateCounter) Inc(n int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rotate()
	c.counts[0] += n
}

func (c *StandardDateCounter) Dec(n int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rotate()
	c.counts[0] -= n
}

func (c *StandardDateCounter) Snapshot() DateCounter {
	c.mu.Lock()
	defer c.mu.Unlock()
	clone := &StandardDateCounter{
		reserveDays:    c.reserveDays,
		counts:         append([]int64{}, c.counts...), // copy
		lastUpdateDate: c.lastUpdateDate,
	}
	return clone
}

func (c *StandardDateCounter) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.counts {
		c.counts[i] = 0
	}
}

func (c *StandardDateCounter) rotate() {
	now := dayStart(time.Now())
	daysPassed := int(now.Sub(c.lastUpdateDate).Hours() / 24)

	if daysPassed <= 0 {
		return
	}

	if daysPassed >= int(c.reserveDays) {
		for i := range c.counts {
			c.counts[i] = 0
		}
	} else {
		copy(c.counts[daysPassed:], c.counts[:len(c.counts)-daysPassed])
		for i := 0; i < daysPassed; i++ {
			c.counts[i] = 0
		}
	}

	c.lastUpdateDate = now
}

func dayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
