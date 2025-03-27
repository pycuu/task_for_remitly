package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

// response structure for a single SWIFT code
type SwiftCodeResponse struct {
	Address       string              `json:"address"`
	BankName      string              `json:"bankName"`
	CountryISO2   string              `json:"countryISO2"`
	CountryName   string              `json:"countryName"`
	IsHeadquarter bool                `json:"isHeadquarter"`
	SwiftCode     string              `json:"swiftCode"`
	Branches      []SwiftCodeResponse `json:"branches,omitempty"`
}

// response structure for a country' s SWIFT codes
type CountrySwiftCodesResponse struct {
	CountryISO2 string              `json:"countryISO2"`
	CountryName string              `json:"countryName"`
	SwiftCodes  []SwiftCodeResponse `json:"swiftCodes"`
}

// endpoint 1 - Retrieve details of a single SWIFT code whether for a headquarters or branches
func getSwiftCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	swiftCode := strings.ToUpper(vars["swift-code"])

	sc, err := GetSwiftCodeDetails(swiftCode)
	if err != nil {
		http.Error(w, "SWIFT code not found", http.StatusNotFound)
		fmt.Println(err)
		return
	}

	response := SwiftCodeResponse{
		Address:       sc.Address,
		BankName:      sc.BankName,
		CountryISO2:   sc.CountryCode,
		CountryName:   sc.CountryName,
		IsHeadquarter: sc.IsHeadquarter,
		SwiftCode:     sc.SwiftCode,
	}

	if sc.IsHeadquarter {
		branches, err := GetSwiftCodesByCountry(sc.CountryCode)
		if err == nil {
			for _, branch := range branches {
				if branch.HeadquarterSWIFTCode.String == sc.SwiftCode {
					response.Branches = append(response.Branches, SwiftCodeResponse{
						Address:       branch.Address,
						BankName:      branch.BankName,
						CountryISO2:   branch.CountryCode,
						IsHeadquarter: branch.IsHeadquarter,
						SwiftCode:     branch.SwiftCode,
					})
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// endpoint 2 - Return all SWIFT codes with details for a specific country (both headquarters and branches)
func getSwiftCodesByCountry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	countryISO2 := strings.ToUpper(vars["countryISO2code"])

	swiftCodes, err := GetSwiftCodesByCountry(countryISO2)
	if err != nil || len(swiftCodes) == 0 {
		http.Error(w, "No SWIFT codes found for this country", http.StatusNotFound)
		return
	}

	response := struct {
		CountryISO2 string `json:"countryISO2"`
		CountryName string `json:"countryName"`
		SwiftCodes  []struct {
			SwiftCode     string `json:"swiftCode"`
			CountryISO2   string `json:"countryISO2"`
			IsHeadquarter bool   `json:"isHeadquarter"`
			Address       string `json:"address"`
			BankName      string `json:"bankName"`
		} `json:"swiftCodes"`
	}{
		CountryISO2: swiftCodes[0].CountryCode,
		CountryName: swiftCodes[0].CountryName,
	}

	for _, sc := range swiftCodes {
		response.SwiftCodes = append(response.SwiftCodes, struct {
			SwiftCode     string `json:"swiftCode"`
			CountryISO2   string `json:"countryISO2"`
			IsHeadquarter bool   `json:"isHeadquarter"`
			Address       string `json:"address"`
			BankName      string `json:"bankName"`
		}{
			SwiftCode:     sc.SwiftCode,
			CountryISO2:   sc.CountryCode,
			IsHeadquarter: sc.IsHeadquarter,
			Address:       sc.Address,
			BankName:      sc.BankName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// endpoint 3 - Adds new SWIFT code entries to the database for a specific country
func addSwiftCode(w http.ResponseWriter, r *http.Request) {
	var sc SwiftCode
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&sc); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	sc.CountryCode = strings.ToUpper(sc.CountryCode)
	sc.CountryName = strings.ToUpper(sc.CountryName)

	err := InsertSwiftCode(sc)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting SWIFT code: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "SWIFT code added successfully"})
}

// endpoint 4 - Deletes swift-code data if swiftCode matches the one in the database
func deleteSwiftCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	swiftCode := vars["swift-code"]

	message, err := DeleteSwiftCodeData(swiftCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func main() {
	// Get database connection details from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	// Ensure all required environment variables are set
	if dbHost == "" || dbUser == "" || dbPassword == "" || dbName == "" || dbPort == "" {
		log.Fatal("One or more required environment variables are missing: DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, DB_PORT")
	}

	// Initialize the database connection
	initDB(dbHost, dbUser, dbPassword, dbName, dbPort)

	// Create the table and parse data
	createTable()
	parseAndInsertData()

	// Set up the router and endpoints
	r := mux.NewRouter()
	r.HandleFunc("/v1/swift-codes/{swift-code}", getSwiftCode).Methods("GET")
	r.HandleFunc("/v1/swift-codes/country/{countryISO2code}", getSwiftCodesByCountry).Methods("GET")
	r.HandleFunc("/v1/swift-codes", addSwiftCode).Methods("POST")
	r.HandleFunc("/v1/swift-codes/{swift-code}", deleteSwiftCode).Methods("DELETE")

	// Start the server
	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
