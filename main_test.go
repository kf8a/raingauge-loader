package main

import "testing"

func TestPosFindingString(t *testing.T) {
	variates := stringSlice{"one", "two", "three"}
	if x := variates.pos("two"); x != 1 {
		t.Errorf("Expected 'two' to be at postion 1 but go %s", x)
	}
}
func TestPostNotFindingString(t *testing.T) {
	variates := stringSlice{"one", "two", "three"}
	if x := variates.pos("nothing"); x != -1 {
		t.Errorf("Expected 'two' to be at postion -1 but go %s", x)
	}
}
