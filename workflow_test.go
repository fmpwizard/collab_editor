package main

import (
	"testing"

	"github.com/fmpwizard/go-quilljs-delta/delta"
)

func TestFlow(t *testing.T) {
	// https://medium.com/coinmonks/operational-transformations-as-an-algorithm-for-automatic-conflict-resolution-3bf8920ea447
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

	alice.AppendEvent(*delta.New(nil).Insert("Hello", nil))
	if x := alice.PendingLen(); x != 1 {
		t.Errorf("failed to get one pending item in alice's queue, got %d", x)
	}

	alice.SendPendingEvents(server)
	if x := alice.SentLen(); x != 1 {
		t.Errorf("\nexpected:\t1\ngot:\t\t%d", x)
	}

	alice.AppendEvent(*delta.New(nil).Insert(" World", nil))
	if x := alice.PendingLen(); x != 1 {
		t.Errorf("failed to get one pending item in alice's queue, got %d", x)
	}

	aliceDoc := alice.Document
	aliceExpected := delta.New(nil).Insert("Hello World", nil)
	if aliceDoc.Length() < 1 {
		t.Errorf("failed to get a non empty document: %d", aliceDoc.Length())
		t.FailNow()
	}

	if string(aliceDoc.Ops[0].Insert) != string(aliceExpected.Ops[0].Insert) {
		t.Errorf("\nexpected:\t%s\ngot:\t\t%s", string(aliceExpected.Ops[0].Insert), string(aliceDoc.Ops[0].Insert))
	}

	alice.SendPendingEvents(server)
	if x := alice.SentLen(); x != 1 {
		t.Errorf("\nexpected:\t1\ngot:\t\t%d", x)
	}

	bob.AppendEvent(*delta.New(nil).Insert("!", nil))
	if x := bob.PendingLen(); x != 1 {
		t.Errorf("failed to get one pending item in bob's queue, got %d", x)
	}

	bobDoc := bob.Document
	bobExpected := delta.New(nil).Insert("!", nil)
	if bobDoc.Length() < 1 {
		t.Errorf("failed to get a non empty document: %d", bobDoc.Length())
		t.FailNow()
	}

	if string(bobDoc.Ops[0].Insert) != string(bobExpected.Ops[0].Insert) {
		t.Errorf("\nexpected:\t%s\ngot:\t\t%s", string(bobExpected.Ops[0].Insert), string(bobDoc.Ops[0].Insert))
	}

	if string(bob.PendingChanges.Ops[0].Insert) != string(bobExpected.Ops[0].Insert) {
		t.Errorf("\nexpected:\t%s\ngot:\t\t%s", string(bobExpected.Ops[0].Insert), string(bob.PendingChanges.Ops[0].Insert))
	}

	server.ProcessEvents()
	// After  server processes events, it will send an ack to alice, when alice gets it, she will
	// reset her sentChanges queue.
	if x := alice.SentChanges.Length(); x != 0 {
		t.Errorf("failed to reset alice's sent queue, got: %d", x)
	}
	// check if Bob has the correct document
	aliceDoc = alice.Document
	aliceExpected = delta.New(nil).Insert("Hello World", nil)

	if string(aliceDoc.Ops[0].Insert) != string(aliceExpected.Ops[0].Insert) {
		t.Errorf("\nexpected:\t%s\ngot:\t\t%s", string(aliceExpected.Ops[0].Insert), string(aliceDoc.Ops[0].Insert))
	}

	if alice.LastSyncedRev != 1 {
		t.Errorf("expected rev 1, got %d", alice.LastSyncedRev)
	}

	bobDoc = bob.Document
	bobExpected = delta.New(nil).Insert("Hello!", nil)

	if string(bobDoc.Ops[0].Insert) != string(bobExpected.Ops[0].Insert) {
		t.Errorf("\nexpected:\t%s\ngot:\t\t%s", string(bobExpected.Ops[0].Insert), string(bobDoc.Ops[0].Insert))
	}

	if bob.LastSyncedRev != 1 {
		t.Errorf("expected rev 1, got %d", bob.LastSyncedRev)
	}

	bobExpected = delta.New(nil).Retain(5, nil).Insert("!", nil)
	if string(bobExpected.Ops[1].Insert) != string(bob.PendingChanges.Ops[1].Insert) {
		t.Errorf("\nexpected:\t%s\ngot:\t\t%s", string(bobExpected.Ops[1].Insert), string(bob.PendingChanges.Ops[1].Insert))
	}
	if *bobExpected.Ops[0].Retain != *bob.PendingChanges.Ops[0].Retain {
		t.Errorf("\nexpected:\t%d\ngot:\t\t%d", *bobExpected.Ops[0].Retain, *bob.PendingChanges.Ops[0].Retain)
	}

	// we make sure the pending list is empty for alice
	// before we send a new pending payload to the server
	// TODO: this has to be automatic, not a manual call
	if x := alice.SentLen(); x != 0 {
		t.Errorf("\nexpected:\t0\ngot:\t\t%d", x)
	}
	// send alice's event to the server
	alice.SendPendingEvents(server)
	if x := alice.SentLen(); x != 1 {
		t.Errorf("\nexpected:\t1\ngot:\t\t%d", x)
	}
	if x := len(alice.PendingChanges.Ops); x != 0 {
		t.Errorf("\nexpected:\t0\ngot:\t\t%d", x)
	}
	if x := alice.LastSyncedRev; x != 1 {
		t.Errorf("\nexpected:\t1\ngot:\t\t%d", x)
	}

	// send bob's event to the server
	bob.SendPendingEvents(server)
	if x := len(bob.PendingChanges.Ops); x != 0 {
		t.Errorf("\nexpected:\t0\ngot:\t\t%d", x)
	}

	if x := bob.SentLen(); x != 1 {
		t.Errorf("\nexpected:\t1\ngot:\t\t%d", x)
	}

	if x := bob.LastSyncedRev; x != 1 {
		t.Errorf("\nexpected:\t1\ngot:\t\t%d", x)
	}

}
