package data

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

const icaoListPath = "./data/ICAOList.csv"

var errParseIcaoAircraftMap = errors.New("failed to parse ICAO to aircraft map")

type AircraftTypeRepo struct {
	icaoToAircraft map[string]ref.IcaoAircraftSpec
}

func NewAircraftTypeRepo() (*AircraftTypeRepo, error) {
	const initError = "NewAircraftDescriptionRepo: %w caused by %w"
	icaoToAircraftMap, aircraftErr := getIcaoToAircraftMap()
	if aircraftErr != nil {
		return nil, fmt.Errorf(initError, errParseIcaoAircraftMap, aircraftErr)
	}

	repo := AircraftTypeRepo{
		icaoToAircraft: icaoToAircraftMap,
	}

	return &repo, nil
}

// GetAircraftType implements the AircraftTypeRepo interface of
// the same name.
// TODO: Turn IcaoCode into dedicated type.
func (acr *AircraftTypeRepo) GetAircraftType(icaoCode string) (ref.IcaoAircraftSpec, bool) {
	spec, exists := acr.icaoToAircraft[icaoCode]
	return spec, exists
}

// getIcaoToAircraftMap returns an ICAO id to aircraft record mapping.
func getIcaoToAircraftMap() (map[string]ref.IcaoAircraftSpec, error) {
	// Parse the CSV file
	icaoAircraftMap, err := parseIcaoCsvToMap(icaoListPath)
	if err != nil {
		return nil, fmt.Errorf("getIcaoToAircraftMap: %w: %w", errParseCSV, err)
	}

	return icaoAircraftMap, nil
}

// parseIcaoCsvToMap reads a CSV file and parses it into a map ICAO -> aircraft spec.
func parseIcaoCsvToMap(filePath string) (map[string]ref.IcaoAircraftSpec, error) {
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

	records := make(map[string]ref.IcaoAircraftSpec)

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
		records[key] = ref.IcaoAircraftSpec{
			Class:  class,
			Engine: engine,
			Make:   manufacturer,
		}
	}

	return records, nil
}
