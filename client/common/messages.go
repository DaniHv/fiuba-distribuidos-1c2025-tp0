package common

import (
	"fmt"
	"strconv"
)

type PlaceBetsMessage struct {
	bets []*Bet
}

func NewPlaceBetMessage(bets []*Bet) *PlaceBetsMessage {
	msg := &PlaceBetsMessage{
		bets: bets,
	}

	return msg
}

func (m *PlaceBetsMessage) GetMessage() (*MBPMessage, error) {
	var serializedBets [][]string = make([][]string, 0)

	for _, bet := range m.bets {
		serializedBets = append(serializedBets, []string{bet.FirstName, bet.LastName, bet.Document, bet.BirthDate, bet.Number})
	}
	
	msg, err := NewMBPMessage("PLACE_BETS", SBDSerializeArray(serializedBets))

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
	msg, err := NewMBPMessage("REGISTER", SBDSerialize([]string{m.ID}))
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
	Winners map[string]int
}

func NewWinnersMessage(msg *MBPMessage) (*WinnersMessage, error) {
	if msg.action != "WINNERS" {
		return nil, fmt.Errorf("unexpected winner message action %v", msg.action)
	}

	winnersArr, err := SBDDeserializeArray(msg.data, 2)

	if err != nil {
		return nil, err
	}

	winners := make(map[string]int)

	for _, winnerParts := range winnersArr {
		qty, err := strconv.Atoi(winnerParts[1])

		if err != nil {
			return nil, fmt.Errorf("unexpected winner message value %v", err)
		}

		winners[winnerParts[0]] = qty
	}

	return &WinnersMessage{ Winners: winners }, nil
}

func (m * WinnersMessage) GetTotal() int {
	total := 0

	for _, v := range m.Winners {
		total += v
	}

	return total
}