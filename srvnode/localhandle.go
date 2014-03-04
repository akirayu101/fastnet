package srvnode

import (
	"../model"
	"errors"
	"fmt"
	"log"
)

type LocalFunc func(*model.UserInfo, []byte) ([]byte, error)

type localservice struct {
	Name string
	Run  LocalFunc
}

type LocalServices struct {
	localservices map[int32]*localservice
}

func NewLocalServices(FuncNum int32) *LocalServices {
	s := new(LocalServices)
	s.localservices = make(map[int32]*localservice, FuncNum)
	return s
}

func (ls *LocalServices) Register(name string, msgid int32, run LocalFunc) (err error) {

	if _func, ok := ls.localservices[msgid]; !ok {
		s := new(localservice)
		s.Name = name
		s.Run = run
		ls.localservices[msgid] = s
		log.Println("Binding local service:", name, " msgid:", msgid)

	} else {
		err = errors.New(fmt.Sprintf("Can not rebind local service, mgsid:%d,name:%s", msgid, _func.Name))
	}
	return
}

func (ls *LocalServices) GetFunc(msgid int32) (run LocalFunc, err error) {
	if _func, ok := ls.localservices[msgid]; ok {
		run = _func.Run
	} else {
		err = errors.New(fmt.Sprintf("Can not get local service, mgsid:%d", msgid))
	}
	return
}
