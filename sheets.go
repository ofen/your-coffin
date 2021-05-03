package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func createTempCredentials() error {
	content := os.Getenv("GOOGLE_CREDENTIALS")
	if content == "" {
		return nil
	}

	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}

	if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmpfile.Name()); err != nil {
		return err
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return err
	}

	if err := tmpfile.Close(); err != nil {
		return err
	}
	return nil
}

func newSpreadsheet(sheet string) (*spreadsheet, error) {
	sheetPair := strings.Split(sheet, ":")
	if len(sheetPair) != 2 {
		return nil, fmt.Errorf("incorrect sheet name")
	}

	err := createTempCredentials()
	if err != nil {
		return nil, err
	}

	svc, err := sheets.NewService(context.TODO(), option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return nil, err
	}

	return &spreadsheet{
		spreadsheetID: sheetPair[0],
		sheetName:     sheetPair[1],
		svc:           svc,
	}, nil
}

type spreadsheet struct {
	svc           *sheets.Service
	spreadsheetID string
	sheetName     string
}

func (s *spreadsheet) appendRow(values []interface{}) error {
	valuerange := &sheets.ValueRange{}
	valuerange.Values = append(valuerange.Values, values)
	_, err := s.svc.Spreadsheets.Values.Append(s.spreadsheetID, s.sheetName, valuerange).ValueInputOption("USER_ENTERED").Do()

	if err != nil {
		return err
	}
	return nil
}

func (s *spreadsheet) get() (*sheets.ValueRange, error) {
	return s.svc.Spreadsheets.Values.Get(s.spreadsheetID, s.sheetName).Do()
}
