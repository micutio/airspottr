package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

var IcaoAircraftHeaders = [...]string{
	"Aircraft TypeDesignator",
	"Class",
	"Number+Engine Type",
	"\"MANUFACTURER, Model\"",
}

type IcaoAircraft struct {
	Class     string
	Engine    string
	ModelCode string
}

func GetIcaoAircraftMap() map[string]IcaoAircraft {

	const icaoListPath string = "./data/ICAOList.csv"
	fmt.Printf("Parsing CSV file: %s\n\n", icaoListPath)

	// Parse the CSV file
	icaoAircraftMap, err := parseIcaoCsvToMap(icaoListPath)
	if err != nil {
		fmt.Printf("Error parsing CSV: %v\n", err)
		return nil
	}

	return icaoAircraftMap
}

// parseIcaoCsvToMap reads a CSV file and parses it into a map ICAO -> aircraft spec.
func parseIcaoCsvToMap(filePath string) (map[string]IcaoAircraft, error) {
	// Open the CSV file
	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		return nil, fmt.Errorf("failed to open file: %w", fileErr)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("failed to close file: %v", err)
		}
	}(file) // Ensure the file is closed when the function exits

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header row
	headers, headerErr := reader.Read()
	if headerErr != nil {
		return nil, fmt.Errorf("failed to read headers: %w", headerErr)
	}
	if len(headers) != len(IcaoAircraftHeaders) {
		return nil, fmt.Errorf("unexpected header length")
	}

	records := make(map[string]IcaoAircraft)

	// Loop through the remaining records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break // End of file
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		key := record[0]
		class := record[1]
		engine := record[2]
		manufacturer := record[3]
		records[key] = IcaoAircraft{class, engine, manufacturer}
	}

	return records, nil
}

func GetMilCodeMap() map[string]string {

	const milCodeFilePath string = "./data/MILICAOOperatorLookup.csv"
	fmt.Printf("Parsing CSV file: %s\n\n", milCodeFilePath)

	// Parse the CSV file
	icaoAircraftMap, err := parseMilCodeToMap(milCodeFilePath)
	if err != nil {
		fmt.Printf("Error parsing CSV: %v\n", err)
		return nil
	}

	return icaoAircraftMap
}

// parseMilCodeToMap reads a CSV file and parses it into a map code -> military operator
func parseMilCodeToMap(filePath string) (map[string]string, error) {
	// Open the CSV file
	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		return nil, fmt.Errorf("failed to open file: %w", fileErr)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("failed to close file: %v", err)
		}
	}(file) // Ensure the file is closed when the function exits

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header row
	headers, headerErr := reader.Read()
	if headerErr != nil {
		return nil, fmt.Errorf("failed to read headers: %w", headerErr)
	}
	if len(headers) != 2 {
		return nil, fmt.Errorf("unexpected header length")
	}

	records := make(map[string]string)

	// Loop through the remaining records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break // End of file
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		if len(record) < 2 {
			continue
		}

		key := record[1]
		militaryOperator := record[0]
		records[key] = militaryOperator
	}

	return records, nil
}
