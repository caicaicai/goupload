package trigger

import "time"

// TimerTrigger 基于固定时间间隔周期性触发。
type TimerTrigger struct {
	interval time.Duration
}

// NewTimerTrigger 创建一个按interval周期触发的TimerTrigger。
func NewTimerTrigger(interval time.Duration) *TimerTrigger {
	return &TimerTrigger{interval: interval}
}

// Start 启动定时循环，每隔interval调用一次onTrigger，返回的stop函数用于停止定时器。
func (t *TimerTrigger) Start(onTrigger func()) (stop func()) {
	ticker := time.NewTicker(t.interval)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				onTrigger()
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return func() {
		close(done)
	}
}
