package mspooler

import (
	"github.com/elastic/beats/filebeat/input"

	"sync"
	"sync/atomic"
)

type mSpoolerOutlet struct {
	wg       *sync.WaitGroup
	done     <-chan struct{}
	mSpooler *Mspooler
	isOpen   int32 // atomic indicator
}

func NewMspoolerOutlet(done <-chan struct{}, m *Mspooler, wg *sync.WaitGroup) *mSpoolerOutlet {
	return &mSpoolerOutlet{
		done:     done,
		mSpooler: m,
		wg:       wg,
		isOpen:   1,
	}
}

func (o *mSpoolerOutlet) OnEvent(event *input.Data) bool {
	open := atomic.LoadInt32(&o.isOpen) == 1
	if !open {
		return false
	}

	if o.wg != nil {
		o.wg.Add(1)
	}

	select {
	case <-o.done:
		if o.wg != nil {
			o.wg.Done()
		}
		atomic.StoreInt32(&o.isOpen, 0)
		return false
	case o.mSpooler.Channel <- event:
		return true
	}
}
