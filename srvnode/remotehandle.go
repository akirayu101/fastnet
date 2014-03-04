package srvnode

import (
	"errors"
	"fmt"
)

type RemoteFunc func([]byte) ([]byte, error)

type remoteservice struct {
	GroupName string
	MsgId     int32
	Run       []RemoteFunc
}

type RemoteServices struct {
	remoteservices map[int32]*remoteservice
}

func NewRemoteServices(FuncNum int32) *RemoteServices {
	s := new(RemoteServices)
	s.remoteservices = make(map[int32]*remoteservice)
	return s
}

func (rs *RemoteServices) Register(groupname string, msgid int32, run RemoteFunc) {

	if _func, ok := rs.remoteservices[msgid]; !ok {
		s := new(remoteservice)
		s.MsgId = msgid
		s.Run = append(s.Run, run)
		rs.remoteservices[msgid] = s
	} else {
		_func.Run = append(_func.Run, run)
	}

}

func (rs *RemoteServices) remote_hash() (index int) {
	return 0
}

func (rs *RemoteServices) GetFunc(msgid int32) (run RemoteFunc, err error) {

	if _func, ok := rs.remoteservices[msgid]; ok {
		run = _func.Run[rs.remote_hash()]
	} else {
		err = errors.New(fmt.Sprintf("Can not get remote service, msgid:%d", msgid))
	}
	return

}
