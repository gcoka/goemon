package main

import "testing"

func TestHello(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"world", "world", "hello world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Hello(tt.arg); got != tt.want {
				t.Errorf("Hello() = %v, want %v", got, tt.want)
			}
		})
	}
}
