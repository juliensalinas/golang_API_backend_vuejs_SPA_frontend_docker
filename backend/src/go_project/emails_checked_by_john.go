/*
Return emails from the email_checked_by_john table.
*/

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Tried to be clean here and map the structs data types exactly to the
// DB data types (even if putting "string" everywhere seems to work).
// Biggest issue is about null values from DB. I found 3 options:
// 1) not change anything: errors will be raised when a null value tries
// to be allocated to the wrong type but it does not stop the script.
// Problem: some data seem to be missing in the final Json eventually.
// 2) use the sql lib nullable types: replace string with sql.NullString,
// int with sql.NullInt64, bool with sql.NullBool, and time with sql.NullTime
// but then we obtain something like {"Valid":true,"String":"Smith"} which is
// not directly ok in JSON. So it requires extra steps before Marshalling to
// Json.
// 3) use pointers for nullable values (*string,...): it works but null values
// are not detected by the 'omitempty' keyword during marshalling so an empty
// string will be displayed in JSON. This is my option for the moment.
type EmailCheckedByPA struct {
	Id                               int        `json:"id,omitempty"`
	MissionNumber                    int        `json:"missionnumber,omitempty"`
	FirstName                        *string    `json:"firstname,omitempty"`
	LastName                         *string    `json:"lastname,omitempty"`
	EmailDomain                      *string    `json:"emaildomain,omitempty"`
	Email                            string     `json:"email,omitempty"`
	ContactFromC2LId                 *int       `json:"contactfromc2lid,omitempty"`
	QEVResult                        *string    `json:"qevresult,omitempty"`
	QEVReason                        *string    `json:"qevreason,omitempty"`
	QEVDisposable                    *bool      `json:"qevdisposable,omitempty"`
	QEVAcceptAll                     *bool      `json:"qevacceptall,omitempty"`
	QEVRole                          *bool      `json:"qevrole,omitempty"`
	QEVFree                          *bool      `json:"qevfree,omitempty"`
	QEVSafeToSend                    *bool      `json:"qevsafetosend,omitempty"`
	QEVDidYouMean                    *string    `json:"qevdidyoumean,omitempty"`
	QEVSuccess                       *bool      `json:"qevsuccess,omitempty"`
	QEVMessage                       *string    `json:"qevmessage,omitempty"`
	APICheckDateTime                 *time.Time `json:"apicheckdatetime,omitempty"`
	ManualEmailSendingDatetime       *time.Time `json:"manualemailsendingdatetime,omitempty"`
	ManualEmailErrorResponseDatetime *time.Time `json:"manualemailerrorresponsedatetime,omitempty"`
	ContactId                        int        `json:"contactid,omitempty"`
}

// getResFromDB queries DB and stores results in []EmailCheckedByJohn
func getResFromDB(missionNumber string, w http.ResponseWriter) ([]EmailCheckedByjJhn, error) {

	var emails []EmailCheckedByJohn
	var err error

	// Connect to db:
	dbinfo := fmt.Sprintf(`host=%s port=%d user=%s password=%s dbname=%s
        sslmode=disable`, localHost, localPort, localUser, localPassword, localDbname)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		err = CustErr(err, "DB connection failed\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return emails, err
	}
	defer db.Close()

	// Make the sql query:
	sqlStatement := `SELECT * FROM email_checked_by_john 
    	WHERE mission_number = $1`
	rows, err := db.Query(sqlStatement, missionNumber)
	if err != nil {
		err = CustErr(err, "Following query failed: "+sqlStatement+"\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return emails, err
	}
	defer rows.Close()

	// Get results from db and put it in struct:
	for rows.Next() {
		var email EmailCheckedByJohn
		err = rows.Scan(
			&email.Id,
			&email.MissionNumber,
			&email.FirstName,
			&email.LastName,
			&email.EmailDomain,
			&email.Email,
			&email.ContactFromC2LId,
			&email.QEVResult,
			&email.QEVReason,
			&email.QEVDisposable,
			&email.QEVAcceptAll,
			&email.QEVRole,
			&email.QEVFree,
			&email.QEVSafeToSend,
			&email.QEVDidYouMean,
			&email.QEVSuccess,
			&email.QEVMessage,
			&email.APICheckDateTime,
			&email.ManualEmailSendingDatetime,
			&email.ManualEmailErrorResponseDatetime,
			&email.ContactId,
		)
		if err != nil {
			err = CustErr(err, "One row could not be retrieved from DB.\nNOT Stopping here.")
			log.Println(err)
			return emails, err
		}
		emails = append(emails, email)
	}

	return emails, err

}

// ReturnEmailsCheckedByJohn returns results through a REST API
func ReturnEmailsCheckedByJohn(w http.ResponseWriter, r *http.Request) {

	var err error

	// Get the parameter passed to url and query the database based on this
	// parameter:
	params := mux.Vars(r)

	// If not an integer, we return an error:
	missionNumber := params["missionnumber"]
	_, err = strconv.Atoi(missionNumber)
	if err != nil {
		err = CustErr(err, "Mission number is not an integer.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get results from DB
	emails, err := getResFromDB(missionNumber, w)
	if err != nil {
		return
	}

	// If no result found in DB, return a 404 page:
	if len(emails) == 0 {
		log.Println("\nNo result found for this mission number: " + missionNumber + "\nStopping here.")
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Turn struct into a proper JSON response:
	returnedJson, err := json.Marshal(emails)
	if err != nil {
		err = CustErr(err, "Could not marshall to JSON.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set a custom JSON header and send response to client:
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", returnedJson)

}
