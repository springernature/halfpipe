package main_test

import "testing"

func TestSomethingStupid(t *testing.T) {
	a := 1
	b := 1
	if a != b {
		t.Errorf("Failure")
	}
}
