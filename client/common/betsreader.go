package common

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

type Bet struct {
	FirstName string
	LastName  string
	Document  string
	BirthDate string
	Number    string
}

func NewBet(firstName string, lastName string, document string, birthDate string, number string) *Bet {
	return &Bet{
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		BirthDate: birthDate,
		Number: number,
	}
}

type BetsReader struct {
	csvReader *csv.Reader
}

func NewBetsReader() (*BetsReader, error) {
	file, err := os.Open("bets.csv")

	if err != nil {
		return nil, errors.Wrapf(err, "Could not open bets.csv file")
	}

	reader := csv.NewReader(file)

	return &BetsReader{csvReader: reader}, nil
}

// Read reads a single bet from the file, returning it or an error.
// If the end of the file is reached, it returns nil, nil.
func (br *BetsReader) Read() (*Bet, error) {
	bet, err := br.csvReader.Read(); 

	if err == io.EOF {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrapf(err, "Could not read bet")
	}

	if len(bet) != 5 {
		return nil, errors.New(fmt.Sprintf("Invalid number of fields in bet %v", bet))
	}

	return NewBet(bet[0], bet[1], bet[2], bet[3], bet[4]), nil
}

func (br *BetsReader) ReadN(n int) ([]*Bet, error) {
	bets := make([]*Bet, 0)

	for i := 0; i < n; i++ {
		bet, err := br.Read()

		if err != nil {
			return nil, errors.Wrapf(err, "Could not read bet %d", i)
		}

		if bet == nil {
			break
		}

		bets = append(bets, bet)
	}

	return bets, nil
}
