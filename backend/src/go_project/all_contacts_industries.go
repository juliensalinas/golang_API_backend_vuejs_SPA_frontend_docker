/*
Send to frontend a list of industries existing the prospectsocialprofile.industry.
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

type ContSocProfRow struct {
	Industry string `json:"industryName"`
}

// getAllContactsIndustries queries db to retrieve a distinct list of all industries
// in prospectsocialprofile table
func getAllContactsIndustries(w http.ResponseWriter) ([]ContSocProfRow, error) {

	var contSocProfRows []ContSocProfRow
	var err error

	dbinfo := fmt.Sprintf(`host=%s port=%d user=%s password=%s dbname=%s
        sslmode=disable`, localHost, localPort, localUser, localPassword, localDbname)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		err = CustErr(err, "DB connection failed\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return contSocProfRows, err
	}
	defer db.Close()

	sqlStmt := "SELECT DISTINCT(industry) FROM prospectsocialprofile WHERE industry <> ''"

	rows, err := db.Query(sqlStmt)
	if err != nil {
		err = CustErr(err, "SQL query failed.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return contSocProfRows, err
	}
	defer rows.Close()

	for rows.Next() {
		var contSocProfRow ContSocProfRow
		if err := rows.Scan(&contSocProfRow.Industry); err != nil {
			err = CustErr(err, "A row could not be read from SQL query results.\nNOT stopping here.")
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return contSocProfRows, err
		}
		contSocProfRows = append(contSocProfRows, contSocProfRow)
	}

	return contSocProfRows, err
}

// ReturnContactsIndustriesList loads all contacts industries from db and send it in JSON to frontend
func ReturnContactsIndustriesList(w http.ResponseWriter, r *http.Request) {

	contSocProfRows, err := getAllContactsIndustries(w)
	if err != nil {
		return
	}

	if len(contSocProfRows) == 0 {
		log.Println("No result found\nStopping here.")
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	returnedJson, err := json.Marshal(contSocProfRows)
	if err != nil {
		err = CustErr(err, "Could not not marshall to JSON.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", returnedJson)

}
