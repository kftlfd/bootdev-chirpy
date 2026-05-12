package metrics

import "sync/atomic"

type Metrics struct {
	FileserverHits atomic.Int32
}

func New() *Metrics {
	return &Metrics{}
}
