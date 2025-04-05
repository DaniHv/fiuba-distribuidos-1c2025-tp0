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

func (b *Bet) Size() int {
	return len(b.FirstName) + 
		len(b.LastName) +
		len(b.Document) +
		len(b.BirthDate) +
		len(b.Number)
}

type BetsReader struct {
	file *os.File
	csvReader *csv.Reader

  // If a bet is read from the file using ReadN, but cannot
	// be returned because it exceeds the maxSize, it is stored
	// in this bufferedBet. The next call to Read will return
	// this bufferedBet.
	bufferedBet *Bet
}

func NewBetsReader() (*BetsReader, error) {
	file, err := os.Open("bets.csv")

	if err != nil {
		return nil, errors.Wrapf(err, "Could not open bets.csv file")
	}

	reader := csv.NewReader(file)

	return &BetsReader{file: file, csvReader: reader}, nil
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

// Reads N bets from the file until EOF
func (br *BetsReader) ReadN(n int, maxSize int) ([]*Bet, error) {
	bets := make([]*Bet, 0)
	accumulatedSize := 0

	if (br.bufferedBet != nil) {
		bets = append(bets, br.bufferedBet)
		br.bufferedBet = nil
		accumulatedSize += br.bufferedBet.Size()
	}

	for i := 0; i < n; i++ {
		// If the accumulated size of the bets exceeds the maxSize,
		// we stop reading more bets and return the ones we have.
		// The last bet read is stored in bufferedBet, so it can
		// be returned in the next call to Read.
		if accumulatedSize >= maxSize {
			br.bufferedBet = bets[len(bets)-1]
			bets = bets[:len(bets)-1]
			break
		}

		bet, err := br.Read()

		if err != nil {
			return nil, errors.Wrapf(err, "Could not read bet %d", i)
		}

		if bet == nil  {
			break
		}

		bets = append(bets, bet)
		accumulatedSize += bet.Size()
	}

	return bets, nil
}

func (br *BetsReader) Close() error {
	return br.file.Close()
}