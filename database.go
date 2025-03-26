package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type SwiftCode struct {
	SwiftCode            string
	CountryCode          string
	CountryName          string
	IsHeadquarter        bool
	HeadquarterSWIFTCode sql.NullString
	Address              string
	BankName             string
}

var db *sql.DB

// init the database
func initDB(host, user, password, dbname, port string) {
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening database connection: ", err)
	}

	// Check the connection
	if err := db.Ping(); err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
}

// create the table
func createTable() {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS swift_codes (
        swift_code VARCHAR(11) PRIMARY KEY,
        country_code CHAR(2) NOT NULL,
        country_name VARCHAR(100) NOT NULL,
        is_headquarter BOOLEAN NOT NULL,
        headquarter_swift_code VARCHAR(11),
        address VARCHAR(255),
		bank_name VARCHAR(255)
    );`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Error creating table: ", err)
	}

}

// parse the CSV and insert data into the database
func parseAndInsertData() {
	file, err := os.Open("data.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	headquarters := make(map[string]string)

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		countryCode := strings.ToUpper(record[0])
		swiftCode := strings.ToUpper(record[1])
		countryName := strings.ToUpper(record[6])
		address := strings.ToUpper(record[4])
		bankName := strings.ToUpper(record[3])

		isHeadquarter := strings.HasSuffix(swiftCode, "XXX")
		var headquarterSWIFTCode string

		if !isHeadquarter {
			headquarterSWIFTCode = headquarters[swiftCode[:8]]
		}

		insertSQL := `
        INSERT INTO swift_codes (swift_code, country_code, country_name, is_headquarter, headquarter_swift_code, address, bank_name)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (swift_code) DO NOTHING;`
		// fmt.Println(bankName)

		if isHeadquarter {
			headquarters[swiftCode[:8]] = swiftCode
		}

		_, err = db.Exec(insertSQL, swiftCode, countryCode, countryName, isHeadquarter, headquarterSWIFTCode, address, bankName)
		if err != nil {
			log.Printf("Error inserting data for %s: %v\n", swiftCode, err)
		}

	}
}

func GetSwiftCodeDetails(swiftCode string) (SwiftCode, error) {
	var sc SwiftCode
	var address sql.NullString
	var bankName sql.NullString
	query := `
    SELECT swift_code, country_code, country_name, is_headquarter, headquarter_swift_code, address, bank_name
    FROM swift_codes
    WHERE swift_code = $1;`

	err := db.QueryRow(query, swiftCode).Scan(&sc.SwiftCode, &sc.CountryCode, &sc.CountryName, &sc.IsHeadquarter, &sc.HeadquarterSWIFTCode, &address, &bankName)
	if err != nil {
		return SwiftCode{}, err
	}

	// handle nulls
	if address.Valid {
		sc.Address = address.String
	} else {
		sc.Address = "N/A"
	}

	return sc, nil
}

// deletes a SWIFT code from the database
func DeleteSwiftCodeData(swiftCode string) (string, error) {
	query := `
    DELETE FROM swift_codes
    WHERE swift_code = $1;`

	result, err := db.Exec(query, swiftCode)
	if err != nil {
		return "", fmt.Errorf("error deleting SWIFT code %s: %v", swiftCode, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("error fetching rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return "", fmt.Errorf("no SWIFT code found with code %s", swiftCode)
	}

	return fmt.Sprintf("SWIFT code %s successfully deleted", swiftCode), nil
}

// insert new SWIFT code
func InsertSwiftCode(sc SwiftCode) error {
	insertSQL := `
    INSERT INTO swift_codes (swift_code, country_code, country_name, is_headquarter, headquarter_swift_code, address, bank_name)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    ON CONFLICT (swift_code) DO NOTHING;`

	var headquarterSWIFTCode *string
	if sc.IsHeadquarter {
		headquarterSWIFTCode = nil
	} else {
		headquarterSWIFTCode = &sc.HeadquarterSWIFTCode.String
	}

	_, err := db.Exec(insertSQL, sc.SwiftCode, sc.CountryCode, sc.CountryName, sc.IsHeadquarter, headquarterSWIFTCode, sc.Address, sc.BankName)
	return err
}

// get all SWIFT codes for a specific country
func GetSwiftCodesByCountry(countryCode string) ([]SwiftCode, error) {
	var swiftCodes []SwiftCode
	query := `
    SELECT swift_code, country_code, country_name, is_headquarter, headquarter_swift_code, address, bank_name
    FROM swift_codes
    WHERE country_code = $1;`

	//fmt.Println("Executing query:", query, "with countryCode:", countryCode)

	rows, err := db.Query(query, countryCode)
	if err != nil {
		fmt.Println("Query error:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sc SwiftCode
		var headquarterSWIFTCode sql.NullString
		var address sql.NullString
		var bankName sql.NullString

		err := rows.Scan(&sc.SwiftCode, &sc.CountryCode, &sc.CountryName, &sc.IsHeadquarter, &headquarterSWIFTCode, &address, &bankName)
		if err != nil {
			fmt.Println("Row scan error:", err)
			return nil, err
		}

		// handling nulls
		sc.HeadquarterSWIFTCode = headquarterSWIFTCode
		sc.Address = address.String
		if !address.Valid {
			sc.Address = "N/A"
		}

		sc.BankName = bankName.String
		if !bankName.Valid {
			sc.BankName = "Unknown Bank"
		}

		swiftCodes = append(swiftCodes, sc)
	}

	//fmt.Println("Retrieved SWIFT codes:", swiftCodes)
	return swiftCodes, nil
}
