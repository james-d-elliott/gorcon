package gorcon

import (
	"bytes"
	"encoding/binary"
	"time"
)

type ServerPacket struct {
	Size      int32
	ID        int32
	Type      ServerPacketType
	BodyBytes []byte
}

func NewServerPacketFromRemoteConsole(rc *RemoteConsole) (n int64, packet *ServerPacket, err error) {
	if rc.settings.ReadDeadline > 0 {
		err = rc.conn.SetReadDeadline(time.Now().Add(rc.settings.ReadDeadline))
		return n, packet, err
	}

	packet = &ServerPacket{}

	if err := binary.Read(rc.conn, binary.LittleEndian, &packet.Size); err != nil {
		return n, packet, err
	}
	n += 4

	if err := binary.Read(rc.conn, binary.LittleEndian, &packet.ID); err != nil {
		return n, packet, err
	}
	n += 4

	if err := binary.Read(rc.conn, binary.LittleEndian, &packet.Type); err != nil {
		return n, packet, err
	}
	n += 4

	var i int32
	packet.BodyBytes = make([]byte, packet.Size-PacketHeaderSize)

	for i < packet.Size-PacketHeaderSize {
		var m int
		if m, err := rc.conn.Read(packet.BodyBytes[i:]); err != nil {
			return n + int64(m) + int64(i), packet, err
		}
		i += int32(m)
	}

	n += int64(i)

	if !bytes.Equal(packet.BodyBytes[len(packet.BodyBytes)-int(PacketTerminatorSize):], []byte{0x00, 0x00}) {
		return n, packet, ErrInvalidPacketTerminator
	}

	// Ignore Rust's packet type 4
	if packet.Type != ResponseValue && packet.Type != AuthResponse {
		if packet.Type == RustUndocumentedPacket {
			return n, packet, ErrInvalidPacketTypeRust
		}
		return n, packet, ErrInvalidPacketType
	}
	return n, packet, nil
}

func (sp *ServerPacket) Body() string {
	return string(sp.BodyBytes[0 : len(sp.BodyBytes)-int(PacketTerminatorSize)])
}
