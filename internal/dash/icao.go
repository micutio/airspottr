package dash

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	icaoListPath      = "./data/ICAOList.csv"
	airlineListPath   = "./data/Airlines.csv"
	regPrefixListPath = "./data/RegPrefixList.csv"
	hexRangeListPath  = "./data/ICAOHexRange.csv"
	milCodeFilePath   = "./data/MilICAOOperatorLookUp.csv"
	milCodeHeaderLen  = 2
)

var (
	errParseCSV  = errors.New("error parsing CSV")
	errHeaderLen = errors.New("unexpected header length")
	errParseHex  = errors.New("unable to parse hexadecimal string")
)

type icaoAircraft struct {
	Class  string
	Engine string
	Make   string
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
		manufacturer := strings.Trim(record[3], "\"")
		records[key] = icaoAircraft{class, engine, manufacturer}
	}

	return records, nil
}

type icaoOperator struct {
	Company string
	Country string
}

// getIcaoToAircraftMap returns a three-letter code to airline record mapping.
func getIcaoToAirlineMap() (map[string]icaoOperator, error) {
	// Parse the CSV file
	icaoAirlineMap, err := parseAirlineCsvToMap(airlineListPath)
	if err != nil {
		return nil, fmt.Errorf("getIcaoToAirlineMap: %w: %w", errParseCSV, err)
	}

	return icaoAirlineMap, nil
}

// parseAirlineCsvToMap reads a CSV file and parses it into a map ICAO Code -> airline record.
func parseAirlineCsvToMap(filePath string) (map[string]icaoOperator, error) {
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

	// icaoOperator Headers = Company,Country,Telephony,3Ltr
	lenIcaoAirlineHeaders := 4
	if len(headers) != lenIcaoAirlineHeaders {
		return nil, fmt.Errorf("parseAirlineCsvMap: %w", errHeaderLen)
	}

	records := make(map[string]icaoOperator)

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
		records[threeLtrCode] = icaoOperator{company, country}
	}

	return records, nil
}

type hexRange struct {
	LowerBound int64
	UpperBound int64
}

// getHexRangeToCountryMap returns a hex registration range to country mapping.
func getHexRangeToCountryMap() (map[hexRange]string, error) {
	// Parse the CSV file
	hexRangeMap, err := parseHexRangeCsvToMap(hexRangeListPath)
	if err != nil {
		return nil, fmt.Errorf("getRegPrefixMap: %w: %w", errParseCSV, err)
	}

	return hexRangeMap, nil
}

// parseAirlineCsvToMap reads a CSV file and parses it into a map regPrefix -> country.
func parseHexRangeCsvToMap(filePath string) (map[hexRange]string, error) {
	// Open the CSV file
	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		return nil, fmt.Errorf("parseHexRangeCsvToMap: failed to open file: %w", fileErr)
	}
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			fileErr = fmt.Errorf("parseHexRangeCsvToMap: error while closing file %s: %w", filePath, closeErr)
		}
	}()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Does not have a header row, so we don't need to read it first.

	records := make(map[hexRange]string)

	// Loop through the remaining records
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break // End of file
		}

		if err != nil {
			return nil, fmt.Errorf("parseHexRangeCsvToMap: failed to read record: %w", err)
		}

		lowerBound, err := strconv.ParseInt(record[0], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("parseHexRangeCsvToMap: %w: %s", errParseHex, record[0])
		}
		upperBound, err := strconv.ParseInt(record[1], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("parseHexRangeCsvToMap: %w: %s", errParseHex, record[1])
		}
		// skipping comment, record[2] is unused
		records[hexRange{lowerBound, upperBound}] = record[2]
	}

	return records, nil
}

// getRegPrefixMap returns a registration prefix to country mapping.
func getRegPrefixMap() (map[string]string, error) {
	// Parse the CSV file
	regPrefixMap, err := parseRegPrefixCsvToMap(regPrefixListPath)
	if err != nil {
		return nil, fmt.Errorf("getRegPrefixMap: %w: %w", errParseCSV, err)
	}

	return regPrefixMap, nil
}

// parseAirlineCsvToMap reads a CSV file and parses it into a map regPrefix -> country.
func parseRegPrefixCsvToMap(filePath string) (map[string]string, error) {
	// Open the CSV file
	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		return nil, fmt.Errorf("parseRegPrefixCsvToMap: failed to open file: %w", fileErr)
	}
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			fileErr = fmt.Errorf("parseRegPrefixCsvToMap: error while closing file %s: %w", filePath, closeErr)
		}
	}()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header row
	headers, headerErr := reader.Read()
	if headerErr != nil {
		return nil, fmt.Errorf("parseRegPrefixCsvToMap: failed to read header: %w", headerErr)
	}

	// regPrefix Headers = country, prefix, comment
	lenPrefixHeaders := 3
	if len(headers) != lenPrefixHeaders {
		return nil, fmt.Errorf("parseRegPrefixCsvToMap: %w", errHeaderLen)
	}

	records := make(map[string]string)

	// Loop through the remaining records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break // End of file
		}

		if err != nil {
			return nil, fmt.Errorf("parseRegPrefixCsvToMap: failed to read record: %w", err)
		}

		country := record[0]
		prefix := record[1]
		// skipping comment, record[2] is unused
		records[prefix] = country
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
