package wexapi

import (
	"math"
	"testing"
)

func TestClient_nonce(t *testing.T) {
	tests := []struct {
		name    string
		nonce   uint32
		want    string
		wantErr bool
	}{
		{
			name:    "normal flow",
			want:    "1",
			wantErr: false,
			nonce:   1,
		},
		{
			name:    "overflow",
			want:    "",
			wantErr: true,
			nonce:   uint32(math.MaxUint32) - 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := Client{
				noncePool: make(chan uint32),
			}
			go func() {
				cli.noncePool <- tt.nonce
			}()
			got, err := cli.nonce()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.nonce() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Client.nonce() = %v, want %v", got, tt.want)
			}
		})
	}
}
