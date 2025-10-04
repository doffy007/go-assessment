package progress

import "sync"

type Tracker struct {
	Total     int `json:"total"`
	Processed int `json:"processed"`
	Percent   int `json:"percent"`
	mu        sync.Mutex
}

func New(total int) *Tracker {
	return &Tracker{Total: total}
}

func (p *Tracker) Inc() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Processed++
	p.Percent = int(float64(p.Processed) / float64(p.Total) * 100)
	return p.Percent
}

func (p *Tracker) Snapshot() Tracker {
	p.mu.Lock()
	defer p.mu.Unlock()
	return Tracker{
		Total:     p.Total,
		Processed: p.Processed,
		Percent:   p.Percent,
	}
}
