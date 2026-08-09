package app

import (
	"reflect"
	"testing"

	"ssh-manager/internal/config"
)

func TestConnectPreflightArgs(t *testing.T) {
	tests := []struct {
		name string
		conn config.Connection
		want []string
	}{
		{
			name: "normal connection",
			conn: config.Connection{
				Host:     "example.com",
				Port:     22,
				Username: "ubuntu",
			},
			want: []string{
				"-v",
				"-o", "BatchMode=yes",
				"-o", "StrictHostKeyChecking=accept-new",
				"-o", "ConnectTimeout=8",
				"-o", "NumberOfPasswordPrompts=0",
				"-o", "IdentitiesOnly=yes",
				"-p", "22",
				"ubuntu@example.com",
				"exit",
			},
		},
		{
			name: "connection with proxy identity and extra arguments",
			conn: config.Connection{
				Host:      "private.example.com",
				Port:      2200,
				Username:  "admin",
				ProxyJump: "bastion.example.com",
				KeyPath:   "/keys/deploy.pem",
				ExtraArgs: "-o ServerAliveInterval=60 -tt",
			},
			want: []string{
				"-v",
				"-o", "BatchMode=yes",
				"-o", "StrictHostKeyChecking=accept-new",
				"-o", "ConnectTimeout=8",
				"-o", "NumberOfPasswordPrompts=0",
				"-o", "IdentitiesOnly=yes",
				"-J", "bastion.example.com",
				"-i", "/keys/deploy.pem",
				"-p", "2200",
				"-o", "ServerAliveInterval=60", "-tt",
				"admin@private.example.com",
				"exit",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := connectPreflightArgs(tt.conn)
			if err != nil {
				t.Fatalf("connectPreflightArgs returned error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("connectPreflightArgs() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
