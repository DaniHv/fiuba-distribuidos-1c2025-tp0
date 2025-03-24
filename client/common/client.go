package common

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config   ClientConfig
	socket   *MBPSocket
	shutdown chan os.Signal
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	c.socket = NewMBPSocket()

	return c.socket.Connect(c.config.ServerAddress)
}

// SetupGracefulShutdown sets up a handler for SIGINT and SIGTERM
// signals to gracefully shutdown the client
func (c *Client) SetupGracefulShutdown() {
	c.shutdown = make(chan os.Signal)
	signal.Notify(c.shutdown, os.Interrupt, syscall.SIGTERM)
}

// waitLoopOrShutdown waits for the loop period (from the LoopPeriod config)
// or for a shutdown signal to be received.
// 
// Returns true if the loop should continue, false otherwise.
func (c *Client) waitLoopOrShutdown() bool {
	select {
		case <-time.After(c.config.LoopPeriod):
			return true

		case signal := <-c.shutdown:
			log.Infof("action: graceful_shutdown | result: in_progress | client_id: %v | signal: %s", c.config.ID, signal.String())
			return false
	}
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		if err := c.createClientSocket(); err != nil {
			log.Criticalf(
				"action: connect | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
	
			return
		}

		msg, err := NewMBPMessage("MESSAGE", []byte(fmt.Sprintf("[CLIENT %v] Message N°%v", c.config.ID,msgID)))

		if err != nil {
			log.Errorf("action: create_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
	
			return
		}

		if err := c.socket.SendMessage(msg); err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)

			return
		}

		response, err := c.socket.ReceiveMessage()

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			string(response.data),
		)

		if err := c.socket.Close(); err != nil {
			log.Errorf("action: close_connection | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
		}

		if shouldContinue := c.waitLoopOrShutdown(); !shouldContinue {
			log.Infof("action: graceful_shutdown | result: success | client_id: %v", c.config.ID)

			return
		}
	}

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
