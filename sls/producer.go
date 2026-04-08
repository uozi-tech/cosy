package sls

import (
	"fmt"
	"sync"
	"time"
)

// ProducerConfig configures the batch log producer.
type ProducerConfig struct {
	Endpoint    string
	Credentials Credentials
	LogTags     []*LogTag
	LogFunc     LogFunc // optional, for internal diagnostics

	MaxBatchSize  int           // logs per flush, default 4096
	MaxBatchDelay time.Duration // max time before flush, default 200ms
	MaxRetries    int           // retry on 500/502/503, default 3
}

func (c *ProducerConfig) defaults() {
	if c.MaxBatchSize <= 0 {
		c.MaxBatchSize = 4096
	}
	if c.MaxBatchDelay <= 0 {
		c.MaxBatchDelay = 200 * time.Millisecond
	}
	if c.MaxRetries <= 0 {
		c.MaxRetries = 3
	}
}

type logEntry struct {
	project  string
	logstore string
	topic    string
	source   string
	log      *Log
}

// batchKey groups logs destined for the same endpoint.
type batchKey struct {
	project  string
	logstore string
	topic    string
	source   string
}

type batch struct {
	key  batchKey
	logs []*Log
}

// Producer is an asynchronous, batched SLS log producer.
type Producer struct {
	config ProducerConfig
	client *Client
	ch     chan logEntry
	quit   chan struct{}
	wg     sync.WaitGroup
}

// NewProducer creates a new Producer. Call Start() to begin processing.
func NewProducer(config ProducerConfig) (*Producer, error) {
	config.defaults()
	return &Producer{
		config: config,
		client: NewClient(config.Endpoint, config.Credentials),
		ch:     make(chan logEntry, 65536),
		quit:   make(chan struct{}),
	}, nil
}

// Start launches the background flush goroutine.
func (p *Producer) Start() {
	p.wg.Add(1)
	go p.loop()
}

// SendLog enqueues a log for async batched delivery.
func (p *Producer) SendLog(project, logstore, topic, source string, l *Log) error {
	select {
	case p.ch <- logEntry{project, logstore, topic, source, l}:
		return nil
	default:
		return fmt.Errorf("sls producer: queue full")
	}
}

// SafeClose signals shutdown, flushes remaining logs and blocks until done.
func (p *Producer) SafeClose() {
	close(p.quit)
	p.wg.Wait()
}

func (p *Producer) loop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.MaxBatchDelay)
	defer ticker.Stop()

	batches := make(map[batchKey]*batch)

	flush := func() {
		for _, b := range batches {
			p.send(b)
		}
		batches = make(map[batchKey]*batch)
	}

	for {
		select {
		case entry, ok := <-p.ch:
			if !ok {
				flush()
				return
			}
			key := batchKey{entry.project, entry.logstore, entry.topic, entry.source}
			b, exists := batches[key]
			if !exists {
				b = &batch{key: key}
				batches[key] = b
			}
			b.logs = append(b.logs, entry.log)
			if len(b.logs) >= p.config.MaxBatchSize {
				p.send(b)
				delete(batches, key)
			}

		case <-ticker.C:
			flush()

		case <-p.quit:
			// drain remaining entries
		drain:
			for {
				select {
				case entry, ok := <-p.ch:
					if !ok {
						break drain
					}
					key := batchKey{entry.project, entry.logstore, entry.topic, entry.source}
					b, exists := batches[key]
					if !exists {
						b = &batch{key: key}
						batches[key] = b
					}
					b.logs = append(b.logs, entry.log)
				default:
					break drain
				}
			}
			flush()
			return
		}
	}
}

func (p *Producer) send(b *batch) {
	if len(b.logs) == 0 {
		return
	}
	lg := &LogGroup{
		Logs:    b.logs,
		Topic:   b.key.topic,
		Source:  b.key.source,
		LogTags: p.config.LogTags,
	}

	var err error
	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		err = p.client.PutLogs(b.key.project, b.key.logstore, lg)
		if err == nil {
			return
		}
		// only retry on server errors
		if slsErr, ok := err.(*Error); ok {
			if slsErr.HTTPCode == 500 || slsErr.HTTPCode == 502 || slsErr.HTTPCode == 503 {
				time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
				continue
			}
		}
		break
	}
	if err != nil && p.config.LogFunc != nil {
		p.config.LogFunc("error", "sls put logs failed", "error", err)
	}
}
