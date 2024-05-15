package groqgo

import (
	"testing"
)

func TestClient(t *testing.T) {
	type args struct {
		qp *ChatArgs
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Client(tt.args.qp)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Client() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStreamClient(t *testing.T) {
	type args struct {
		qp *ChatArgs
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StreamClient(tt.args.qp)
			if (err != nil) != tt.wantErr {
				t.Errorf("StreamClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StreamClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
