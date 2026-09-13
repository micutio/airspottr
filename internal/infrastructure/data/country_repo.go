package data

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log" //nolint:depguard // Don't feel like using slog
	"os"
	"strconv"
	"strings"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

const (
	regPrefixListPath = "./data/RegPrefixList.csv"
	hexRangeListPath  = "./data/ICAOHexRange.csv"
)

var (
	errParseCSV                  = errors.New("error parsing CSV")
	errHeaderLen                 = errors.New("unexpected header length")
	errParseHex                  = errors.New("unable to parse hexadecimal string")
	errParseRegToCountryMap      = errors.New("failed to parse reg-prefix to country map")
	errParseHexRangeToCountryMap = errors.New("failed to parse hex-range to country map")
)

type CountryRepo struct {
	hexRangeToCountry  map[ref.HexRange]string
	regPrefixToCountry map[string]string
	errOut             log.Logger
}

func NewCountryRepo(stderr *io.Writer) (*CountryRepo, error) {
	const initError = "NewCountryRepo: %w caused by %w"
	hexRangeToCountryMap, hexRangeErr := getHexRangeToCountryMap()
	if hexRangeErr != nil {
		return nil, fmt.Errorf(initError, errParseHexRangeToCountryMap, hexRangeErr)
	}

	regPrefixToCountryMap, regPrefixErr := getRegPrefixMap()
	if regPrefixErr != nil {
		return nil, fmt.Errorf(initError, errParseRegToCountryMap, regPrefixErr)
	}

	repo := CountryRepo{
		hexRangeToCountry:  hexRangeToCountryMap,
		regPrefixToCountry: regPrefixToCountryMap,
		errOut:             *log.New(*stderr, "CountryRepository ", log.LstdFlags),
	}

	return &repo, nil
}

// TODO: Dedicated type for hex code.

/// Interface Methods ////////////////////////////////////////////////////////

// GetCountryByHexCode implements the CountryRepo interface of the same name.
func (cr *CountryRepo) GetCountryByHexCode(hexCode string) string {
	hexAsInt, err := strconv.ParseInt(hexCode, 16, 64)
	if err != nil {
		cr.errOut.Printf("unable to convert hex to int: %s\n", hexCode)
		return ref.CountryUnknown
	}
	for key, value := range cr.hexRangeToCountry {
		if hexAsInt > key.LowerBound && hexAsInt < key.UpperBound {
			return value
		}
	}
	return ref.CountryUnknown
}

// GetCountryByRegistration implements the CountryRepo interface of the same name.
func (cr *CountryRepo) GetCountryByRegistration(registration string) (string, bool) {
	for key, value := range cr.regPrefixToCountry {
		if strings.Contains(registration, key) {
			return value, true
		}
	}

	return "", false
}

//////////////////////////////////////////////////////////////////////////////

// getHexRangeToCountryMap returns a hex registration range to country mapping.
func getHexRangeToCountryMap() (map[ref.HexRange]string, error) {
	// Parse the CSV file
	hexRangeMap, err := parseHexRangeCsvToMap(hexRangeListPath)
	if err != nil {
		return nil, fmt.Errorf("getRegPrefixMap: %w: %w", errParseCSV, err)
	}

	return hexRangeMap, nil
}

// parseAirlineCsvToMap reads a CSV file and parses it into a map regPrefix -> country.
func parseHexRangeCsvToMap(filePath string) (map[ref.HexRange]string, error) {
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

	records := make(map[ref.HexRange]string)

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
		hexRange := ref.HexRange{
			LowerBound: lowerBound,
			UpperBound: upperBound,
		}
		records[hexRange] = record[2]
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
