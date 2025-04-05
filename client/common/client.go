package common

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	BatchAmount   int
}

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

// SetupGracefulShutdown sets up a handler for SIGINT and SIGTERM
// signals to gracefully shutdown the client
func (c *Client) SetupGracefulShutdown() {
	c.shutdown = make(chan os.Signal)
	signal.Notify(c.shutdown, os.Interrupt, syscall.SIGTERM)
}

// Connect establishes a connection with the server
// and registers the client in the server (to let the server know which
// agency is connected).
func (c *Client) Connect() error {
	c.socket = NewMBPSocket()

	if err := c.socket.Connect(c.config.ServerAddress); err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ServerAddress,
			err,
		)
	
		return err
	}

	msg, err := NewRegisterMessage(c.config.ID).GetMessage()

	if err != nil {
		return err
	}

	if err := c.socket.SendMessage(msg); err != nil {
		log.Criticalf(
			"action: register | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)

		return err
	}

	return nil
}

// WaitBetConfirmation waits for the confirmation message from the server
// after sending a bet. If the client receives a shutdown signal, the
// confirmation waiting will be interrupted gracefully.
func (c *Client) WaitBetsProcessing() error {
	select {
		case signal := <-c.shutdown:
			log.Infof("action: graceful_shutdown | result: in_progress | client_id: %v | signal: %s", c.config.ID, signal.String())

		default:
			msg, err := c.socket.ReceiveMessage()

			if err != nil {
				log.Errorf("action: apuestas_enviadas | result: fail | error: %v", err)

				return err
			}

			if msg.action != "BETS_PROCESSED" {
				err := errors.New(fmt.Sprintf("Unexpected server response: %v", msg.action))

				log.Errorf("action: apuestas_enviadas | result: fail | error: %v", err)

				return err
			}
	}

	return nil
}

func (c* Client) sendBetsBatchMessage(bets []*Bet) error {
	msg, err := NewPlaceBetMessage(bets).GetMessage()
	if err != nil {
		return err
	}

	if err := c.socket.SendMessage(msg); err != nil {
		return err
	}

	return nil
}

// PlaceBets sends bets in batches (as configured in the client) to the
// server. The bets are read from BetsReader in batches too, to be able
// to handle large amounts of bets without memory constraints.
func (c *Client) PlaceBets(br *BetsReader) error {
	total := 0

	for {
		select {
			case signal := <-c.shutdown:
				log.Infof("action: graceful_shutdown | result: in_progress | client_id: %v | signal: %s", c.config.ID, signal.String())
				return nil

			default:
				// Note: Since ReadN doesn't consider added sizes from the tranmission protocols (message action + whitespace + \n) and serialization
				// (\0 between each part), this limit is rounded down to 7800 bytes to avoid exceeding the limit by this approximation.
				bets, err := br.ReadN(c.config.BatchAmount, 7800)
				log.Debugf("action: start_bets_batch | client_id: %v | bets: %v", c.config.ID, len(bets))

				if err != nil {
					return err
				}

				// End the loop if there are no more bets to read
				if len(bets) == 0 {
					log.Infof("action: place_bets | result: success | client_id: %v | cantidad: %v", c.config.ID, total)

					msg, err := NewEndBetsMessage().GetMessage()
					if err != nil {
						return err
					}
					
					if err := c.socket.SendMessage(msg); err != nil {
						return err
					}

					return nil
				}

				if err := c.sendBetsBatchMessage(bets); err != nil {
					log.Debugf("action: place_bets | result: error | client_id: %v | error: %v", c.config.ID, err)
					return err
				}

				if err := c.WaitBetsProcessing(); err != nil {
					log.Debugf("action: wait_processing | result: error | client_id: %v | error: %v", c.config.ID, err)
					return err
				}

				total += len(bets)
				log.Infof("action: place_bets | result: success | client_id: %v | bets batch: %v", c.config.ID, len(bets))
		}
	}
}

// Disconnect closes the connection with the server
func (c *Client) Disconnect() error {
	if err := c.socket.Close(); err != nil {
		log.Errorf("action: close_connection | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)

		return err
	}

	return nil
}
