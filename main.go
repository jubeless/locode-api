package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
	"golang.org/x/exp/rand"
)

type Location struct {
	Locode       string `json:"locode"`
	CountryCode  string `json:"country_code"`
	CountryName  string `json:"country_name"`
	AdminCode    string `json:"admin_code"`
	LocationCode string `json:"location_code"`
	Name         string `json:"name"`
	AltName      string `json:"alt_name"`
	Coordinates  string `json:"coordinates"`
}

func extractLocations(filePath string) ([]Location, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','          // Set the delimiter to comma
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	var locations []Location
	var currentCountryCode string
	var countryName string

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV record: %w", err)
		}

		// Check if the row defines a new country
		if len(record) > 1 && record[0] == "" && record[1] != "" && record[2] == "" && record[3] != "" {
			currentCountryCode = record[1] // Set the current country code.
			countryName = strings.ToLower(strings.ReplaceAll(record[3], ".", ""))
			continue // Skip to the next record.
		}

		if len(record) < 11 {
			fmt.Printf("Skipping incomplete record: %v\n", record)
			continue
		}

		if currentCountryCode == "" && record[1] != "" {
			currentCountryCode = record[1]
		}

		location := Location{
			Locode:       currentCountryCode + record[2],
			CountryCode:  currentCountryCode,
			CountryName:  countryName,
			AdminCode:    record[1],
			LocationCode: record[2],
			Name:         record[3],
			AltName:      record[4],
			Coordinates:  record[10],
		}

		locations = append(locations, location)
	}
	return locations, nil
}

func loadLocodes() (cmap.ConcurrentMap[string, Location], error) {
	files := []string{
		"locode_1.csv",
		"locode_2.csv",
		"locode_3.csv",
	}

	locodes := cmap.New[Location]()

	for _, filepath := range files {
		locations, err := extractLocations(filepath)
		if err != nil {
			return locodes, fmt.Errorf("failed to extract locations: %w", err)
		}

		for _, location := range locations {
			locodes.Set(location.Locode, location)
		}
	}
	fmt.Printf("Loaded %d locations\n", locodes.Count())
	return locodes, nil
}

func main() {
	locodes, err := loadLocodes()
	if err != nil {
		log.Fatalf("Failed to load locodes: %v", err)
		return
	}

	h := &handler{
		locodes: locodes,
	}

	http.HandleFunc("/locode", h.locodeHandler)
	http.HandleFunc("/random", h.random)
	fmt.Println("Starting server on port 5555")
	if err := http.ListenAndServe(":5555", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

type handler struct {
	locodes cmap.ConcurrentMap[string, Location]
}

func (h *handler) locodeHandler(w http.ResponseWriter, r *http.Request) {
	locode := r.URL.Query().Get("locode")
	if locode == "" {
		http.Error(w, "Missing 'locode' parameter", http.StatusBadRequest)
		return
	}

	location, ok := h.locodes.Get(locode)
	if !ok {
		http.Error(w, fmt.Sprintf("Unknown locode %s", locode), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(location)
}

func (h *handler) random(w http.ResponseWriter, r *http.Request) {
	numToSelect := 3000
	v, err := strconv.ParseUint(r.URL.Query().Get("count"), 10, 64)
	if err == nil && v > 0 {
		numToSelect = int(v)
	}

	rand.Seed(uint64(time.Now().UnixNano()))
	allLocodes := h.locodes.Items()
	keys := make([]string, 0, len(allLocodes))
	for k := range allLocodes {
		keys = append(keys, k)
	}

	rand.Shuffle(len(keys), func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	if len(keys) < numToSelect {
		numToSelect = len(keys)
	}
	selectedKeys := keys[:numToSelect]
	w.Header().Set("Content-Type", "text/plain")

	for _, key := range selectedKeys {
		w.Write([]byte(fmt.Sprintf("%s\n", allLocodes[key].Locode)))
	}
}
