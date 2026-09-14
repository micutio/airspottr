package reference

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	icaoListPath = "./data/ICAOList.csv"

	// CountryUnknown is what we use for aircraft with a type that's either empty or can't be found.
	CountryUnknown = "unknown"
)

var (
	errParseCSV  = errors.New("error parsing CSV")
	errHeaderLen = errors.New("unexpected header length")
)

type IcaoAircraftSpec struct {
	Class  string
	Engine string
	Make   string
}

// GetIcaoToAircraftMap returns an ICAO id to aircraft record mapping.
func GetIcaoToAircraftMap() (map[string]IcaoAircraftSpec, error) {
	// Parse the CSV file
	icaoAircraftMap, err := parseIcaoCsvToMap(icaoListPath)
	if err != nil {
		return nil, fmt.Errorf("getIcaoToAircraftMap: %w: %w", errParseCSV, err)
	}

	return icaoAircraftMap, nil
}

// parseIcaoCsvToMap reads a CSV file and parses it into a map ICAO -> aircraft spec.
func parseIcaoCsvToMap(filePath string) (map[string]IcaoAircraftSpec, error) {
	// Open the CSV file
	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		return nil, fmt.Errorf("parseIcaoCsvToMap: failed to open file: %w", fileErr)
	}
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			fileErr = fmt.Errorf("parseIcaoCsvToMap: error while closing file %s: %w", filePath, closeErr)
		}
	}()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header row
	headers, headerErr := reader.Read()
	if headerErr != nil {
		return nil, fmt.Errorf("parseIcaoCsvToMap: failed to read header: %w", headerErr)
	}

	// icaoAircraftHeaders := [...]string{
	//	"aircraft TypeDesignator",
	//	"Class",
	//	"Number+Engine Type",
	//	"\"MANUFACTURER, Model\"",
	// }
	lenIcaoAircraftHeaders := 4
	if len(headers) != lenIcaoAircraftHeaders {
		return nil, fmt.Errorf("parseIcaoCsvToMap: %w", errHeaderLen)
	}

	records := make(map[string]IcaoAircraftSpec)

	// Loop through the remaining records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break // End of file
		}

		if err != nil {
			return nil, fmt.Errorf("parseIcaoCsvToMap: failed to read record: %w", err)
		}

		key := record[0]
		class := record[1]
		engine := record[2]
		manufacturer := strings.Trim(record[3], "\"")
		records[key] = IcaoAircraftSpec{class, engine, manufacturer}
	}

	return records, nil
}

type IcaoOperator struct {
	Company string
	Country string
}

type HexRange struct {
	LowerBound int64
	UpperBound int64
}
