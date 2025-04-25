package main

import "testing"

func Example_makit() {
	goMain([]string{"makit"})
	// Output:
	// Welcome to WildCherry!
}

func TestHello(t *testing.T) {
	got := hello()
	want := "Welcome to makit!"
	if got != want {
		t.Errorf("hello() = %q, want %q", got, want)
	}
}