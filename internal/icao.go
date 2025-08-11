package internal

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	icaoListPath     = "./data/ICAOList.csv"
	milCodeFilePath  = "./data/MilICAOOperatorLookUp.csv"
	milCodeHeaderLen = 2
)

var (
	errParseCSV  = errors.New("error parsing CSV")
	errHeaderLen = errors.New("unexpected header length")
)

type icaoAircraft struct {
	Class     string
	Engine    string
	ModelCode string
}

// getIcaoToAircraftMap returns an ICAO id to aircraft record mapping.
func getIcaoToAircraftMap() (map[string]icaoAircraft, error) {

	// Parse the CSV file
	icaoAircraftMap, err := parseIcaoCsvToMap(icaoListPath)
	if err != nil {
		return nil, fmt.Errorf("getIcaoToAircraftMap: %w: %w", errParseCSV, err)
	}

	return icaoAircraftMap, nil
}

// parseIcaoCsvToMap reads a CSV file and parses it into a map ICAO -> aircraft spec.
func parseIcaoCsvToMap(filePath string) (map[string]icaoAircraft, error) {
	// Open the CSV file
	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		return nil, fmt.Errorf("parseIcaoToCsvMap: failed to open file: %w", fileErr)
	}
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			fileErr = fmt.Errorf("parseIcaoToCsvMap: error while closing file %s: %w", filePath, closeErr)
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
		return nil, fmt.Errorf("parseIcaoToCsvMap: %w", errHeaderLen)
	}

	records := make(map[string]icaoAircraft)

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
		manufacturer := record[3]
		records[key] = icaoAircraft{class, engine, manufacturer}
	}

	return records, nil
}

// getMilCodeToOperatorMap returns a militar code to operator mapping.
func getMilCodeToOperatorMap() (map[string]string, error) {

	// Parse the CSV file
	icaoAircraftMap, err := parseMilCodeToMap(milCodeFilePath)
	if err != nil {
		return nil, fmt.Errorf("milCodeFilePath: %w", err)
	}

	return icaoAircraftMap, nil
}

// parseMilCodeToMap reads a CSV file and parses it into a map code -> military operator.
func parseMilCodeToMap(filePath string) (map[string]string, error) {
	// Open the CSV file
	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		return nil, fmt.Errorf("parseMilCodeToMap: failed to open file: %w", fileErr)
	}
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			fileErr = fmt.Errorf("parseMilCodeToCsvMap: error while closing file %s: %w", filePath, closeErr)
		}
	}()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header row
	headers, headerErr := reader.Read()
	if headerErr != nil {
		return nil, fmt.Errorf("parseMilCodeToMap: failed to read headers: %w", headerErr)
	}

	if len(headers) != milCodeHeaderLen {
		return nil, fmt.Errorf("parseMilCodeToMap: %w", errHeaderLen)
	}

	records := make(map[string]string)

	// Loop through the remaining records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break // End of file
		}

		if err != nil {
			return nil, fmt.Errorf("parseMilCodeToMap: failed to read record: %w", err)
		}

		key := record[1]

		if len(key) == 0 {
			continue
		}

		militaryOperator := record[0]
		records[key] = militaryOperator
	}

	return records, nil
}
