package server

import (
	"log"
	"time"
)

func (l *Lobby) allAnswered() bool {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for _, client := range l.Clients {
		if !client.Answered {
			return false
		}
	}
	return true
}

func (l *Lobby) allVoted() bool {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for _, client := range l.Clients {
		if !client.Voted {
			return false
		}
	}
	return true
}

func startTimer(duration time.Duration, shouldExpireEarly func() bool, onExpire func()) {
	go func() {
		timer := time.NewTimer(duration)
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timer.C:
				log.Println("Timer expired.")
				onExpire()
				return
			case <-ticker.C:
				if shouldExpireEarly() {
					log.Println("All players done before timer.")
					timer.Stop()
					onExpire()
					return
				}
			}
		}
	}()
}
