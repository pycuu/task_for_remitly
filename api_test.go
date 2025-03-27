package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGetSwiftCode(t *testing.T) {
	// Initialize the database
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	initDB(dbHost, dbUser, dbPassword, dbName, dbPort)
	createTable()

	// Insert test data
	swiftCode := SwiftCode{
		SwiftCode:     "TESTSWIFTXXX",
		CountryCode:   "US",
		CountryName:   "UNITED STATES",
		IsHeadquarter: true,
		Address:       "123 Test Address",
		BankName:      "Test Bank",
	}
	InsertSwiftCode(swiftCode)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/v1/swift-codes/TESTSWIFTXXX", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new HTTP recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getSwiftCode)

	// Serve the HTTP request
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := `{"address":"123 Test Address","bankName":"Test Bank","countryISO2":"US","countryName":"UNITED STATES","isHeadquarter":true,"swiftCode":"TESTSWIFTXXX"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetSwiftCodesByCountry(t *testing.T) {
	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/v1/swift-codes/country/US", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new HTTP recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getSwiftCodesByCountry)

	// Serve the HTTP request
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := `{"countryISO2":"US","countryName":"UNITED STATES","swiftCodes":[{"swiftCode":"TESTSWIFTXXX","countryISO2":"US","isHeadquarter":true,"address":"123 Test Address","bankName":"Test Bank"}]}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
