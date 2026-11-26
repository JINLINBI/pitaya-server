package main

import "testing"

func TestNewRouterMonolithPassthrough(t *testing.T) {
	rt := newRouter(false)
	if got := rt("account.register"); got != "account.register" {
		t.Fatalf("got %q", got)
	}
}

func TestNewRouterClusterPrefixes(t *testing.T) {
	rt := newRouter(true)
	cases := map[string]string{
		"account.register": "account.account.register",
		"account.login":    "account.account.login",
		"mailbox.send":     "account.mailbox.send",
		"mailbox.list":     "account.mailbox.list",
		"room.join":        "room.room.join",
		"room.message":     "room.room.message",
		"connector.bind":   "connector.bind",
		"connector.login":  "connector.login",
	}
	for in, want := range cases {
		if got := rt(in); got != want {
			t.Fatalf("%s: want %q got %q", in, want, got)
		}
	}
}
