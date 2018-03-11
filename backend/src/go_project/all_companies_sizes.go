/*
Send to frontend a list of companies sizes existing in company.size
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

type CompanyRow struct {
	Size string `json:"sizeName"`
}

// getAllCompaniesSizes queries db to retrieve a distinct list of all sizes
// in company table
func getAllCompaniesSizes(w http.ResponseWriter) ([]CompanyRow, error) {

	var companyRows []CompanyRow
	var err error

	dbinfo := fmt.Sprintf(`host=%s port=%d user=%s password=%s dbname=%s
        sslmode=disable`, localHost, localPort, localUser, localPassword, localDbname)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		err = CustErr(err, "DB connection failed\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return companyRows, err
	}
	defer db.Close()

	sqlStmt := "SELECT DISTINCT(size) FROM company WHERE size <> ''"

	rows, err := db.Query(sqlStmt)
	if err != nil {
		err = CustErr(err, "SQL query failed.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return companyRows, err
	}
	defer rows.Close()

	for rows.Next() {
		var companyRow CompanyRow
		if err := rows.Scan(&companyRow.Size); err != nil {
			err = CustErr(err, "A row could not be read from SQL query results.\nNOT stopping here.")
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return companyRows, err
		}
		companyRows = append(companyRows, companyRow)
	}

	return companyRows, err
}

// ReturnCompaniesSizesList loads all companies sizes from db and send it in JSON to frontend
func ReturnCompaniesSizesList(w http.ResponseWriter, r *http.Request) {

	companyRows, err := getAllCompaniesSizes(w)
	if err != nil {
		return
	}

	if len(companyRows) == 0 {
		log.Println("No result found\nStopping here.")
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	returnedJson, err := json.Marshal(companyRows)
	if err != nil {
		err = CustErr(err, "Could not not marshall to JSON.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", returnedJson)

}
