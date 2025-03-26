package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
)

// TestGetSwiftCode tests the getSwiftCode endpoint
func TestGetSwiftCode(t *testing.T) {
	// Initialize the database and insert test data
	initDB()
	createTable()
	parseAndInsertData()

	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/v1/swift-codes/BERLMCMCXXX", nil)
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
	expected := `{"address":"LES TERRASSES, CARLO 2 AVENUE DE MONTE MONACO, MONACO, 98000","bankName":"EDMOND DE ROTHSCHILD-MONACO",
	"countryISO2":"MC","countryName":"MONACO","isHeadquarter":true,"swiftCode":"BERLMCMCXXX",
	"branches":[{"address":"Branch Address","bankName":"Branch Bank","countryISO2":"AD","isHeadquarter":false,"swiftCode":"ADCRBGS1YYY"}]}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
