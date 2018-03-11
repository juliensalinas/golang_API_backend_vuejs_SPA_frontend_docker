/*
Send to frontend a list of industries existing the companysocialprofile.industry.
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

type CompSocProfRow struct {
	Industry string `json:"industryName"`
}

// getAllCompaniesIndustries queries db to retrieve a distinct list of all industries
// in companysocialprofile table
func getAllCompaniesIndustries(w http.ResponseWriter) ([]CompSocProfRow, error) {

	var compSocProfRows []CompSocProfRow
	var err error

	dbinfo := fmt.Sprintf(`host=%s port=%d user=%s password=%s dbname=%s
        sslmode=disable`, localHost, localPort, localUser, localPassword, localDbname)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		err = CustErr(err, "DB connection failed\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return compSocProfRows, err
	}
	defer db.Close()

	sqlStmt := "SELECT DISTINCT(industry) FROM companysocialprofile WHERE industry <> ''"

	rows, err := db.Query(sqlStmt)
	if err != nil {
		err = CustErr(err, "SQL query failed.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return compSocProfRows, err
	}
	defer rows.Close()

	for rows.Next() {
		var compSocProfRow CompSocProfRow
		if err := rows.Scan(&compSocProfRow.Industry); err != nil {
			err = CustErr(err, "A row could not be read from SQL query results.\nNOT stopping here.")
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return compSocProfRows, err
		}
		compSocProfRows = append(compSocProfRows, compSocProfRow)
	}

	return compSocProfRows, err
}

// ReturnCompaniesIndustriesList loads all companies industries from db and send it in JSON to frontend
func ReturnCompaniesIndustriesList(w http.ResponseWriter, r *http.Request) {

	compSocProfRows, err := getAllCompaniesIndustries(w)
	if err != nil {
		return
	}

	if len(compSocProfRows) == 0 {
		log.Println("No result found\nStopping here.")
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	returnedJson, err := json.Marshal(compSocProfRows)
	if err != nil {
		err = CustErr(err, "Could not not marshall to JSON.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", returnedJson)

}
