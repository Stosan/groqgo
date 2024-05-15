package main

import (
	sr "groqgo/src"
	"reflect"
	"testing"
)

func TestChatGroq(t *testing.T) {
	type args struct {
		kwargs []map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want *sr.ChatArgs
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ChatGroq(tt.args.kwargs...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ChatGroq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroqChatArgs_Call(t *testing.T) {
	type args struct {
		prompt string
	}
	tests := []struct {
		name string
		ca   *GroqChatArgs
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ca.Call(tt.args.prompt); got != tt.want {
				t.Errorf("GroqChatArgs.Call() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroqChatArgs_Stream_(t *testing.T) {
	type args struct {
		prompt string
	}
	tests := []struct {
		name string
		ca   *GroqChatArgs
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ca.Stream_(tt.args.prompt); got != tt.want {
				t.Errorf("GroqChatArgs.Stream_() = %v, want %v", got, tt.want)
			}
		})
	}
}
