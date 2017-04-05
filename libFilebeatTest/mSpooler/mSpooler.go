package mspooler

import (
	cfg "github.com/elastic/beats/filebeat/config"
	"github.com/elastic/beats/filebeat/input"
	"github.com/elastic/beats/filebeat/publisher"
	"github.com/tdelacour/libFilebeatTest/global"
	"github.com/tdelacour/libFilebeatTest/poster"

	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	mSpoolerStoppedErr = errors.New("the mspooler was stopped")
)

// Number of events the spooler can bugger before blocking
const channelSize = 16

type Mspooler struct {
	Channel   chan *input.Data        // input data from crawler
	timeout   time.Duration           // Maximum time between spool flushes
	spoolSize uint64                  // Maximum number of events batched before flush
	out       publisher.SuccessLogger // output data to registrar
	spool     []*input.Data           // Events being held by spooler
	wg        sync.WaitGroup          // WaitGroup for controlled shutdown
}

func New(out publisher.SuccessLogger, config *cfg.Config) (*Mspooler, error) {
	return &Mspooler{
		Channel:   make(chan *input.Data, channelSize),
		timeout:   config.IdleTimeout,
		spoolSize: config.SpoolSize,
		out:       out,
		spool:     make([]*input.Data, 0, config.SpoolSize),
	}, nil
}

func (m *Mspooler) Start() {
	m.wg.Add(1)
	fmt.Printf("Starting mspooler\n")
	go m.run()
}

func (m *Mspooler) run() {
	defer m.flush()
	defer m.wg.Done()

	timer := time.NewTimer(m.timeout)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()

	for {
		select {
		case event, ok := <-m.Channel:
			if !ok {
				return
			}
			if event != nil {
				flushed, err := m.queue(event)
				if err != nil {
					fmt.Printf("Error occured while flushing spool, stopping prematurely: %v\n", err)
				}
				if flushed {
					if !timer.Stop() {
						<-timer.C
					}
					timer.Reset(m.timeout)
				}
			}
		case <-timer.C:
			fmt.Printf("Flushing spooler because of timeout. Flushed: %d\n", len(m.spool))
			err := m.flush()
			if err != nil {
				fmt.Printf("Error occured while flushing spool, stopping prematurely: %v\n", err)
			}
			timer.Reset(m.timeout)
		}
	}
}

func (m *Mspooler) queue(event *input.Data) (bool, error) {
	flushed := false
	m.spool = append(m.spool, event)
	if len(m.spool) == cap(m.spool) {
		fmt.Printf("Flushing mspooler because buffer full. Flushed: %d\n", m.spoolSize)
		err := m.flush()
		if err != nil {
			return flushed, err
		}
		flushed = true
	}

	return flushed, nil
}

func (m *Mspooler) flush() error {
	tmp := make([]*input.Data, len(m.spool))
	copy(tmp, m.spool)

	// Reset the buffer
	m.spool = m.spool[:0]

	// Send this batch to the http server
	err := m.post(tmp)
	if err != nil {
		// If post fails, still want to continue with regitrar
		fmt.Printf("Unable to post input: %v\n", err)
	}

	// Send data events to the registrar
	ok := m.out.Published(tmp) // TODO consider renaming "Published"
	if !ok {
		return mSpoolerStoppedErr
	}

	return nil
}

func (m *Mspooler) post(batch []*input.Data) error {
	_, err := poster.HttpPostJson(global.Url, formatInputData(batch))
	return err
}

func (m *Mspooler) Stop() {
	fmt.Printf("Stopping mspooler\n")

	close(m.Channel)

	m.wg.Wait()
	fmt.Printf("Mongo spooler stopped\n")
}

func formatInputData(batch []*input.Data) []map[string]interface{} {
	events := make([]map[string]interface{}, 0, len(batch))
	for _, data := range batch {
		if data.HasData() {
			e := data.Event

			m, err := e.GetValue("message")
			if err != nil {
				fmt.Printf("Unable to get message: %v\n", err)
				continue
			}

			s, err := e.GetValue("source")
			if err != nil {
				fmt.Printf("Unable to get source: %v\n", err)
				continue
			}

			newEvent := map[string]interface{}{
				"message": m,
				"source":  s,
			}
			events = append(events, newEvent)
		}
	}

	return events
}
