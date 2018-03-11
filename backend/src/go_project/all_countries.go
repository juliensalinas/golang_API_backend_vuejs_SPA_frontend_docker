/*
Send to frontend a list of countries existing in the postal_address table.
This is not all countries but only countries used in Local.
*/

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

type CountryRow struct {
	CountryName string `json:"countryName"`
}

// getAllCountries queries db to retrieve a distinct list of all countries
// in postal_address table
func getAllCountries(w http.ResponseWriter) ([]CountryRow, error) {

	var countriesRows []CountryRow
	var err error

	dbinfo := fmt.Sprintf(`host=%s port=%d user=%s password=%s dbname=%s
        sslmode=disable`, localHost, localPort, localUser, localPassword, localDbname)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		err = CustErr(err, "DB connection failed\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return countriesRows, err
	}
	defer db.Close()

	sqlStmt := "SELECT DISTINCT(country) FROM postal_address WHERE country <> ''"

	rows, err := db.Query(sqlStmt)
	if err != nil {
		err = CustErr(err, "SQL query failed.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return countriesRows, err
	}
	defer rows.Close()

	for rows.Next() {
		var countryRow CountryRow
		if err := rows.Scan(&countryRow.CountryName); err != nil {
			err = CustErr(err, "A row could not be read from SQL query results.\nNOT stopping here.")
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return countriesRows, err
		}
		countriesRows = append(countriesRows, countryRow)
	}

	return countriesRows, err
}

// ReturnCountryList loads all countries from db and send it in JSON to frontend
func ReturnCountriesList(w http.ResponseWriter, r *http.Request) {

	countriesRows, err := getAllCountries(w)
	if err != nil {
		return
	}

	if len(countriesRows) == 0 {
		log.Println("No result found\nStopping here.")
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	returnedJson, err := json.Marshal(countriesRows)
	if err != nil {
		err = CustErr(err, "Could not not marshall to JSON.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", returnedJson)

}
