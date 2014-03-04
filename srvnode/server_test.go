package srvnode

import (
	"fmt"
	"testing"
)

func Test_server(t *testing.T) {
	s := NewServer("Akira", "8089", 10000)
	fmt.Println(MainHandle.Remote.GetFunc(123))
	s.Start()
}
