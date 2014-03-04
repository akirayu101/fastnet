package srvnode

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type Client struct {
	Name     string
	ServAddr string
	conn     net.Conn
	wg       sync.WaitGroup
	mutex    sync.Mutex

	exit            chan bool
	send            chan []byte
	seq             uint64
	remotecall_chan map[uint64]chan []byte
	serve_msgids    map[int32]bool
}

func NewClient(name string, servaddr string, serve_msgids map[int32]bool) *Client {
	client := new(Client)
	client.Name = name
	client.ServAddr = servaddr
	client.serve_msgids = serve_msgids
	client.exit = make(chan bool)
	client.send = make(chan []byte)
	client.remotecall_chan = make(map[uint64]chan []byte)
	return client
}

func (c *Client) Connect() (err error) {
	log.Println("srvnode to ", c.Name)
	addr, err := net.ResolveTCPAddr("tcp", c.ServAddr)

	if err != nil {
		log.Println(err)
		return
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Println(err)
		return
	}
	c.conn = conn

	go c.StartAgent()
	return
}

func (c *Client) StartAgent() {
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

func (c *Client) Close() {
	close(c.exit)
	close(c.send)
	c.conn.Close()
}

func (c *Client) _send() {
	for {
		select {
		case <-c.exit:
			return
		case data := <-c.send:
			if _, err := c.conn.Write(data); err != nil {
				log.Println("sending data error:", data)
				continue
			}
		}
	}
}

func (c *Client) _recv() {
	for {
		select {
		case <-c.exit:
			return
		default:
			break
		}

		var data []byte
		header := make([]byte, 14)
		n, err := io.ReadFull(c.conn, header)
		if n == 0 && err == io.EOF {
			log.Println("server disconnect")
			break
		} else if err != nil {
			log.Println("receiving data error:", err)
		}
		size := binary.BigEndian.Uint16(header)
		crc_recv := binary.BigEndian.Uint32(header[2:6])
		data = make([]byte, size)
		n, err = io.ReadFull(c.conn, data)

		if uint16(n) != size {
			log.Println("data length error:", n, "!=", size)
			continue
		}

		if err != nil {
			log.Println("reading data error", err)
			continue
		}

		crc := crc32.Checksum(data, crc32.IEEETable)

		if crc_recv != crc {
			log.Println("crc error: ", crc_recv, " != ", crc)
			continue
		}
		seqId := binary.BigEndian.Uint64(header[6:])
		if seqId > 0 {
			ch, ok := c.popRemote(seqId)
			if ok {
				ch <- data
				continue
			}
		}
		MainHandle.ClientHandle(data)
	}
	c.exit <- false
	return
}

func (c *Client) RemoteCall(data []byte) (res []byte, err error) {
	ch := make(chan []byte)
	seqid := atomic.AddUint64(&c.seq, 1)

	if err := c.sendRemote(data, seqid); err != nil {
		return nil, err
	}

	c.addRemote(seqid, ch)
	for {
		select {
		case <-c.exit:
			return nil, errors.New("Exiting, cannot recv remote data")
		case res = <-ch:
			break
		}
	}
	close(ch)
	return res, nil

}

func (c *Client) sendRemote(data []byte, seq uint64) (err error) {
	select {
	case <-c.exit:
		return errors.New("Exiting, cannot send data")
	case c.send <- data:
		break
	}
	return nil
}

func (c *Client) addRemote(id uint64, ch chan []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.remotecall_chan[id] = ch
}

func (c *Client) popRemote(id uint64) (chan []byte, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ch, ok := c.remotecall_chan[id]
	if !ok {
		return nil, false
	}
	delete(c.remotecall_chan, id)
	return ch, true

}
