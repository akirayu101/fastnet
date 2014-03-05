package srvconfig

import (
	"github.com/robfig/config"
	"log"
	"strconv"
	"strings"
)

type ServerConfig struct {
	Name    string
	Port    string
	MaxConn int
}

type ClientConfig struct {
	Name     string
	ServAddr string
	Msgids   map[int32]bool
}

type ClientConfigs []ClientConfig

type Config struct {
	SC  ServerConfig
	CCS ClientConfigs
}

func ParseConfig(config_path string) (conf Config, err error) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	c, err := config.ReadDefault(config_path)
	clientnum, err := c.Int("Default", "clientnum")
	servername, err := c.String("Server", "name")
	serverport, err := c.String("Server", "port")
	servermaxconn, err := c.Int("Server", "maxconn")

	conf.SC = ServerConfig{servername, serverport, servermaxconn}

	log.Println(conf)
	log.Println(clientnum)
	for i := 0; i < clientnum; i++ {
		var CS ClientConfig
		name := "Client" + strconv.Itoa(i)
		CS.Name, err = c.String(name, "name")
		CS.ServAddr, err = c.String(name, "addr")

		response_api_str, _ := c.String(name, "response")
		noresponse_api_str, _ := c.String(name, "noresponse")

		response_apis := strings.Fields(response_api_str)
		noresponse_apis := strings.Fields(noresponse_api_str)
		CS.Msgids = make(map[int32]bool)

		for _, v := range response_apis {
			api_id, _ := strconv.Atoi(v)
			CS.Msgids[int32(api_id)] = true
		}

		for _, v := range noresponse_apis {
			api_id, _ := strconv.Atoi(v)
			CS.Msgids[int32(api_id)] = false
		}

		conf.CCS = append(conf.CCS, CS)
	}

	return

}
