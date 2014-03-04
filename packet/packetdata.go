package packet

import (
	"hash/crc32"
)

func PacketData(seqId uint64, data []byte) []byte {
	writer := Writer()
	writer.WriteU16(uint16(len(data)))
	crc32 := crc32.Checksum(data, crc32.IEEETable)
	writer.WriteU32(crc32)
	writer.WriteU64(seqId)
	writer.WriteRawBytes(data)
	return writer.Data()
}
