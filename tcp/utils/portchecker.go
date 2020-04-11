package utils

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func IsServiceAvailable(addr string) (isAvailable bool, err error) {
	urlParts := strings.Split(addr, ":")
	if len(urlParts) < 2 {
		return false, errors.New("please provide a valid address with port")
	}
	port, err := strconv.Atoi(urlParts[1])
	if err != nil {
		return false, err
	}
	return IsPortOpen(urlParts[0], port, time.Second*1), nil
}

func IsPortOpen(host string, port int, timeout time.Duration) bool {
	target := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", target, timeout)

	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			return IsPortOpen(host, port, timeout)
		}
		return false
	}
	defer conn.Close()
	return true
}
