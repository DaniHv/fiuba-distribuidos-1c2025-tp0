package common

import (
	"encoding/json"
	"fmt"
)

/// Client->Server messages

type PlaceBetMessage struct {
	bet *Bet
}

func NewPlaceBetMessage(bet *Bet) *PlaceBetMessage {
	msg := &PlaceBetMessage{
		bet: bet,
	}

	return msg
}

func (m *PlaceBetMessage) GetMessage() (*MBPMessage, error) {
	data, err := json.Marshal(m.bet)
	if err != nil {
		return nil, err
	}

	msg, err := NewMBPMessage("PLACE_BET", data)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

type RegisterMessage struct {
	ID string
}

func NewRegisterMessage(agencyID string) *RegisterMessage {
	msg := &RegisterMessage{
		ID: agencyID,
	}

	return msg
}

func (m *RegisterMessage) GetMessage() (*MBPMessage, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg, err := NewMBPMessage("REGISTER", data)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

type ProcessBetsMessage struct{}

func NewProcessBetsMessage() *ProcessBetsMessage {
	return &ProcessBetsMessage{}
}

func (m *ProcessBetsMessage) GetMessage() (*MBPMessage, error) {
	msg, err := NewMBPMessage("PROCESS_BETS", nil)

	if err != nil {
		return nil, err
	}

	return msg, nil
}

type EndBetsMessage struct{}

func NewEndBetsMessage() *EndBetsMessage {
	return &EndBetsMessage{}
}

func (m *EndBetsMessage) GetMessage() (*MBPMessage, error) {
	msg, err := NewMBPMessage("END", nil)

	if err != nil {
		return nil, err
	}

	return msg, nil
}

/// Server->Client messages

type WinnersMessage struct {
	Winners int
}

func NewWinnersMessage(msg *MBPMessage) (*WinnersMessage, error) {
	if msg.action != "WINNERS" {
		return nil, fmt.Errorf("unexpected winner message action %v", msg.action)
	}

	m := &WinnersMessage{}

	json.Unmarshal(msg.data, &m)

	return m, nil
}