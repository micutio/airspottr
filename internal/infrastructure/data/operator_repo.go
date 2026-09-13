package data

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

const (
	airlineListPath  = "./data/Airlines.csv"
	milCodeFilePath  = "./data/MilICAOOperatorLookUp.csv"
	milCodeHeaderLen = 2
)

var (
	errParseIcaoAirlineMap = errors.New("failed to parse ICAO to airline map")
	errParseMilCodeMap     = errors.New("failed to parse mil code to operator map")
)

type OperatorRepo struct {
	icaoToOperator    map[string]ref.IcaoOperator
	milCodeToOperator map[string]string
}

func NewOperatorRepo() (*OperatorRepo, error) {
	const initError = "NewOperatorRepo: %w caused by %w"
	icaoToOperatorMap, aircraftErr := getIcaoToAirlineMap()
	if aircraftErr != nil {
		return nil, fmt.Errorf(initError, errParseIcaoAirlineMap, aircraftErr)
	}

	milCodeToOperatorMap, milCodeErr := getMilCodeToOperatorMap()
	if milCodeErr != nil {
		return nil, fmt.Errorf(initError, errParseMilCodeMap, milCodeErr)
	}

	repo := OperatorRepo{
		icaoToOperator:    icaoToOperatorMap,
		milCodeToOperator: milCodeToOperatorMap,
	}

	return &repo, nil
}

/// Interface methods ////////////////////////////////////////////////////////

// GetOperatorByIcao implements the OperatorRepo interface of the same name.
func (acr *OperatorRepo) GetOperatorByIcao(icaoCode string) (ref.IcaoOperator, bool) {
	operator, exists := acr.icaoToOperator[icaoCode]
	return operator, exists
}

// GetOperatorByMilCode implements the OperatorRepo interface of the same name.
func (acr *OperatorRepo) GetOperatorByMilCode(milCode string) (string, bool) {
	operator, exists := acr.milCodeToOperator[milCode]
	return operator, exists
}

//////////////////////////////////////////////////////////////////////////////

// getIcaoToAirlineMap returns a three-letter code to airline record mapping.
func getIcaoToAirlineMap() (map[string]ref.IcaoOperator, error) {
	// Parse the CSV file
	icaoAirlineMap, err := parseAirlineCsvToMap(airlineListPath)
	if err != nil {
		return nil, fmt.Errorf("getIcaoToAirlineMap: %w: %w", errParseCSV, err)
	}

	return icaoAirlineMap, nil
}

// parseAirlineCsvToMap reads a CSV file and parses it into a map ICAO Code -> airline record.
func parseAirlineCsvToMap(filePath string) (map[string]ref.IcaoOperator, error) {
	// Open the CSV file
	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		return nil, fmt.Errorf("parseAirlineCsvToMap: failed to open file: %w", fileErr)
	}
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			fileErr = fmt.Errorf("parseAirlineCsvToMap: error while closing file %s: %w", filePath, closeErr)
		}
	}()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header row
	headers, headerErr := reader.Read()
	if headerErr != nil {
		return nil, fmt.Errorf("parseAirlineCsvToMap: failed to read header: %w", headerErr)
	}

	// icaoOperator Headers = Company,country,Telephony,3Ltr
	lenIcaoAirlineHeaders := 4
	if len(headers) != lenIcaoAirlineHeaders {
		return nil, fmt.Errorf("parseAirlineCsvMap: %w", errHeaderLen)
	}

	records := make(map[string]ref.IcaoOperator)

	// Loop through the remaining records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break // End of file
		}

		if err != nil {
			return nil, fmt.Errorf("parseAirlineCsvToMap: failed to read record: %w", err)
		}

		company := record[0]
		country := record[1]
		// skipping telephony, record[2] is unused
		threeLtrCode := record[3][0:3]
		records[threeLtrCode] = ref.IcaoOperator{
			Company: company,
			Country: country,
		}
	}

	return records, nil
}

// getMilCodeToOperatorMap returns a military code to operator mapping.
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
