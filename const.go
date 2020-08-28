package gorcon

import (
	"errors"
	"time"
)

// ClientPacketType represents the packet types that can be sent by the client to the server.
type ClientPacketType int32

const (
	// TODO: Implement this response check.
	// CheckResponse is a SERVERDATA_RESPONSE_VALUE packet type, used for multi-packet response checking.
	CheckResponse ClientPacketType = 0

	// ExecuteCommand represents the SERVERDATA_EXECCOMMAND packet type.
	ExecuteCommand ClientPacketType = 2

	// Auth represents the SERVERDATA_AUTH packet type.
	Auth ClientPacketType = 3
)

// ServerPacketType represents the packet types that can be received by the client from the server.
type ServerPacketType int32

const (
	// ResponseValue represents the SERVERDATA_RESPONSE_VALUE packet type.
	ResponseValue ServerPacketType = 0

	// AuthResponse represents the SERVERDATA_AUTH_RESPONSE packet type.
	AuthResponse ServerPacketType = 2

	// RustUndocumentedPacket represents the unknown packet type from Rust with an ID of 4.
	RustUndocumentedPacket ServerPacketType = 4
)

const (
	// PacketTerminatorSize is the number of bytes the terminator (0x00) at the end of the packet uses.
	PacketTerminatorSize int32 = 2

	// PacketHeaderSize is the number of bytes the ID and Type fields use in bytes, both of which are 32-bit signed integers.
	PacketHeaderSize int32 = 8 //ID and Type, both 32-bit signed integers

	// PacketMaximumBodySize is the number of bytes the Body of the packet uses in bytes, it's 4096 minus the header and terminator.
	PacketMaximumBodySize int32 = 4096 - PacketHeaderSize - PacketTerminatorSize

	// PacketMaximumSize is the total number of bytes a full packet may contain.
	PacketMaximumSize int32 = 4 + PacketHeaderSize + PacketMaximumBodySize + PacketTerminatorSize
)

const (
	// DefaultDialTimeout is the timeout set for a zero value timeout, disabling it requires it be set to -1.
	DefaultDialTimeout = 5 * time.Second

	// DefaultReadDeadline is the deadline for reading from the socket, disabling it requires it be set to -1.
	DefaultReadDeadline = 5 * time.Second

	// DefaultWriteDeadline is the deadline for writing to the socket, disabling it requires it be set to -1.
	DefaultWriteDeadline = 5 * time.Second
)

var (
	ErrInvalidPacketTerminator     = errors.New("the packet was not terminated correctly")
	ErrInvalidResponseToAuthPacket = errors.New("the server responded with an invalid packet type for a auth packet")
	ErrAuthPacketFailed            = errors.New("the authentication attempt with the server failed")
	ErrInvalidPacketIDInResponse   = errors.New("the server replied with a packet ID not in our request")
	ErrCommandEmpty                = errors.New("the command must not be a blank string")
	ErrCommandTooLong              = errors.New("the supplied command was too long")
	ErrInvalidPacketType           = errors.New("packet type is invalid")
	ErrInvalidPacketTypeRust       = errors.New("packet type is invalid (rust undocumented)")
	ErrAlreadyConnected            = errors.New("remote console is already connected")
	ErrAlreadyAuthenticated        = errors.New("remote console is already authenticated")
)
