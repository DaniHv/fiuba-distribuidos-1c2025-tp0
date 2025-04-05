package common

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