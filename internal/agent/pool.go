package agent

import (
	"context"
	"sync"
)

type RunPool struct {
	global chan struct{}
	mu     sync.Mutex
	parent map[string]chan struct{}
	maxParent int
}

func NewRunPool(maxGlobal, maxParent int) *RunPool {
	if maxGlobal < 1 {
		maxGlobal = 1
	}
	if maxParent < 1 {
		maxParent = 1
	}
	return &RunPool{
		global:    make(chan struct{}, maxGlobal),
		parent:    make(map[string]chan struct{}),
		maxParent: maxParent,
	}
}

func (p *RunPool) Acquire(ctx context.Context, parentRunID string) error {
	select {
	case p.global <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	if parentRunID == "" {
		return nil
	}

	p.mu.Lock()
	sem, ok := p.parent[parentRunID]
	if !ok {
		sem = make(chan struct{}, p.maxParent)
		p.parent[parentRunID] = sem
	}
	p.mu.Unlock()

	select {
	case sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		<-p.global
		return ctx.Err()
	}
}

func (p *RunPool) Release(parentRunID string) {
	if parentRunID != "" {
		p.mu.Lock()
		if sem, ok := p.parent[parentRunID]; ok {
			select {
			case <-sem:
			default:
			}
		}
		p.mu.Unlock()
	}
	select {
	case <-p.global:
	default:
	}
}

func (p *RunPool) UpdateLimits(maxGlobal, maxParent int) {
	if maxGlobal < 1 {
		maxGlobal = 1
	}
	if maxParent < 1 {
		maxParent = 1
	}
	p.mu.Lock()
	p.maxParent = maxParent
	p.mu.Unlock()
	p.global = make(chan struct{}, maxGlobal)
}
