package main

import (
	"time"
)

func (p *Plugin) Run() {
	p.Stop()
	if !p.running {
		p.running = true
		p.runner()
	}
}

func (p *Plugin) Stop() {
	p.running = false
}

func (p *Plugin) runner() {
	go func() {
		<-time.NewTimer(time.Second).C
		if !p.running {
			return
		}

		p.TriggerReminders()
		p.runner()
	}()
}
