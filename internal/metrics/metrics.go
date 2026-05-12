package metrics

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Stats struct {
	PacketsSent       uint64
	Success           uint64
	Fail              uint64
	ActiveConnections int32
	mu                sync.RWMutex
	LastSent          string
	LastReceived      string
}

func (s *Stats) AddSent() {
	atomic.AddUint64(&s.PacketsSent, 1)
}

func (s *Stats) AddSuccess() {
	atomic.AddUint64(&s.Success, 1)
}

func (s *Stats) AddFail() {
	atomic.AddUint64(&s.Fail, 1)
}

func (s *Stats) IncConn() {
	atomic.AddInt32(&s.ActiveConnections, 1)
}

func (s *Stats) DecConn() {
	atomic.AddInt32(&s.ActiveConnections, -1)
}

func (s *Stats) SetLastSent(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastSent = msg
}

func (s *Stats) SetLastReceived(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastReceived = msg
}

func (s *Stats) DisplayLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	start := time.Now()

	fmt.Print("\033[H\033[2J") // Clear terminal

	for range ticker.C {
		fmt.Printf("\033[H") // Move cursor to top
		fmt.Println("=== Load Tester Stats ===")
		fmt.Printf("Uptime:      %s\n", time.Since(start).Round(time.Second))
		fmt.Printf("Sent:        %d\n", atomic.LoadUint64(&s.PacketsSent))
		fmt.Printf("Success:     %d\n", atomic.LoadUint64(&s.Success))
		fmt.Printf("Fail:        %d\n", atomic.LoadUint64(&s.Fail))
		fmt.Printf("Active:      %d\n", atomic.LoadInt32(&s.ActiveConnections))
		
		s.mu.RLock()
		lastSent := s.LastSent
		lastRecv := s.LastReceived
		s.mu.RUnlock()
		
		fmt.Printf("Last Sent:   %s\n", lastSent)
		fmt.Printf("Last Recv:   %s\n", lastRecv)
		fmt.Println("=========================")
	}
}
