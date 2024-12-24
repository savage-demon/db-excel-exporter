package main

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestGenerateExcel(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error initializing mock database: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	rows := sqlmock.NewRows([]string{
		"id", "name", "test",
	}).AddRow("1", "name1", "test1")

	mock.ExpectQuery("SELECT (.+) FROM table LIMIT 100 OFFSET 0").
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT (.+) FROM table LIMIT 100 OFFSET 100").
		WillReturnRows(sqlmock.NewRows([]string{}))

	output := "test_output.xlsx"
	generateExcel(sqlxDB, "SELECT * FROM table", output, 100, "")

	f, err := excelize.OpenFile(output)
	if err != nil {
		t.Fatalf("Error opening Excel file: %v", err)
	}
	defer f.Close()

	cellValue, err := f.GetCellValue("Sheet1", "A2")
	if err != nil {
		t.Fatalf("Error reading cell: %v", err)
	}
	assert.Equal(t, "1", cellValue, "Expected cell value does not match")
}

func TestParseHeaders(t *testing.T) {
	headerStr := "id=Идентификатор, document_reference=Код документа"
	headers := parseHeaders(headerStr)

	expected := map[string]string{
		"id":                 "Идентификатор",
		"document_reference": "Код документа",
	}

	assert.Equal(t, expected, headers, "Parsed headers do not match expected values")
}

func TestParseHeadersEmpty(t *testing.T) {
	headerStr := ""
	headers := parseHeaders(headerStr)

	expected := map[string]string{}

	assert.Equal(t, expected, headers, "Parsed headers from empty string should result in empty map")
}
