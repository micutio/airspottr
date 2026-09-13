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
	icaoListPath     = "./data/ICAOList.csv"
	airlineListPath  = "./data/Airlines.csv"
	milCodeFilePath  = "./data/MilICAOOperatorLookUp.csv"
	milCodeHeaderLen = 2

	// CountryUnknown is what we use for aircraft with a type that's either empty or can't be found.
	CountryUnknown = "unknown"
)

var (
	errParseCSV  = errors.New("error parsing CSV")
	errHeaderLen = errors.New("unexpected header length")
)

type IcaoAircraft struct {
	Class  string
	Engine string
	Make   string
}

// GetIcaoToAircraftMap returns an ICAO id to aircraft record mapping.
func GetIcaoToAircraftMap() (map[string]IcaoAircraft, error) {
	// Parse the CSV file
	icaoAircraftMap, err := parseIcaoCsvToMap(icaoListPath)
	if err != nil {
		return nil, fmt.Errorf("getIcaoToAircraftMap: %w: %w", errParseCSV, err)
	}

	return icaoAircraftMap, nil
}

// parseIcaoCsvToMap reads a CSV file and parses it into a map ICAO -> aircraft spec.
func parseIcaoCsvToMap(filePath string) (map[string]IcaoAircraft, error) {
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

	records := make(map[string]IcaoAircraft)

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
		records[key] = IcaoAircraft{class, engine, manufacturer}
	}

	return records, nil
}

type IcaoOperator struct {
	Company string
	Country string
}

// GetIcaoToAirlineMap returns a three-letter code to airline record mapping.
func GetIcaoToAirlineMap() (map[string]IcaoOperator, error) {
	// Parse the CSV file
	icaoAirlineMap, err := parseAirlineCsvToMap(airlineListPath)
	if err != nil {
		return nil, fmt.Errorf("getIcaoToAirlineMap: %w: %w", errParseCSV, err)
	}

	return icaoAirlineMap, nil
}

// parseAirlineCsvToMap reads a CSV file and parses it into a map ICAO Code -> airline record.
func parseAirlineCsvToMap(filePath string) (map[string]IcaoOperator, error) {
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

	records := make(map[string]IcaoOperator)

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
		records[threeLtrCode] = IcaoOperator{company, country}
	}

	return records, nil
}

type HexRange struct {
	LowerBound int64
	UpperBound int64
}

// GetMilCodeToOperatorMap returns a military code to operator mapping.
func GetMilCodeToOperatorMap() (map[string]string, error) {
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
