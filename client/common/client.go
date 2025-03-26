package common

import (
	"encoding/json"
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

type Bet struct {
	Agency    string
	FirstName string
	LastName  string
	Document  string
	BirthDate string
	Number    string
}

func NewBet(firstName string, lastName string, document string, birthdate string, number string) *Bet {
	bet := &Bet{
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		BirthDate: birthdate,
		Number:    number,
	}

	return bet
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

	return nil
}

// WaitBetConfirmation waits for the confirmation message from the server
// after sending a bet. If the client receives a shutdown signal, the
// confirmation waiting will be interrupted gracefully.
func (c *Client) WaitBetsProcessing() error {
	msg, err := NewMBPMessage("PROCESS_BETS", nil)

	if err != nil {
		return err
	}

	if err := c.socket.SendMessage(msg); err != nil {
		return err
	}

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

// PlaceBet sends the bet to the server and waits for the confirmation
// message. If the client receives a shutdown signal, the waiting will
// be interrupted gracefully.
func (c *Client) PlaceBet(bet *Bet) error {
	bet.Agency = c.config.ID
	serialized_bet, err := json.Marshal(bet)

	if err != nil {
		return err
	}

	msg, err := NewMBPMessage("PLACE_BET", serialized_bet)

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
		bets, err := br.ReadN(c.config.BatchAmount)
		log.Debugf("action: start_bets_batch | client_id: %v | bets: %v", c.config.ID, len(bets))

		if err != nil {
			return err
		}

		if bets == nil || len(bets) == 0 {
			break
		}

		total += len(bets)

		for _, bet := range bets {
			select {
				case signal := <-c.shutdown:
					log.Infof("action: graceful_shutdown | result: in_progress | client_id: %v | signal: %s", c.config.ID, signal.String())
					return nil

				default:
						log.Debugf("action: place_bets | result: in_progress | client_id: %v | bet: %v", c.config.ID, bet)

						if err := c.PlaceBet(bet); err != nil {
							log.Errorf("action: place_bets | result: fail | client_id: %v | error: %v",
								c.config.ID,
								err,
							)

							return nil
						}
			}
		}

		if err := c.WaitBetsProcessing(); err != nil {
			return err
		}

		log.Infof("action: place_bets | result: success | client_id: %v | bets batch: %v", c.config.ID, len(bets))
	}

	log.Infof("action: place_bets | result: success | client_id: %v | cantidad: %v", c.config.ID, total)

	msg, err := NewMBPMessage("END", nil)

	if err != nil {
		return err
	}
	
	if err := c.socket.SendMessage(msg); err != nil {
		return err
	}

	return nil
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
