package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestPort(t *testing.T) {
	isOpen := IsPortOpen("localhost", 5679, time.Second*1)
	fmt.Println(isOpen)
}
