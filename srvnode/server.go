package srvnode

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

type Server struct {
	Name    string
	Port    string
	MaxConn int

	namechan      map[uint64]chan []byte
	namechan_lock sync.RWMutex
}

func NewServer(name string, port string, maxconn int) *Server {
	s := new(Server)
	s.Name = name
	s.Port = port
	s.MaxConn = maxconn
	s.namechan = make(map[uint64]chan []byte, 1024)
	return s
}

func (s *Server) Start() {
	log.Println("Start Serving:", s.Name)
	tcp_addr, err := net.ResolveTCPAddr("tcp4", ":"+s.Port)
	if err != nil {
		log.Println("Start Serving Error:", err)
		return
	}

	listener, err := net.ListenTCP("tcp4", tcp_addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Client Error:", err)
			continue
		}
		c := NewConnect(conn, s)
		go c.StartAgent()
	}

}

func (s *Server) AddConnMQ(id uint64, ch chan []byte) {
	s.namechan_lock.Lock()
	defer s.namechan_lock.Unlock()
	s.namechan[id] = ch
}

func (s *Server) RemoveConnMQ(id uint64) {
	s.namechan_lock.Lock()
	defer s.namechan_lock.Unlock()
	delete(s.namechan, id)
}

func (s *Server) GetConnMQ(id uint64) (ch chan []byte, err error) {
	s.namechan_lock.RLock()
	defer s.namechan_lock.RUnlock()
	ch, ok := s.namechan[id]
	if !ok {
		err = errors.New(fmt.Sprintf("uid %x  not login in this server", id))
		return
	}
	return
}
