package common

import (
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

func (b *Bet) GetMessage() (*MBPMessage, error) {
	msg, err := NewMBPMessage("PLACE_BET", SBDSerialize([]string{
		b.Agency,
		b.FirstName,
		b.LastName,
		b.Document,
		b.BirthDate,
		b.Number,
	}))

	if err != nil {
		return nil, err
	}

	return msg, nil
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

func (c *Client) Connect() error {
	if err := c.createClientSocket(); err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
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
func (c *Client) WaitBetConfirmation(bet *Bet) {
	select {
		case signal := <-c.shutdown:
			log.Infof("action: graceful_shutdown | result: in_progress | client_id: %v | signal: %s", c.config.ID, signal.String())

		default:
			msg, err := c.socket.ReceiveMessage()

			if err != nil {
				log.Errorf("action: apuesta_enviada | result: fail | error: %v", err)

				return
			}

			if msg.action != "STORED_BET" {
				log.Errorf("action: apuesta_enviada | result: fail | error: unexpected server response (%v)", msg.action)

				return
			}

			log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
				bet.Document,
				bet.Number,
			)
	}
}

// PlaceBet sends the bet to the server and waits for the confirmation
// message. If the client receives a shutdown signal, the waiting will
// be interrupted gracefully.
func (c *Client) PlaceBet(bet *Bet) error {
	bet.Agency = c.config.ID

	msg, err := bet.GetMessage()

	if err != nil {
		return err
	}

	if err := c.socket.SendMessage(msg); err != nil {
		return err
	}

	c.WaitBetConfirmation(bet)

	return nil
}

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
