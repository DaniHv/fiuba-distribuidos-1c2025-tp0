package common

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"strings"
)

// MBPMessage (Message-based Protocol Message) is the data representation of MBPSocket message
// containing an action and data.

// The action is a string (utf8) representing the action to be taken by the counterpart,
// cannot contain spaces or newlines, for example: EXECUTE_SOMETHING. Can be thought as
// the combination of VERB + URL in an HTTP Request.

// The data is a bytes stream containing the data to be sent or received. MBPSocket doesn't
// enforce any restrictions on the data, it could be a plain text, json or any
// other binary data such as protobufs. Can be though as the body of an HTTP Request.
type MBPMessage struct {
	action string
	data   []byte
}

func NewMBPMessage(action string, data []byte) (*MBPMessage, error) {
	if strings.ContainsAny(action, " \n") {
		return nil, errors.New("action cannot contain spaces or new line character")
	}

	return &MBPMessage{
		action: action,
		data:   data,
	}, nil
}

func NewMBPMessageFromBytes(data []byte) (*MBPMessage, error) {
	parts := bytes.SplitN(data, []byte(" "), 2)

	if len(parts) != 2 {
		return nil, errors.New("invalid message format")
	}

	return NewMBPMessage(string(parts[0]), parts[1])
}

// ToBytes converts the MBPMessage to a byte stream to be sent over the network.
// Doesn't include message separator (\n), which is responsibility of the MBPSocket.SendMessage method.
func (message *MBPMessage) ToBytes() []byte {
	return bytes.Join([][]byte{[]byte(message.action), []byte(" "), message.data}, nil)
}

// MBPSocket (Message-based Protocol Socket) is a TCP-Based, full-duplex and message-oriented protocol.
//
// Messages received and sent are represented by MBPMessage objects,
// containing an action and data, both strings.
//
// Note: "listen" method is not implemented since it won't be required, since this will be used
// as client-only.
type MBPSocket struct {
	tcpSocket net.Conn
}

func NewMBPSocket() *MBPSocket {
	return &MBPSocket{}
}

// Connect establishes a connection with an MBP Server
// given address in the form of "host:port".
func (mbpSocket *MBPSocket) Connect(addr string) error {
	socket, err := net.Dial("tcp", addr)
	mbpSocket.tcpSocket = socket

	return err
}

// SendMessage sends a message to the connected MBP Server.
func (socket *MBPSocket) SendMessage(message *MBPMessage) error {
	if socket.tcpSocket == nil {
		return errors.New("socket is not connected")
	}

	writter := bufio.NewWriter(socket.tcpSocket)

	message_bytes := bytes.Join([][]byte{message.ToBytes(), []byte("\n")}, nil)

	writter.Write(message_bytes)

	// Flush garantees (or fails) that the message is sent
	// entirely (preventing short writes), for that reason
	// Write error is not being handled.
	err := writter.Flush()

	if err != nil {
		return err
	}

	return nil
}

// ReceiveMessage receives a single message (MBPMessage) from the connected MBP Server.
func (socket *MBPSocket) ReceiveMessage() (*MBPMessage, error) {
	reader := bufio.NewReader(socket.tcpSocket)

	bytes, err := reader.ReadBytes('\n')

	if err != nil {
		return nil, err
	}

	message, err := NewMBPMessageFromBytes(bytes[:len(bytes)-1])

	if err != nil {
		return nil, err
	}

	return message, nil
}

// Close closes the connection with the MBP Server.
func (socket *MBPSocket) Close() error {
	if err := socket.tcpSocket.Close(); err != nil {
		return err
	}

	socket.tcpSocket = nil

	return nil
}