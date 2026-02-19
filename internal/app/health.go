package app

import (
	"fmt"
	"net"
	"time"
)

func CheckHealth(host string, port int, timeout time.Duration) (bool, error) {
	if port == 0 {
		port = 22
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false, err
	}
	_ = conn.Close()
	return true, nil
}
