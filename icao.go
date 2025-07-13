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

// parseCSVToMaps reads a CSV file and parses it into a slice of maps.
// Each map represents a row, with keys corresponding to the CSV headers.
func parseCSVToMap(filePath string) (map[string]IcaoAircraft, error) {
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

func GetIcaoAircraftMap() map[string]IcaoAircraft {

	const icaoListPath string = "./data/ICAOList.csv"
	fmt.Printf("Parsing CSV file: %s\n\n", icaoListPath)

	// Parse the CSV file
	icaoAircraftMap, err := parseCSVToMap(icaoListPath)
	if err != nil {
		fmt.Printf("Error parsing CSV: %v\n", err)
		return nil
	}

	return icaoAircraftMap
}
