package ticker

import "time"

type realTicker struct {
	ticker *time.Ticker
}

func NewRealTicker(d time.Duration) *realTicker {
	return &realTicker{ticker: time.NewTicker(d)}
}

func (t *realTicker) Chan() <-chan time.Time {
	return t.ticker.C
}

func (t *realTicker) Stop() {
	t.ticker.Stop()
}
