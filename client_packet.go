package gorcon

import (
	"bytes"
	"encoding/binary"
	"time"
)

type ClientPacket struct {
	ID   int32
	Type ClientPacketType
	Body string
}

func (cp *ClientPacket) Size() int32 {
	return PacketHeaderSize + int32(len(cp.Body)) + PacketTerminatorSize
}

func (cp *ClientPacket) WriteTo(rc *RemoteConsole) (n int64, err error) {
	buffer := bytes.NewBuffer(make([]byte, cp.Size()+4))
	if err := binary.Write(buffer, binary.LittleEndian, cp.Size()); err != nil {
		return n, err
	}
	n += 4

	if err := binary.Write(buffer, binary.LittleEndian, cp.ID); err != nil {
		return n, err
	}
	n += 4

	if err := binary.Write(buffer, binary.LittleEndian, cp.Type); err != nil {
		return n, err
	}
	n += 4

	if bufferN, err := buffer.Write(append([]byte(cp.Body), 0x00, 0x00)); err != nil {
		return n + int64(bufferN), err
	}

	if rc.settings.WriteDeadline > 0 {
		err := rc.conn.SetWriteDeadline(time.Now().Add(rc.settings.WriteDeadline))
		if err != nil {
			return n, err
		}
	}
	bufferN, err := buffer.WriteTo(rc.conn)

	return n + bufferN, err
}
