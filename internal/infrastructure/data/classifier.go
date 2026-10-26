package data

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log" //nolint:depguard // Don't feel like using slog
	"strconv"
	"strings"

	rawdata "github.com/micutio/airspottr/data"
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

const (
	icaoListPath            = "ICAOList.csv"
	airlineListPath         = "Airlines.csv"
	milCodeFilePath         = "MilICAOOperatorLookUp.csv"
	regPrefixListPath       = "RegPrefixList.csv"
	hexRangeListPath        = "ICAOHexRange.csv"
	icaoAircraftHeaderCount = 4
	icaoAirlineHeaderCount  = 4
	milCodeHeaderCount      = 2
	regPrefixHeaderCount    = 3
	airlineCodeLen          = 3
)

var (
	errHeaderLen = errors.New("unexpected header length")
	errParseHex  = errors.New("unable to parse hexadecimal string")
)

// Classifier loads CSV reference datasets and implements obs.Classifier.
type Classifier struct {
	icaoToAircraft     map[string]ref.IcaoAircraftSpec
	icaoToOperator     map[string]ref.IcaoOperator
	milCodeToOperator  map[string]string
	hexRangeToCountry  map[ref.HexRange]string
	regPrefixToCountry map[string]string
	errOut             *log.Logger
}

var _ obs.Classifier = (*Classifier)(nil)

func NewClassifier(stderr io.Writer) (*Classifier, error) {
	if stderr == nil {
		stderr = io.Discard
	}
	classifier := &Classifier{
		icaoToAircraft:     make(map[string]ref.IcaoAircraftSpec),
		icaoToOperator:     make(map[string]ref.IcaoOperator),
		milCodeToOperator:  make(map[string]string),
		hexRangeToCountry:  make(map[ref.HexRange]string),
		regPrefixToCountry: make(map[string]string),
		errOut:             log.New(stderr, "classifier ", log.LstdFlags),
	}

	if err := readCSV(icaoListPath, icaoAircraftHeaderCount, func(r []string) error {
		classifier.icaoToAircraft[r[0]] = ref.IcaoAircraftSpec{
			Class:  r[1],
			Engine: r[2],
			Make:   strings.Trim(r[3], "\""),
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("load aircraft types: %w", err)
	}

	if err := readCSV(airlineListPath, icaoAirlineHeaderCount, func(r []string) error {
		classifier.icaoToOperator[r[3][:airlineCodeLen]] = ref.IcaoOperator{
			Company: r[0],
			Country: r[1],
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("load airlines: %w", err)
	}

	if err := readCSV(milCodeFilePath, milCodeHeaderCount, func(r []string) error {
		if r[1] != "" {
			classifier.milCodeToOperator[r[1]] = r[0]
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("load mil operators: %w", err)
	}

	if err := readCSV(hexRangeListPath, 0, func(hexRange []string) error {
		lower, lowerErr := strconv.ParseInt(hexRange[0], 16, 64)
		if lowerErr != nil {
			return fmt.Errorf("%w: %s", errParseHex, hexRange[0])
		}
		upper, upperErr := strconv.ParseInt(hexRange[1], 16, 64)
		if upperErr != nil {
			return fmt.Errorf("%w: %s", errParseHex, hexRange[1])
		}
		classifier.hexRangeToCountry[ref.HexRange{LowerBound: lower, UpperBound: upper}] = hexRange[2]
		return nil
	}); err != nil {
		return nil, fmt.Errorf("load hex ranges: %w", err)
	}

	if err := readCSV(regPrefixListPath, regPrefixHeaderCount, func(r []string) error {
		classifier.regPrefixToCountry[r[1]] = r[0]
		return nil
	}); err != nil {
		return nil, fmt.Errorf("load registration prefixes: %w", err)
	}

	return classifier, nil
}

func (c *Classifier) AircraftMake(icaoType string) (string, bool) {
	spec, exists := c.icaoToAircraft[icaoType]
	if !exists {
		return "", false
	}
	return spec.Make, true
}

func (c *Classifier) OperatorByIcao(code string) (string, string, bool) {
	op, exists := c.icaoToOperator[code]
	if !exists {
		return "", "", false
	}
	return op.Company, op.Country, true
}

func (c *Classifier) OperatorByMil(code string) (string, bool) {
	op, exists := c.milCodeToOperator[code]
	return op, exists
}

func (c *Classifier) CountryByHex(hexCode string) (string, bool) {
	hexAsInt, err := strconv.ParseInt(hexCode, 16, 64)
	if err != nil {
		if c.errOut != nil {
			c.errOut.Printf("warning: invalid hex code: %v", err)
		}
		return "", false
	}
	for key, value := range c.hexRangeToCountry {
		if hexAsInt > key.LowerBound && hexAsInt < key.UpperBound {
			return value, true
		}
	}
	return ref.CountryUnknown, true
}

func (c *Classifier) CountryByRegistration(registration string) (string, bool) {
	for key, value := range c.regPrefixToCountry {
		if strings.Contains(registration, key) {
			return value, true
		}
	}
	return "", false
}

func readCSV(filePath string, expectedHeaders int, onRow func(record []string) error) error {
	file, err := rawdata.Files.Open(filePath)
	if err != nil {
		return fmt.Errorf("open %s: %w", filePath, err)
	}
	defer func() {
		_ = file.Close()
	}()

	reader := csv.NewReader(file)
	if expectedHeaders > 0 {
		headers, headerErr := reader.Read()
		if headerErr != nil {
			return fmt.Errorf("read header %s: %w", filePath, headerErr)
		}
		if len(headers) != expectedHeaders {
			return fmt.Errorf("%s: %w", filePath, errHeaderLen)
		}
	}

	for {
		record, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("read record %s: %w", filePath, readErr)
		}
		if rowErr := onRow(record); rowErr != nil {
			return rowErr
		}
	}
}
