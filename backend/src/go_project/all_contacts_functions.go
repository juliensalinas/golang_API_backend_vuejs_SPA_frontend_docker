/*
Send to frontend a list of functions existing in the job_function table
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

type JobFunctionRow struct {
	Name string `json:"functionName"`
}

// getAllContactsFunctions queries db to retrieve a distinct list of all functions
// in job_function table
func getAllContactsFunctions(w http.ResponseWriter) ([]JobFunctionRow, error) {

	var jobFunctionRows []JobFunctionRow
	var err error

	dbinfo := fmt.Sprintf(`host=%s port=%d user=%s password=%s dbname=%s
        sslmode=disable`, localHost, localPort, localUser, localPassword, localDbname)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		err = CustErr(err, "DB connection failed\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return jobFunctionRows, err
	}
	defer db.Close()

	sqlStmt := "SELECT name FROM job_function"

	rows, err := db.Query(sqlStmt)
	if err != nil {
		err = CustErr(err, "SQL query failed.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return jobFunctionRows, err
	}
	defer rows.Close()

	for rows.Next() {
		var jobFunctionRow JobFunctionRow
		if err := rows.Scan(&jobFunctionRow.Name); err != nil {
			err = CustErr(err, "A row could not be read from SQL query results.\nNOT stopping here.")
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return jobFunctionRows, err
		}
		jobFunctionRows = append(jobFunctionRows, jobFunctionRow)
	}

	return jobFunctionRows, err
}

// ReturnContactsFunctionsList loads all contacts functions from db and send it in JSON to frontend
func ReturnContactsFunctionsList(w http.ResponseWriter, r *http.Request) {

	jobFunctionRows, err := getAllContactsFunctions(w)
	if err != nil {
		return
	}

	if len(jobFunctionRows) == 0 {
		log.Println("No result found\nStopping here.")
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	returnedJson, err := json.Marshal(jobFunctionRows)
	if err != nil {
		err = CustErr(err, "Could not not marshall to JSON.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", returnedJson)

}
