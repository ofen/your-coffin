package googlesheets

import (
	"context"
	"os"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func newService(ctx context.Context) (*sheets.Service, error) {
	content := []byte(os.Getenv("GOOGLE_CREDENTIALS"))

	return sheets.NewService(ctx, option.WithCredentialsJSON(content))
}

func New(sheetname string) *Sheet {
	parts := strings.Split(sheetname, ":")
	if len(parts) != 2 {
		panic("should be in form <id>:<sheetname>")
	}

	return &Sheet{id: parts[0], name: parts[1]}
}

type Sheet struct {
	id   string
	name string
}

func (s *Sheet) AppendRow(row []interface{}) error {
	svc, err := newService(context.Background())
	if err != nil {
		return err
	}

	vr := &sheets.ValueRange{Values: [][]interface{}{row}}
	_, err = svc.Spreadsheets.Values.Append(s.id, s.name, vr).ValueInputOption("RAW").Do()

	return err
}

func (s *Sheet) Rows() (*sheets.ValueRange, error) {
	svc, err := newService(context.Background())
	if err != nil {
		return nil, err
	}

	return svc.Spreadsheets.Values.Get(s.id, s.name).Do()
}

func (s *Sheet) LastRow() (row []interface{}, err error) {
	values, err := s.Rows()
	if err != nil {
		return nil, err
	}

	if len(values.Values) < 1 {
		return []interface{}{}, nil
	}

	return values.Values[len(values.Values)-1], nil
}
