package gorcon

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

type RemoteConsoleSettings struct {
	DialTimeout   time.Duration
	ReadDeadline  time.Duration
	WriteDeadline time.Duration
}

type RemoteConsole struct {
	connected     bool
	authenticated bool
	conn          net.Conn
	mutex         sync.Mutex
	packetID      int32
	settings      RemoteConsoleSettings
}

// NewRemoteConsole creates a new remote console, dials it, and authenticates it.
func NewRemoteConsole(address string, password string, settings RemoteConsoleSettings) (rc *RemoteConsole, err error) {
	// Check Settings
	if settings.DialTimeout == 0 {
		settings.DialTimeout = DefaultDialTimeout
	}

	if settings.ReadDeadline == 0 {
		settings.ReadDeadline = DefaultReadDeadline
	}

	if settings.WriteDeadline == 0 {
		settings.WriteDeadline = DefaultWriteDeadline
	}

	rc = &RemoteConsole{
		packetID: 0,
		settings: settings,
	}

	if err := rc.Dial(address); err != nil {
		return nil, err
	}

	if err := rc.Authenticate(password); err != nil {
		if closeErr := rc.Close(); closeErr != nil {
			return nil, fmt.Errorf("an error occurred closing the connection after another error: %s (first error was: %s)", closeErr, err)
		}
		return nil, err
	}

	return rc, nil
}

// Dial sets up the connection and dials it.
func (rc *RemoteConsole) Dial(address string) (err error) {
	if rc.connected {
		return ErrAlreadyConnected
	}
	if rc.settings.DialTimeout > 0 {
		rc.conn, err = net.DialTimeout("tcp", address, rc.settings.DialTimeout)
		if err != nil {
			return err
		}
	} else {
		rc.conn, err = net.Dial("tcp", address)
		if err != nil {
			return err
		}
	}
	rc.connected = true
	return nil
}

// Authenticate sends the authentication packet.
func (rc *RemoteConsole) Authenticate(password string) (err error) {
	if rc.authenticated {
		return ErrAlreadyAuthenticated
	}
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	packet := ClientPacket{
		ID:   rc.packetID,
		Type: Auth,
		Body: password,
	}

	if _, err = rc.writePacket(packet); err != nil {
		return err
	}

	_, response, err := NewServerPacketFromRemoteConsole(rc)
	if err != nil {
		if err != ErrInvalidPacketTypeRust {
			return err
		}

		_, response, err = NewServerPacketFromRemoteConsole(rc)
		if err != nil {
			return err
		}
	}

	if response == nil {
		return errors.New("nil ptr")
	}

	if response.Type == ResponseValue {
		_, response, err = NewServerPacketFromRemoteConsole(rc)
		if err != nil {
			return err
		}
	}

	if response.Type != AuthResponse {
		return ErrInvalidResponseToAuthPacket
	}

	if response.ID == -1 {
		return ErrAuthPacketFailed
	}

	if response.ID != packet.ID {
		return ErrInvalidPacketIDInResponse
	}

	rc.authenticated = true
	return nil
}

// Close forwards the Close() method from the underlying net.Conn.
func (rc *RemoteConsole) Close() (err error) {
	return rc.conn.Close()
}

// RemoteAddr returns the RemoteAddr() from the underlying net.Conn.
func (rc *RemoteConsole) RemoteAddr() (addr net.Addr) {
	return rc.conn.RemoteAddr()
}

// LocalAddr returns the LocalAddr() from the underlying net.Conn.
func (rc *RemoteConsole) LocalAddr() (addr net.Addr) {
	return rc.conn.LocalAddr()
}

// Execute runs a command on the remote console.
func (rc *RemoteConsole) Execute(command string) (response string, err error) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	if command == "" {
		return response, ErrCommandEmpty
	}

	if int32(len(command)) > PacketMaximumBodySize/4 {
		return response, ErrCommandTooLong
	}

	packet := ClientPacket{
		ID:   rc.packetID,
		Type: ExecuteCommand,
		Body: command,
	}

	if _, err := rc.writePacket(packet); err != nil {
		return response, err
	}

	_, responsePacket, err := NewServerPacketFromRemoteConsole(rc)
	if err != nil {
		if err != ErrInvalidPacketTypeRust {
			return response, err
		}

		_, responsePacket, err = NewServerPacketFromRemoteConsole(rc)
		if err != nil {
			return response, err
		}
	}

	if responsePacket.ID != packet.ID {
		return response, ErrInvalidPacketIDInResponse
	}

	return responsePacket.Body(), nil
}

func (rc *RemoteConsole) writePacket(packet ClientPacket) (n int64, err error) {
	rc.packetID += 1
	n, err = packet.WriteTo(rc)
	return n, err
}
