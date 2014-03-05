package fastnet

import (
	"github.com/akirayu101/fastnet/srvconfig"
	"github.com/akirayu101/fastnet/srvnode"
	"log"
)

type FastNet struct {
	Server  *srvnode.Server
	Clients []*srvnode.Client
	Handle  *srvnode.Handle
}

func NewFastNet(configfile string) (fs *FastNet) {

	fs = new(FastNet)
	fs.Handle = srvnode.MainHandle
	config, err := srvconfig.ParseConfig(configfile)

	if err != nil {
		log.Println("config parse error", err)
		return
	}

	fs.Server = srvnode.NewServer(config.SC.Name, config.SC.Port, config.SC.MaxConn)

	for _, v := range config.CCS {
		client := srvnode.NewClient(v.Name, v.ServAddr, v.Msgids)
		fs.Clients = append(fs.Clients, client)

	}

	return

}
