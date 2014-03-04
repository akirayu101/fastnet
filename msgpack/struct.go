package msgpack

import (
	"../packet"
	"github.com/ugorji/go/codec"
)

func StructEncode(in interface{}) (out []byte) {
	var mh codec.MsgpackHandle
	mh.EncodeOptions.StructToArray = true
	encoder := codec.NewEncoderBytes(&out, &mh)
	encoder.Encode(in)
	return out
}

func StructDecode(in []byte, out interface{}) (err error) {
	var mh codec.MsgpackHandle
	decoder := codec.NewDecoderBytes(in, &mh)
	err = decoder.Decode(out)
	return
}

func InterEncode(uid uint64, msgid int32, reCode int32, ack interface{}) []byte {
	var mh codec.MsgpackHandle
	var out []byte
	mh.EncodeOptions.StructToArray = true
	encoder := codec.NewEncoderBytes(&out, &mh)
	encoder.Encode(ack)
	writer := packet.Writer()
	writer.WriteU64(uid)
	writer.WriteS32(msgid)
	writer.WriteS32(reCode)
	writer.WriteRawBytes(out)
	return writer.Data()
}
