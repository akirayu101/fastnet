package srvnode

import (
	"encoding/binary"
	"github.com/akirayu101/fastnet/model"
	"github.com/akirayu101/fastnet/packet"
	"hash/crc32"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Connect struct {

	//connect related
	conn net.Conn
	wg   sync.WaitGroup
	exit chan bool
	send chan []byte
	recv chan []byte
	serv *Server

	//session related
	User    *model.UserInfo
	IP      net.IP
	KickOut bool
	MQ      chan []byte

	// time related for stat
	ConnectTime    time.Time
	LastPacketTime int64
	LastFlushTime  int64
	OpCount        int
}

func NewConnect(conn net.Conn, s *Server) *Connect {

	connect := new(Connect)
	connect.conn = conn
	connect.exit = make(chan bool)
	connect.send = make(chan []byte)
	connect.recv = make(chan []byte)
	connect.IP = net.ParseIP(strings.Split(connect.conn.RemoteAddr().String(), ":")[0])
	connect.KickOut = false
	connect.MQ = make(chan []byte, 1024)
	connect.ConnectTime = time.Now()
	connect.LastPacketTime = time.Now().Unix()
	connect.LastFlushTime = time.Now().Unix()
	connect.OpCount = 0
	connect.serv = s
	connect.User = new(model.UserInfo)
	return connect

}

func (c *Connect) StartAgent() {
	c.wg.Add(2)
	go func() {
		defer c.wg.Done()
		c._recv()
	}()

	go func() {
		defer c.wg.Done()
		c._send()
	}()

	c.wg.Wait()
	c.Close()

}

//receive from client
func (c *Connect) _recv() {
	timer := time.NewTicker(1000000 * time.Nanosecond)
	var mq_data []byte

	for {
		select {
		case <-c.exit:
			return
		case <-timer.C:
		case mq_data = <-c.MQ:
			ackData, err := MainHandle.ServerHandle(c, mq_data)
			if err != nil {
				log.Println(err)
				continue
			}

			if ackData != nil {
				c.send <- packet.PacketData(0, ackData)
			}
		}
		// transport spec

		// read header
		header := make([]byte, 14)
		n, err := io.ReadFull(c.conn, header)
		if n == 0 && err == io.EOF {
			break
		} else if err != nil {
			log.Println("Data Error:", err)
		}

		//data size
		size := binary.BigEndian.Uint16(header)
		log.Println("Receive size:", size, "header data :", header)

		crc_recv := binary.BigEndian.Uint32(header[2:6])
		log.Println("Receive crc", crc_recv)
		data := make([]byte, size)
		n, err = io.ReadFull(c.conn, data)

		log.Println("Receive data:", data)
		if uint16(n) != size {
			log.Println("Data size incorrect:", n, "!=", size)
			continue
		}
		if err != nil {
			log.Println("Read data error:", err)
			continue
		}

		crc := crc32.Checksum(data, crc32.IEEETable)
		if crc_recv != crc {
			log.Println("Incorrect crc: ", crc_recv, " != ", crc)
			continue
		}

		seqid := binary.BigEndian.Uint64(header[6:])
		ackData, err := MainHandle.ServerHandle(c, data)
		if err != nil {
			log.Println(err)
			continue
		}

		if ackData != nil {
			c.send <- packet.PacketData(seqid, ackData)
		}

		//handle data here
		log.Println(data)
	}

	c.exit <- false

}

//send to client
func (c *Connect) _send() {
	for {
		select {
		case <-c.exit:
			return
		case data := <-c.send:
			if _, err := c.conn.Write(data); err != nil {
				log.Println("sending data error :", data)
				continue
			}
		}
	}
}

func (c *Connect) Close() {
	close(c.exit)
	c.conn.Close()
	close(c.send)
	close(c.recv)
	close(c.MQ)
	log.Println("closing connect", c.IP)
	c.RemoveFromServerMQ()
}

func (c *Connect) AddToServerMQ() {
	c.serv.AddConnMQ(c.User.Uid, c.MQ)
}

func (c *Connect) RemoveFromServerMQ() {
	c.serv.RemoveConnMQ(c.User.Uid)
}
