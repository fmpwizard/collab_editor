package main

import (
	"log"

	"github.com/fmpwizard/go-quilljs-delta/delta"
)

// Client represents the data stored in the client, it keeps track
// of data sent, revision id, etc
type Client struct {
	LastSyncedRev  int64
	ID             string
	SentChanges    delta.Delta
	PendingChanges delta.Delta
	Document       delta.Delta
}

// Server represents the server state, pending changes, etc
type Server struct {
	ProcessedRev   int64
	PendingChanges []ChangePayload
	Document       delta.Delta
	// clients is a map of clientID => Client instance
	// unexported to control concurrent access
	clients map[string]*Client
}

// ChangePayload holds a clientID and their Delta. This is sent to the server, and the server
// can keep track of which client sent which Delta instance.
type ChangePayload struct {
	Delta         delta.Delta
	ClientID      string
	LastSyncedRev int64
}

// AppendEvent adds an Op to the client instance, this could be
// an insert, remain, or any other operation
func (c *Client) AppendEvent(ops delta.Delta) {
	c.PendingChanges = *c.PendingChanges.Concat(ops)
}

// SendPendingEvents sends the pending Op to the server and
// moves the Op to the SentChanges Delta
// TODO: looks at how to syncronize, because we would add more events to PendingChanges while we sent
// old changes to the server
func (c *Client) SendPendingEvents(s *Server) {
	if c.SentChanges.Length() > 0 {
		// There are changes we sent but were not ack'ed, so don't send
		// new changes
		return
	}
	// send ops to server
	p := ChangePayload{
		Delta:         c.PendingChanges,
		ClientID:      c.ID,
		LastSyncedRev: c.LastSyncedRev,
	}
	s.AppendEvent(p)
	// copy pending changes to sent changes
	c.SentChanges = *c.SentChanges.Compose(c.PendingChanges)
	// reset pending changes
	c.PendingChanges = *delta.New(nil)
}

// AckEvents receives an ack form the server that the last Delta sent was applied.
// We update our internal rev number to match the server
func (c *Client) AckEvents(rev int64) {
	c.LastSyncedRev = rev
	c.SentChanges = *delta.New(nil)
}

// PendingLen return the number of ops in the queue that have not been sent
// to the server.
// This helps in testing
func (c *Client) PendingLen() int {
	return len(c.PendingChanges.Ops)
}

// SentLen return the number of ops in the sent queue that have been sent
// to the server.
// This helps in testing
func (c *Client) SentLen() int {
	return len(c.SentChanges.Ops)
}

// AppendEvent takes an ops and adds it to the pending list of Deltas to apply
func (s *Server) AppendEvent(ops ChangePayload) {
	s.PendingChanges = append(s.PendingChanges, ops)
}

// ProcessEvents the server moves pending changes to the ProcessedChanges Delta
func (s *Server) ProcessEvents() {
	for _, p := range s.PendingChanges {
		s.Document = *s.Document.Compose(p.Delta)
		s.ProcessedRev++
		s.BroadcastChangeApplied(p.ClientID, s.ProcessedRev)
	}
	s.PendingChanges = nil
}

// BroadcastChangeApplied sends an ack to the client that sent the last Op
// and sends the op to all other clients
func (s *Server) BroadcastChangeApplied(cid string, rev int64) {
	log.Printf("ack for\ncid:\t%s\nrev:\t%d", cid, rev)
	x := s.clients[cid]
	x.AckEvents(rev)
}

// RegisterClient adds the client to the server Clients' list
func (s *Server) RegisterClient(c *Client) {
	if s.clients == nil {
		s.clients = make(map[string]*Client)
	}
	s.clients[c.ID] = c
}

// UnregisterClient removes the client from the server Clients' list
func (s *Server) UnregisterClient(c Client) {
	if s.clients == nil {
		s.clients = make(map[string]*Client)
	}
	delete(s.clients, c.ID)
}

// IsClientRegistered returns true if the clientID is in the server map
func (s *Server) IsClientRegistered(cid string) bool {
	_, ok := s.clients[cid]
	return ok
}
