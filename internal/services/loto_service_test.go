package services

import (
	"testing"

	"github.com/lib/pq"
	"lotoMironBot/internal/database"
)

func TestAllTicketsClosed(t *testing.T) {
	tests := []struct {
		name   string
		total  int
		closed int
		want   bool
	}{
		{name: "one of five", total: 5, closed: 1, want: false},
		{name: "four of five", total: 5, closed: 4, want: false},
		{name: "five of five", total: 5, closed: 5, want: true},
		{name: "three of three", total: 3, closed: 3, want: true},
		{name: "one of one", total: 1, closed: 1, want: true},
		{name: "no tickets", total: 0, closed: 0, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := allTicketsClosed(tt.total, tt.closed); got != tt.want {
				t.Fatalf("allTicketsClosed(%d, %d) = %t, want %t", tt.total, tt.closed, got, tt.want)
			}
		})
	}
}

func TestReachedMissLimit(t *testing.T) {
	for _, tt := range []struct {
		failures int
		want     bool
	}{{0, false}, {1, false}, {3, false}, {8, false}, {9, true}, {10, true}} {
		if got := reachedMissLimit(tt.failures); got != tt.want {
			t.Errorf("reachedMissLimit(%d) = %t, want %t", tt.failures, got, tt.want)
		}
	}
}

func TestIsTicketClosed(t *testing.T) {
	if !isTicketClosed(&database.GameTicket{Completed: true, Numbers: pq.Int32Array{1}}) {
		t.Fatal("completed ticket should be closed")
	}
	if !isTicketClosed(&database.GameTicket{Numbers: pq.Int32Array{}}) {
		t.Fatal("ticket with no remaining numbers should be closed")
	}
	if !isTicketClosed(&database.GameTicket{MarkedNumbers: pq.Int32Array{1, 2, 3, 4, 5, 6}}) {
		t.Fatal("ticket with all six numbers marked should be closed")
	}
	if isTicketClosed(&database.GameTicket{Numbers: pq.Int32Array{1}}) {
		t.Fatal("ticket with remaining numbers should remain open")
	}
}
