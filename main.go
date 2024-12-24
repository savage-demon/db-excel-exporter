package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/xuri/excelize/v2"
)

var (
	query          string
	output         string
	pageSize       int
	dbConfigString string
	dbType         string
	headers        string
)

func main() {
	flag.StringVar(&dbConfigString, "db-url", "", "Connection string for the database")
	flag.StringVar(&dbType, "db-type", "postgres", "Database type, e.g. 'postgres'")
	flag.StringVar(&query, "query", "", "SELECT query")
	flag.StringVar(&output, "output", "output.xlsx", "Output filename")
	flag.IntVar(&pageSize, "page-size", 1000000, "Page size")
	flag.StringVar(&headers, "headers", "", "Headers in the format 'key1=value1, key2=value2'")
	flag.Parse()

	if query == "" || dbConfigString == "" {
		flag.PrintDefaults()

		os.Exit(0)
	}

	db := getDB(dbConfigString, dbType)
	defer db.Close()

	generateExcel(db, query, output, pageSize, headers)
}

func getDB(dbConfigString string, dbType string) *sqlx.DB {
	db, err := sqlx.Connect(dbType, dbConfigString)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	return db
}

func generateExcel(db *sqlx.DB, query string, output string, pageSize int, headers string) {
	rowLimit := 1000000

	sheetNum := 1

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sw := streamWriter(f, sheetNum)

	customHeaders := parseHeaders(headers)

	offset := 0
	rowIndex := 1

	for {
		paginatedQuery := fmt.Sprintf("%s LIMIT %d OFFSET %d", query, pageSize, offset)
		rows, err := db.Queryx(paginatedQuery)
		if err != nil {
			log.Fatal("Failed to execute query:", err)
		}

		columns, err := rows.Columns()
		if err != nil {
			log.Fatal("Failed to get columns:", err)
		}

		if rowIndex > rowLimit {
			sheetNum++
			rowIndex = 1
			sw = streamWriter(f, sheetNum)
		}

		if rowIndex == 1 {
			headers := make([]interface{}, len(columns))
			for i, col := range columns {
				headers[i] = col
				if customHeader, exists := customHeaders[col]; exists {
					headers[i] = customHeader
				} else {
					headers[i] = col
				}
			}

			err = sw.SetColWidth(1, len(columns), 40)
			if err != nil {
				log.Fatal("Failed to set column width:", err)
			}

			cell, _ := excelize.CoordinatesToCellName(1, rowIndex)
			if err := sw.SetRow(cell, headers); err != nil {
				log.Fatal("Failed to write headers:", err)
			}

			rowIndex++
		}

		hasData := false

		for rows.Next() {
			hasData = true
			values, err := rows.SliceScan()
			if err != nil {
				log.Fatal("Failed to scan row:", err)
			}

			for i, v := range values {
				if v == nil {
					values[i] = ""
				}
			}

			cell, _ := excelize.CoordinatesToCellName(1, rowIndex)
			if err := sw.SetRow(cell, values); err != nil {
				log.Fatal("Failed to write row:", err)
			}
			rowIndex++
		}

		if !hasData {
			break
		}

		offset += pageSize
	}

	if err := sw.Flush(); err != nil {
		log.Fatal("Failed to flush StreamWriter:", err)
	}

	if err := f.SaveAs(output); err != nil {
		log.Fatal("Failed to save Excel file:", err)
	}
}

func streamWriter(f *excelize.File, sheetNum int) *excelize.StreamWriter {
	_, err := f.NewSheet("Sheet" + fmt.Sprint(sheetNum))
	if err != nil {
		log.Fatal("Failed to create sheet:", err)
	}

	sw, err := f.NewStreamWriter("Sheet" + fmt.Sprint(sheetNum))
	if err != nil {
		log.Fatal("Failed to create StreamWriter:", err)
	}
	return sw
}

func parseHeaders(headerStr string) map[string]string {
	headerMap := make(map[string]string)
	pairs := strings.Split(headerStr, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		splitPair := strings.Split(pair, "=")
		if len(splitPair) == 2 {
			headerMap[splitPair[0]] = splitPair[1]
		}
	}
	return headerMap
}
