package main

import (
	"testing"
)

func TestFlow(t *testing.T) {

	alice := new(Client)
	server := new(Server)
	bob := new(Client)

	alice.ID = "alice"
	bob.ID = "bob"
	// Register clients

	server.RegisterClient(alice)
	if !server.IsClientRegistered(alice.ID) {
		t.Error("failed to register alice as a client")
	}

	server.RegisterClient(bob)
	if !server.IsClientRegistered(bob.ID) {
		t.Error("failed to register bob as a client")
	}

	alice.AppendEvent(*alice.PendingChanges.Insert("Hello", nil))
	if x := alice.PendingLen(); x != 1 {
		t.Errorf("failed to get one pending item in alice's queue, got %d", x)
	}

	alice.SendPendingEvents(server)
	if x := alice.SentLen(); x != 1 {
		t.Errorf("\nexpected:\t1\ngot:\t\t%d", x)
	}

	alice.AppendEvent(*alice.PendingChanges.Insert(" World", nil))
	if x := alice.PendingLen(); x != 1 {
		t.Errorf("failed to get one pending item in alice's queue, got %d", x)
	}

	alice.SendPendingEvents(server)
	if x := alice.SentLen(); x != 1 {
		t.Errorf("\nexpected:\t1\ngot:\t\t%d", x)
	}
	bob.AppendEvent(*bob.PendingChanges.Insert("!", nil))
	if x := bob.PendingLen(); x != 1 {
		t.Errorf("failed to get one pending item in bob's queue, got %d", x)
	}

	server.ProcessEvents()
	if x := alice.SentChanges.Length(); x != 0 {
		t.Errorf("failed to reset alice's sent queue, got: %d", x)
	}

}
