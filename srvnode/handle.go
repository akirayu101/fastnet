package srvnode

import (
	"../message"
	"../packet"
	"errors"
	"fmt"
	"log"
)

var MainHandle = NewHandle()

type Handle struct {
	Local  *LocalServices
	Remote *RemoteServices
}

func NewHandle() *Handle {
	h := new(Handle)
	h.Local = NewLocalServices(100)
	h.Remote = NewRemoteServices(100)
	return h
}

func (h *Handle) ServerHandle(conn *Connect, data []byte) (ackData []byte, err error) {
	reader := packet.Reader(data)
	uid, err := reader.ReadU64()
	if err != nil {
		errstr := fmt.Sprintf("读取用户UID出错")
		err = errors.New(errstr)
		return
	}

	if uid != conn.User.Uid {
		errstr := fmt.Sprintf("用户UID不正确,非法请求, uid:%d != sess.Uid:%d", uid, conn.User.Uid)
		err = errors.New(errstr)
		return
	}

	//读取消息ID
	msgid, err := reader.ReadS32()

	if err != nil {
		errstr := fmt.Sprintf("读取消息ID出错")
		err = errors.New(errstr)
		return
	}

	//读取MsgPack的数据
	reqData, err := reader.ReadAtLeast()

	if err != nil {
		errstr := fmt.Sprintf("读取数据包内容出错", err)
		err = errors.New(errstr)
		return
	}

	log.Println("接受用户:", uid, " 消息ID为:", msgid, " 数据为:", reqData)

	_handle, err := MainHandle.Local.GetFunc(msgid)

	if err == nil {
		ackData, err = _handle(conn.User, reqData)
		if err != nil {
			return
		}
		if msgid == message.MSG_REGISTER || msgid == message.MSG_LOGIN {
			conn.AddToServerMQ()
		} else if msgid == message.MSG_LOGOUT {
			conn.RemoveFromServerMQ()
		}
		return
	}

	_rhandle, err := MainHandle.Remote.GetFunc(msgid)

	if err == nil {
		//转发给其他服务器处理
		ackData, err = _rhandle(data)
		return
	}
	return
}

func (h *Handle) ClientHandle(data []byte) {
}

func (h *Handle) RegisterLocalService(name string, msgid int32, run LocalFunc) (err error) {
	return h.Local.Register(name, msgid, run)
}

func (h *Handle) RegisterRemoteService(groupname string, msgid int32, run RemoteFunc) {
	h.Remote.Register(groupname, msgid, run)
}
