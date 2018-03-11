/*
companies_and_contacts.go get an http request routed by main.go and
queries db for companies and associated contacts.
SQL query is a complex and modular query based on various parameters
passed in an input JSON from frontend.
Results are returned as JSON through a REST API.
*/

package main

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"gopkg.in/gomail.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	returnedArchiveName = "results.zip"
	returnedCSVName     = "companies_and_contacts_extracted.csv"
)

// UserInput stores user input sent through JSON.
// CompanyHasPhone, CompanyHasEmail, and ContactHasEmail are fake
// booleans: 0: not set, 1: false, 2: true
type UserInput struct {
	Step                          string   `json:"step"`
	CompanyCity                   string   `json:"companyCity"`
	CompanyPostCode               string   `json:"companyPostCode"`
	CompanyCountries              []string `json:"companyCountries"`
	CompanyIndustries             []string `json:"companyIndustries"`
	CompanySizes                  []string `json:"companySizes"`
	CompanyTypes                  []string `json:"companyTypes"`
	CompanyHasPhone               int      `json:"companyHasPhone"`
	CompanyHasEmail               int      `json:"companyHasEmail"`
	CompanyDomains                []string `json:"companyDomains"`
	ExcludedCompanyDomains        []string `json:"excludedCompanyDomains"`
	ContactCity                   string   `json:"contactCity"`
	ContactPostCode               string   `json:"contactPostCode"`
	ContactCountries              []string `json:"contactCountries"`
	ContactIndustries             []string `json:"contactIndustries"`
	ContactJobTitle               string   `json:"contactJobTitle"`
	ContactFunctions              []string `json:"contactFunctions"`
	ContactJobLevels              []string `json:"contactJobLevels"`
	ContactHasEmail               int      `json:"contactHasEmail"`
	ContactRemoteAccounts         []string `json:"contactRemoteAccounts"`
	ExcludedContactRemoteAccounts []string `json:"excludedContactRemoteAccounts"`
}

// Created a custom type + method that implements the json.Marshaler
// so a simple string is returned in JSON for nullable values
// coming from database.
// https://stackoverflow.com/questions/33072172/how-can-i-work-with-sql-null-values-and-json-in-golang-in-a-good-way

// JsonNullString is a custom type replacing sql.NullString so when returning a JSON
// we only have a simple string or null returned.
// We are using an embedded type here (anonymous field): https://www.golang-book.com/books/intro/9
type JsonNullString struct {
	sql.NullString
}

// MarshalJSON is ou custom method for JsonNullString returning a string if
// sql.NullString is valid, or null otherwise.
func (v JsonNullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	} else {
		return json.Marshal(nil)
	}
}

// CompAndContRow stores results sent back to frontend in JSON.
// Empty values are present in JSON with a "null" value.
type CompAndContRow struct {
	CompId                       string         `json:"compId"`
	CompName                     JsonNullString `json:"compName"`
	CompDomain                   JsonNullString `json:"compDomain"`
	CompWebsite                  JsonNullString `json:"compWebsite"`
	CompTelephone                JsonNullString `json:"compTelephone"`
	CompFaxNumber                JsonNullString `json:"compFaxNumber"`
	CompSize                     JsonNullString `json:"compSize"`
	CompFounded                  JsonNullString `json:"compFounded"`
	CompCreatedOn                JsonNullString `json:"compCreatedOn"`
	CompUpdatedOn                JsonNullString `json:"compUpdatedOn"`
	CompStreetNumber             JsonNullString `json:"compStreetNumber"`
	CompRoute                    JsonNullString `json:"compRoute"`
	CompPostalCode               JsonNullString `json:"compPostalCode"`
	CompLocality                 JsonNullString `json:"compLocality"`
	CompAdministrativeAreaLevel2 JsonNullString `json:"compAdministrativeAreaLevel2"`
	CompAdministrativeAreaLevel1 JsonNullString `json:"compAdministrativeAreaLevel1"`
	CompCountry                  JsonNullString `json:"compCountry"`
	CompEmail                    JsonNullString `json:"compEmail"`
	CompSocProfURL               JsonNullString `json:"compSocProfURL"`
	CompType                     JsonNullString `json:"compType"`
	CompIndustry                 JsonNullString `json:"compIndustry"`
	ContId                       JsonNullString `json:"contId"`
	ContGender                   JsonNullString `json:"contGender"`
	ContFirstName                JsonNullString `json:"contFirstName"`
	ContLastName                 JsonNullString `json:"contLastName"`
	ContJobTitle                 JsonNullString `json:"contJobTitle"`
	ContTelephone                JsonNullString `json:"contTelephone"`
	ContCreatedOn                JsonNullString `json:"contCreatedOn"`
	ContUpdatedOn                JsonNullString `json:"contUpdatedOn"`
	ContStreetNumber             JsonNullString `json:"contStreetNumber"`
	ContRoute                    JsonNullString `json:"contRoute"`
	ContPostalCode               JsonNullString `json:"contPostalCode"`
	ContLocality                 JsonNullString `json:"contLocality"`
	ContAdministrativeAreaLevel2 JsonNullString `json:"contAdministrativeAreaLevel2"`
	ContAdministrativeAreaLevel1 JsonNullString `json:"contAdministrativeAreaLevel1"`
	ContCountry                  JsonNullString `json:"contCountry"`
	ContJobFunction              JsonNullString `json:"contJobFunction"`
	ContJobLevel                 JsonNullString `json:"contJobLevel"`
	ContEmail                    JsonNullString `json:"contEmail"`
	ContEmailStatus              JsonNullString `json:"contEmailStatus"`
	ContEmailCreatedOn           JsonNullString `json:"contEmailCreatedOn"`
	ContSocProfURL               JsonNullString `json:"contSocProfURL"`
	ContIndustry                 JsonNullString `json:"contIndustry"`
}

// CountRes stores only a number of rows return from SQL count
type CountRes struct {
	rowsNb int `json:"rowsNb"`
}

// convStringToWhereClause takes a user input and builds a piece of WHERE SQL query
// based on it. It consists of concatenating smartly one or several arguments in
// a WHERE clause with AND keywords (always using AND, not OR as asked by John for the
// moment but we can still change it later). We're using PostgreSQL prepared queries so
// positional arguments are necessary: "WHERE city = $1 AND post_code = $2 ...". It
// means me need to increment an index for every argument, and put the real argument
// value in a separate array that will be used later in db.Query(...).
// Arguments should not be case sensitive so we convert everything to UPPERCASE thanks
// to the UPPER() SQL function.
func convStringToWhereClause(
	userInputString string,
	attribute string,
	posIndex int,
	sqlArgs []interface{},
	sqlStmtPtr *strings.Builder,
) ([]interface{}, int) {

	// If user input was empty, do nothing
	if userInputString == "" {
		return sqlArgs, posIndex
	}

	// If posIndex = 1 it means it is the first argument of the big WHERE clause
	// so no need to add a "AND". Otherwise need to add a "AND" between arguments.
	// The piece of SQL created here could be something like:
	// AND UPPER(comp_ad.locality) = UPPER($5)
	posIndex += 1
	if posIndex > 1 {
		sqlStmtPtr.WriteString("AND ")
	}
	sqlStmtPtr.WriteString("UPPER(")
	sqlStmtPtr.WriteString(attribute)
	sqlStmtPtr.WriteString(") = UPPER($")
	sqlStmtPtr.WriteString(strconv.Itoa(posIndex))
	sqlStmtPtr.WriteString(") ")

	sqlArgs = append(sqlArgs, userInputString)

	return sqlArgs, posIndex

}

// convStringToWhereLikeClause does basically the same as convStringToWhereClause plus
// adds % before and after the user input because we want it to be an approximate search
func convStringToWhereLikeClause(
	userInputString string,
	attribute string,
	posIndex int,
	sqlArgs []interface{},
	sqlStmtPtr *strings.Builder,
) ([]interface{}, int) {

	if userInputString == "" {
		return sqlArgs, posIndex
	}

	var newUserInput strings.Builder
	newUserInput.WriteString("%")
	newUserInput.WriteString(userInputString)
	newUserInput.WriteString("%")
	newUserInputString := newUserInput.String()

	// The piece of SQL created here could be something like:
	// AND UPPER(cont.job_title) = LIKE UPPER($5)
	posIndex += 1
	if posIndex > 1 {
		sqlStmtPtr.WriteString("AND ")
	}
	sqlStmtPtr.WriteString("UPPER(")
	sqlStmtPtr.WriteString(attribute)
	sqlStmtPtr.WriteString(") LIKE UPPER($")
	sqlStmtPtr.WriteString(strconv.Itoa(posIndex))
	sqlStmtPtr.WriteString(") ")

	sqlArgs = append(sqlArgs, newUserInputString)

	return sqlArgs, posIndex

}

// convBoolToWhereClause does basically the same as convStringToWhereClause
// but converts a fake boolean (integer) to a corresponding SQL where clause.
// userInputInt = 0 --> not set so quit
// userInputInt = 1 --> false
// userInputInt = 2 --> true
// Also work for contHasEmail because prospect and prospectemail
// have a one-to-one relationship same behavior as for telephone.
// Also work for compHasEmail. company and companyemail have a
// one-to-many relationship but in DB we never have cases where a same company id has
// rows with an email and rows without an email. We either have one row only for a company
// id with a null email, or 1 or multiple rows for a company id with 1 or multiple emails.
func convBoolToWhereClause(
	userInputInt int,
	attribute string,
	posIndex int,
	sqlArgs []interface{},
	sqlStmtPtr *strings.Builder,
) ([]interface{}, int) {

	if userInputInt == 0 {
		return sqlArgs, posIndex
	}

	// The piece of SQL created here could be something like:
	// AND (comp.telephone IS NOT NULL AND comp.telephone <> '')
	// or:
	// AND (comp.telephone IS NULL OR comp.telephone = '')
	if posIndex > 0 {
		sqlStmtPtr.WriteString("AND ")
	}
	sqlStmtPtr.WriteString("(")
	sqlStmtPtr.WriteString(attribute)
	if userInputInt == 2 {
		sqlStmtPtr.WriteString(" IS NOT NULL AND ")
		sqlStmtPtr.WriteString(attribute)
		sqlStmtPtr.WriteString(" <> '') ")
	} else {
		sqlStmtPtr.WriteString(" IS NULL OR ")
		sqlStmtPtr.WriteString(attribute)
		sqlStmtPtr.WriteString(" = '') ")
	}

	return sqlArgs, posIndex

}

// convStringArrayToWhereClause does basically the same as convStringToWhereClause
// but creates several WHERE clauses if user input contains several strings (in an array)
func convStringArrayToWhereClause(
	userInputStringArray []string,
	attribute string,
	posIndex int,
	sqlArgs []interface{},
	sqlStmtPtr *strings.Builder,
) ([]interface{}, int) {

	if len(userInputStringArray) == 0 {
		return sqlArgs, posIndex
	}

	// The piece of SQL created here could be something like:
	// AND (UPPER(comp.domain) = UPPER($8) OR UPPER(comp.domain) = UPPER($9))
	posIndex += 1
	if posIndex > 1 {
		sqlStmtPtr.WriteString("AND ")
	}
	if len(userInputStringArray) == 1 {
		sqlStmtPtr.WriteString("UPPER(")
		sqlStmtPtr.WriteString(attribute)
		sqlStmtPtr.WriteString(") = UPPER($")
		sqlStmtPtr.WriteString(strconv.Itoa(posIndex))
		sqlStmtPtr.WriteString(") ")
		sqlArgs = append(sqlArgs, userInputStringArray[0])
	} else {
		sqlStmtPtr.WriteString("(")
		for i, element := range userInputStringArray {
			if i > 0 {
				posIndex += 1
			}
			sqlStmtPtr.WriteString("UPPER(")
			sqlStmtPtr.WriteString(attribute)
			sqlStmtPtr.WriteString(") = UPPER($")
			sqlStmtPtr.WriteString(strconv.Itoa(posIndex))
			sqlStmtPtr.WriteString(") ")
			if i == len(userInputStringArray)-1 {
				sqlStmtPtr.WriteString(") ")
			} else {
				sqlStmtPtr.WriteString("OR ")
			}
			sqlArgs = append(sqlArgs, element)
		}
	}

	return sqlArgs, posIndex
}

// convStringArrayToWhereNotClause does basically the same as convArrayToWhereClause
// but the WHERE clauses should exclude user inputs
func convStringArrayToWhereNotClause(
	userInputStringArray []string,
	attribute string,
	posIndex int,
	sqlArgs []interface{},
	sqlStmtPtr *strings.Builder,
) ([]interface{}, int) {

	if len(userInputStringArray) == 0 {
		return sqlArgs, posIndex
	}

	// The piece of SQL created here could be something like:
	// AND (UPPER(comp.domain) <> UPPER($8) OR UPPER(comp.domain) <> UPPER($9))
	posIndex += 1
	if posIndex > 1 {
		sqlStmtPtr.WriteString("AND ")
	}
	if len(userInputStringArray) == 1 {
		sqlStmtPtr.WriteString("UPPER(")
		sqlStmtPtr.WriteString(attribute)
		sqlStmtPtr.WriteString(") <> UPPER($")
		sqlStmtPtr.WriteString(strconv.Itoa(posIndex))
		sqlStmtPtr.WriteString(") ")
		sqlArgs = append(sqlArgs, userInputStringArray[0])
	} else {
		sqlStmtPtr.WriteString("(")
		for i, element := range userInputStringArray {
			if i > 0 {
				posIndex += 1
			}
			sqlStmtPtr.WriteString("UPPER(")
			sqlStmtPtr.WriteString(attribute)
			sqlStmtPtr.WriteString(") <> UPPER($")
			sqlStmtPtr.WriteString(strconv.Itoa(posIndex))
			sqlStmtPtr.WriteString(") ")
			if i == len(userInputStringArray)-1 {
				sqlStmtPtr.WriteString(") ")
			} else {
				sqlStmtPtr.WriteString("AND ")
			}
			sqlArgs = append(sqlArgs, element)
		}
	}

	return sqlArgs, posIndex
}

// convIntArrayToWhereClause does basically the same as convStringArrayToWhereClause
// but does not apply UPPER because we have integers here
func convIntArrayToWhereClause(
	userInputIntArray []string,
	attribute string,
	posIndex int,
	sqlArgs []interface{},
	sqlStmtPtr *strings.Builder,
) ([]interface{}, int) {

	if len(userInputIntArray) == 0 {
		return sqlArgs, posIndex
	}

	// The piece of SQL created here could be something like:
	// AND (cont_group.group_id = $8 OR cont_group.group_id = $9)
	posIndex += 1
	if posIndex > 1 {
		sqlStmtPtr.WriteString("AND ")
	}
	if len(userInputIntArray) == 1 {
		sqlStmtPtr.WriteString(attribute)
		sqlStmtPtr.WriteString(" = $")
		sqlStmtPtr.WriteString(strconv.Itoa(posIndex))
		sqlStmtPtr.WriteString(" ")
		sqlArgs = append(sqlArgs, userInputIntArray[0])
	} else {
		sqlStmtPtr.WriteString("(")
		for i, element := range userInputIntArray {
			if i > 0 {
				posIndex += 1
			}
			sqlStmtPtr.WriteString(attribute)
			sqlStmtPtr.WriteString(" = $")
			sqlStmtPtr.WriteString(strconv.Itoa(posIndex))
			sqlStmtPtr.WriteString(" ")
			if i == len(userInputIntArray)-1 {
				sqlStmtPtr.WriteString(") ")
			} else {
				sqlStmtPtr.WriteString("OR ")
			}
			sqlArgs = append(sqlArgs, element)
		}
	}

	return sqlArgs, posIndex
}

// convIntArrayToWhereNotClause does basically the same as convIntArrayToWhereNotClause
// but does not apply UPPER because we have integers here
func convIntArrayToWhereNotClause(
	userInputIntArray []string,
	attribute string,
	posIndex int,
	sqlArgs []interface{},
	sqlStmtPtr *strings.Builder,
) ([]interface{}, int) {

	if len(userInputIntArray) == 0 {
		return sqlArgs, posIndex
	}

	// The piece of SQL created here could be something like:
	// AND (cont_group.group_id <> $8 OR comp.domain <> $9)
	posIndex += 1
	if posIndex > 1 {
		sqlStmtPtr.WriteString("AND ")
	}
	if len(userInputIntArray) == 1 {
		sqlStmtPtr.WriteString(attribute)
		sqlStmtPtr.WriteString(" <> $")
		sqlStmtPtr.WriteString(strconv.Itoa(posIndex))
		sqlStmtPtr.WriteString(" ")
		sqlArgs = append(sqlArgs, userInputIntArray[0])
	} else {
		sqlStmtPtr.WriteString("(")
		for i, element := range userInputIntArray {
			if i > 0 {
				posIndex += 1
			}
			sqlStmtPtr.WriteString(attribute)
			sqlStmtPtr.WriteString(" <> $")
			sqlStmtPtr.WriteString(strconv.Itoa(posIndex))
			sqlStmtPtr.WriteString(" ")
			if i == len(userInputIntArray)-1 {
				sqlStmtPtr.WriteString(") ")
			} else {
				sqlStmtPtr.WriteString("AND ")
			}
			sqlArgs = append(sqlArgs, element)
		}
	}

	return sqlArgs, posIndex
}

// runFullSQLReq executes an SQL query with an variable number of arguments and returns
// results in an array.
// Arguments are contained in the sqlArgs array. sqlArgs must be of type []interface{} because
// this is what db.Query() is expecting.
func runFullSQLReq(sqlStmtStr string, sqlArgs []interface{}, w http.ResponseWriter) ([]CompAndContRow, error) {

	var compAndContRows []CompAndContRow
	var err error

	// Connect to db:
	dbinfo := fmt.Sprintf(`host=%s port=%d user=%s password=%s dbname=%s
        sslmode=disable`, remoteHost, remotePort, remoteUser, remotePassword, remoteDbname)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		err = CustErr(err, "DB connection failed\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return compAndContRows, err
	}
	defer db.Close()

	// Executes SQL query using a variable number of arguments contained in the sqlArgs array
	// thanks to the fact that db.Query is a variadic function
	rows, err := db.Query(sqlStmtStr, sqlArgs...)
	if err != nil {
		err = CustErr(err, "SQL query failed.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return compAndContRows, err
	}
	defer rows.Close()

	// For each row, put the results into the compAndContRows array
	for rows.Next() {
		var compAndContRow CompAndContRow
		if err := rows.Scan(
			&compAndContRow.CompId,
			&compAndContRow.CompName,
			&compAndContRow.CompDomain,
			&compAndContRow.CompWebsite,
			&compAndContRow.CompTelephone,
			&compAndContRow.CompFaxNumber,
			&compAndContRow.CompSize,
			&compAndContRow.CompFounded,
			&compAndContRow.CompCreatedOn,
			&compAndContRow.CompUpdatedOn,
			&compAndContRow.CompStreetNumber,
			&compAndContRow.CompRoute,
			&compAndContRow.CompPostalCode,
			&compAndContRow.CompLocality,
			&compAndContRow.CompAdministrativeAreaLevel2,
			&compAndContRow.CompAdministrativeAreaLevel1,
			&compAndContRow.CompCountry,
			&compAndContRow.CompEmail,
			&compAndContRow.CompSocProfURL,
			&compAndContRow.CompType,
			&compAndContRow.CompIndustry,
			&compAndContRow.ContId,
			&compAndContRow.ContGender,
			&compAndContRow.ContFirstName,
			&compAndContRow.ContLastName,
			&compAndContRow.ContJobTitle,
			&compAndContRow.ContTelephone,
			&compAndContRow.ContCreatedOn,
			&compAndContRow.ContUpdatedOn,
			&compAndContRow.ContStreetNumber,
			&compAndContRow.ContRoute,
			&compAndContRow.ContPostalCode,
			&compAndContRow.ContLocality,
			&compAndContRow.ContAdministrativeAreaLevel2,
			&compAndContRow.ContAdministrativeAreaLevel1,
			&compAndContRow.ContCountry,
			&compAndContRow.ContJobFunction,
			&compAndContRow.ContJobLevel,
			&compAndContRow.ContEmail,
			&compAndContRow.ContEmailStatus,
			&compAndContRow.ContEmailCreatedOn,
			&compAndContRow.ContSocProfURL,
			&compAndContRow.ContIndustry,
		); err != nil {
			err = CustErr(err, "A row could not be read from SQL query results.\nNOT stopping here.")
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return compAndContRows, err
		}
		compAndContRows = append(compAndContRows, compAndContRow)
	}

	return compAndContRows, err

}

// runCountSQLReq executes the same query as runFullSQLReq but only for Count
func runCountSQLReq(sqlStmtStr string, sqlArgs []interface{}, w http.ResponseWriter) (int, error) {

	var rowsNb int
	var err error

	// Connect to db:
	dbinfo := fmt.Sprintf(`host=%s port=%d user=%s password=%s dbname=%s
        sslmode=disable`, remoteHost, remotePort, remoteUser, remotePassword, remoteDbname)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		err = CustErr(err, "DB connection failed\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return rowsNb, err
	}
	defer db.Close()

	// Executes SQL query using a variable number of arguments contained in the sqlArgs array
	// thanks to the fact that db.QueryRow is a variadic function
	row := db.QueryRow(sqlStmtStr, sqlArgs...)
	err = row.Scan(&rowsNb)
	if err != nil {
		err = CustErr(err, "SQL query failed.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return rowsNb, err
	}

	return rowsNb, err

}

// buildSQLReq builds incrementally the big SQL query
func buildSQLReq(sqlStmtPtr *strings.Builder, isCount bool, userInput UserInput) []interface{} {

	// This is the hardcoded base of the query.
	// We have lots of rows in the query because of cartesian product:
	// 1 company has multiple multiple emails, multiple linkedin urls.
	// 1 contact has multiple job functions, multiple remote groups, multiple linkedin urls.
	// 1 company also has multiple contacts but this is OK we want to keep them on multiple lines.
	// So we need to remove duplicates and concatenate values above whithin the same row.
	// For that we GROUP BY every values except those mentioned above and use the string_agg() Postgresql function
	// in the SELECT clause for concatenation.
	// The string_agg function concatenates results from multiple rows but we can still have duplicates (I'm not sure
	// why) so we remove them with DISTINCT
	// A simple COUNT used together with group by would give multiple rows with different counts on every
	// row, but the number of rows is correct. So we use the OVER() function in order to count the number of
	// rows returned by the group by. It still give multiple lines with all the same number so we will read
	// the first one only with queryRow.
	sqlStmtPtr.WriteString("SELECT ")
	if isCount {
		sqlStmtPtr.WriteString("COUNT(comp.id) OVER() ")
	} else {
		sqlStmtPtr.WriteString("comp.id, comp.name, comp.domain, comp.website, comp.telephone, comp.faxnumber, comp.size, comp.founded, comp.created_on, comp.updated_on, ")
		sqlStmtPtr.WriteString("comp_ad.street_number, comp_ad.route, comp_ad.postal_code, comp_ad.locality, comp_ad.administrative_area_level_2, comp_ad.administrative_area_level_1, comp_ad.country, ")
		sqlStmtPtr.WriteString("string_agg(DISTINCT companyemail.email,'¤'), ")
		sqlStmtPtr.WriteString("string_agg(DISTINCT comp_soc_prof.url,'¤'), comp_soc_prof.type, comp_soc_prof.industry, ")
		sqlStmtPtr.WriteString("cont.id, cont.gender, cont.first_name, cont.last_name, cont.job_title, cont.telephone, cont.created_on, cont.updated_on, ")
		sqlStmtPtr.WriteString("cont_ad.street_number, cont_ad.route, cont_ad.postal_code, cont_ad.locality, cont_ad.administrative_area_level_2, cont_ad.administrative_area_level_1, cont_ad.country, ")
		sqlStmtPtr.WriteString("string_agg(DISTINCT job_function.name,'¤'), ")
		sqlStmtPtr.WriteString("job_level.name, ")
		sqlStmtPtr.WriteString("cont_email.email, cont_email.status, cont_email.created_on, ")
		sqlStmtPtr.WriteString("string_agg(DISTINCT cont_soc_prof.url,'¤'), cont_soc_prof.industry ")
	}
	sqlStmtPtr.WriteString("FROM company AS comp ")
	sqlStmtPtr.WriteString("LEFT JOIN postal_address AS comp_ad ON comp_ad.id = comp.postal_address_id ")
	sqlStmtPtr.WriteString("LEFT JOIN companyemail ON companyemail.company_id = comp.id ")
	sqlStmtPtr.WriteString("LEFT JOIN companysocialprofile AS comp_soc_prof ON comp_soc_prof.company_id = comp.id ")
	sqlStmtPtr.WriteString("LEFT JOIN prospect AS cont ON cont.company_id = comp.id ")
	sqlStmtPtr.WriteString("LEFT JOIN postal_address AS cont_ad ON cont_ad.id = cont.postal_address_id ")
	sqlStmtPtr.WriteString("LEFT JOIN prospect_job_function_mapping ON prospect_job_function_mapping.prospect_id = cont.id ")
	sqlStmtPtr.WriteString("LEFT JOIN job_function ON job_function.id = prospect_job_function_mapping.job_function_id ")
	sqlStmtPtr.WriteString("LEFT JOIN job_level ON job_level.id = cont.job_level_id ")
	sqlStmtPtr.WriteString("LEFT JOIN prospectemail AS cont_email ON cont_email.id = cont.email_id ")
	sqlStmtPtr.WriteString("LEFT JOIN prospectsocialprofile AS cont_soc_prof ON cont_soc_prof.id = cont.social_profile_id ")
	sqlStmtPtr.WriteString("LEFT JOIN savelistprospectcustomersgroup AS cont_group ON cont_group.prospect_id = cont.id ")
	sqlStmtPtr.WriteString("WHERE ")

	// In order to build query incrementally based on a variable number
	// of parameters, tried first to use NamedArg (https://golang.org/pkg/database/sql/#NamedArg)
	// but not working (while officially Postgresql is supposed to support it) and docs
	// does not help.
	// So I'm building the query with positional arguments ($1, $2, ...) by incrementing
	// dynamically the value. The index values are stored in posIndex.
	// The corresponding value for each index is stored in sqlArgs. Type of sqlArgs is []interface{}
	// because the is what the db.Query() variadic function expects.
	var sqlArgs []interface{}
	var posIndex int

	// Incrementally add WHERE clauses to the SQL query
	sqlArgs, posIndex = convStringToWhereClause(userInput.CompanyCity, "comp_ad.locality", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convStringToWhereClause(userInput.CompanyPostCode, "comp_ad.postal_code", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convStringArrayToWhereClause(userInput.CompanyCountries, "comp_ad.country", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convStringArrayToWhereClause(userInput.CompanyIndustries, "comp_soc_prof.industry", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convStringArrayToWhereClause(userInput.CompanySizes, "comp.size", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convStringArrayToWhereClause(userInput.CompanyTypes, "comp_soc_prof.type", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convBoolToWhereClause(userInput.CompanyHasEmail, "companyemail.email", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convBoolToWhereClause(userInput.CompanyHasPhone, "comp.telephone", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convStringArrayToWhereClause(userInput.CompanyDomains, "comp.domain", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convStringArrayToWhereNotClause(userInput.ExcludedCompanyDomains, "comp.domain", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convStringToWhereClause(userInput.ContactCity, "cont_ad.locality", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convStringToWhereClause(userInput.ContactPostCode, "cont_ad.postal_code", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convStringArrayToWhereClause(userInput.ContactCountries, "cont_ad.country", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convStringArrayToWhereClause(userInput.ContactIndustries, "cont_soc_prof.industry", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convStringToWhereLikeClause(userInput.ContactJobTitle, "cont.job_title", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convStringArrayToWhereClause(userInput.ContactFunctions, "job_function.name", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convStringArrayToWhereClause(userInput.ContactJobLevels, "job_level.name", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convBoolToWhereClause(userInput.ContactHasEmail, "cont_email.email", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convIntArrayToWhereClause(userInput.ContactRemoteAccounts, "cont_group.group_id", posIndex, sqlArgs, sqlStmtPtr)
	sqlArgs, posIndex = convIntArrayToWhereNotClause(userInput.ExcludedContactRemoteAccounts, "cont_group.group_id", posIndex, sqlArgs, sqlStmtPtr)

	// GROUP BY part necessary in order to remove duplicates (used together with string_add() )
	sqlStmtPtr.WriteString("GROUP BY comp.id, comp.name, comp.domain, comp.website, comp.telephone, comp.faxnumber, comp.size, comp.founded, comp.created_on, comp.updated_on, ")
	sqlStmtPtr.WriteString("comp_ad.street_number, comp_ad.route, comp_ad.postal_code, comp_ad.locality, comp_ad.administrative_area_level_2, comp_ad.administrative_area_level_1, comp_ad.country, ")
	sqlStmtPtr.WriteString("comp_soc_prof.type, comp_soc_prof.industry, ")
	sqlStmtPtr.WriteString("cont.id, cont.gender, cont.first_name, cont.last_name, cont.job_title, cont.telephone, cont.created_on, cont.updated_on, ")
	sqlStmtPtr.WriteString("cont_ad.street_number, cont_ad.route, cont_ad.postal_code, cont_ad.locality, cont_ad.administrative_area_level_2, cont_ad.administrative_area_level_1, cont_ad.country, ")
	sqlStmtPtr.WriteString("job_level.name, ")
	sqlStmtPtr.WriteString("cont_email.email, cont_email.status, cont_email.created_on, ")
	sqlStmtPtr.WriteString("cont_soc_prof.industry")

	return sqlArgs

}

// validateUserInput apply validation rule on user input
func validateUserInput(userInputPtr *UserInput) error {

	var err error

	// Check that user input is not empty.
	// Not checked in frontend.
	if userInputPtr.CompanyCity == "" &&
		userInputPtr.CompanyPostCode == "" &&
		len(userInputPtr.CompanyCountries) == 0 &&
		len(userInputPtr.CompanyIndustries) == 0 &&
		len(userInputPtr.CompanySizes) == 0 &&
		len(userInputPtr.CompanyTypes) == 0 &&
		userInputPtr.CompanyHasPhone == 0 &&
		userInputPtr.CompanyHasEmail == 0 &&
		len(userInputPtr.CompanyDomains) == 0 &&
		len(userInputPtr.ExcludedCompanyDomains) == 0 &&
		userInputPtr.ContactCity == "" &&
		userInputPtr.ContactPostCode == "" &&
		len(userInputPtr.ContactCountries) == 0 &&
		len(userInputPtr.ContactIndustries) == 0 &&
		userInputPtr.ContactJobTitle == "" &&
		len(userInputPtr.ContactFunctions) == 0 &&
		len(userInputPtr.ContactJobLevels) == 0 &&
		userInputPtr.ContactHasEmail == 0 &&
		len(userInputPtr.ContactRemoteAccounts) == 0 &&
		len(userInputPtr.ExcludedContactRemoteAccounts) == 0 {

		return errors.New("All search criteria are empty.")

	}

	// Check that strings are strings.
	// Also checked in frontend.
	if _, err := strconv.Atoi(userInputPtr.CompanyCity); err == nil {
		return errors.New("Company City should be text.")
	}
	if _, err := strconv.Atoi(userInputPtr.ContactCity); err == nil {
		return errors.New("Contact City should be text.")
	}
	if _, err := strconv.Atoi(userInputPtr.ContactJobTitle); err == nil {
		return errors.New("Company Job Title should be text.")
	}

	// Check that arrays of strings are arrays of strings
	for _, elem := range userInputPtr.CompanyCountries {
		if _, err := strconv.Atoi(elem); err == nil {
			return errors.New("Company Countries should be text.")
		}
	}
	for _, elem := range userInputPtr.CompanyIndustries {
		if _, err := strconv.Atoi(elem); err == nil {
			return errors.New("Company Industries should be text.")
		}
	}
	for _, elem := range userInputPtr.CompanySizes {
		if _, err := strconv.Atoi(elem); err == nil {
			return errors.New("Company Sizes should be text.")
		}
	}
	for _, elem := range userInputPtr.CompanyTypes {
		if _, err := strconv.Atoi(elem); err == nil {
			return errors.New("Company Types should be text.")
		}
	}
	for _, elem := range userInputPtr.ContactCountries {
		if _, err := strconv.Atoi(elem); err == nil {
			return errors.New("Contact Countries should be text.")
		}
	}
	for _, elem := range userInputPtr.ContactIndustries {
		if _, err := strconv.Atoi(elem); err == nil {
			return errors.New("Contact Industries should be text.")
		}
	}
	for _, elem := range userInputPtr.ContactFunctions {
		if _, err := strconv.Atoi(elem); err == nil {
			return errors.New("Contact Functions should be text.")
		}
	}
	for _, elem := range userInputPtr.ContactJobLevels {
		if _, err := strconv.Atoi(elem); err == nil {
			return errors.New("Contact Job Levels should be text.")
		}
	}

	// Check that arrays of ints are arrays of ints
	for _, elem := range userInputPtr.ContactRemoteAccounts {
		if _, err := strconv.Atoi(elem); err != nil {
			return errors.New("Contact Remote Accounts ids should be integers.")
		}
	}
	for _, elem := range userInputPtr.ExcludedContactRemoteAccounts {
		if _, err := strconv.Atoi(elem); err != nil {
			return errors.New("Contact Remote Accounts ids should be integers.")
		}
	}

	// No need to check if int here because Unmarshalling would have broken if
	// not an int
	if userInputPtr.CompanyHasPhone < 0 || userInputPtr.CompanyHasPhone > 2 {
		return errors.New("Company Has Phone should be integer: 1, 2, or 0.")
	}
	if userInputPtr.CompanyHasEmail < 0 || userInputPtr.CompanyHasEmail > 2 {
		return errors.New("Company Has Email should be integer: 1, 2, or 0.")
	}
	if userInputPtr.ContactHasEmail < 0 || userInputPtr.ContactHasEmail > 2 {
		return errors.New("Contact Has Email should be integer: 1, 2, or 0.")
	}

	return err

}

// cleanUserInput removes spaces, tabs, newlines at the beginning and end of inputs.
// Converting to uppercase is not addressed here but directly in SQL with UPPER().
func cleanUserInput(userInputPtr *UserInput) {

	// The TrimeSpace function does everything
	userInputPtr.CompanyCity = strings.TrimSpace(userInputPtr.CompanyCity)

	userInputPtr.CompanyPostCode = strings.TrimSpace(userInputPtr.CompanyPostCode)

	// For array inputs, same principle but must parse the whole array
	// and put results in a new array
	var newCompanyDomains []string
	for _, domain := range userInputPtr.CompanyDomains {
		domain = strings.TrimSpace(domain)
		newCompanyDomains = append(newCompanyDomains, domain)
	}
	userInputPtr.CompanyDomains = newCompanyDomains

	var newExcludedCompanyDomains []string
	for _, domain := range userInputPtr.ExcludedCompanyDomains {
		domain = strings.TrimSpace(domain)
		newExcludedCompanyDomains = append(newExcludedCompanyDomains, domain)
	}
	userInputPtr.ExcludedCompanyDomains = newExcludedCompanyDomains

	userInputPtr.ContactCity = strings.TrimSpace(userInputPtr.ContactCity)

	userInputPtr.ContactPostCode = strings.TrimSpace(userInputPtr.ContactPostCode)

	userInputPtr.ContactJobTitle = strings.TrimSpace(userInputPtr.ContactJobTitle)

	var newContactRemoteAccounts []string
	for _, remoteAccount := range userInputPtr.ContactRemoteAccounts {
		remoteAccount = strings.TrimSpace(remoteAccount)
		newContactRemoteAccounts = append(newContactRemoteAccounts, remoteAccount)
	}
	userInputPtr.ContactRemoteAccounts = newContactRemoteAccounts

	var newExcludedContactRemoteAccounts []string
	for _, remoteAccount := range userInputPtr.ExcludedContactRemoteAccounts {
		remoteAccount = strings.TrimSpace(remoteAccount)
		newExcludedContactRemoteAccounts = append(newExcludedContactRemoteAccounts, remoteAccount)
	}
	userInputPtr.ExcludedContactRemoteAccounts = newExcludedContactRemoteAccounts

}

// compressCSV turns the CSV file into a .zip archive
func compressCSV() error {

	newfile, err := os.Create(returnedArchiveName)
	if err != nil {
		return err
	}
	defer newfile.Close()

	zipWriter := zip.NewWriter(newfile)
	defer zipWriter.Close()

	zipfile, err := os.Open(returnedCSVName)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	info, err := zipfile.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, zipfile)
	if err != nil {
		return err
	}

	return err

}

// createCSV puts results returned from DB into a CSV file.
// Write CSV file to disk in order to avoid RAM problems.
func createCSV(compAndContRows []CompAndContRow) error {

	csvFile, err := os.Create(returnedCSVName)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	csvWriter.Comma = ';'
	defer csvWriter.Flush()

	csvFirstRow := []string{
		"Company Id",
		"Company Name",
		"Company Domain",
		"Company Website",
		"Company Telephone",
		"Company Fax Number",
		"Company Size",
		"Company Founded",
		"Company Street Number",
		"Company Route",
		"Company Postal Code",
		"Company Locality",
		"Company Admin Area Level 1",
		"Company Admin Area Level 2",
		"Company Country",
		"Company Email",
		"Company Social Profile URL",
		"Company Type",
		"Company Industry",
		"Company Creation Date",
		"Company Update Date",
		"Contact Id",
		"Contact Gender",
		"Contact First Name",
		"Contact Last Name",
		"Contact Job Title",
		"Contact Job Function",
		"Contact Job Level",
		"Contact Telephone",
		"Contact Street Number",
		"Contact Route",
		"Contact Postal Code",
		"Contact Locality",
		"Contact Admin Area Level 1",
		"Contact Admin Area Level 2",
		"Contact Country",
		"Contact Email",
		"Contact Email Status",
		"Contact Email Creation Date",
		"Contact Social Profile URL",
		"Contact Industry",
		"Contact Creation Date",
		"Contact Update Date",
	}
	csvWriter.Write(csvFirstRow)

	for _, row := range compAndContRows {
		csvRow := []string{
			row.CompId,
			row.CompName.String,
			row.CompDomain.String,
			row.CompWebsite.String,
			row.CompTelephone.String,
			row.CompFaxNumber.String,
			row.CompSize.String,
			row.CompFounded.String,
			row.CompStreetNumber.String,
			row.CompRoute.String,
			row.CompPostalCode.String,
			row.CompLocality.String,
			row.CompAdministrativeAreaLevel2.String,
			row.CompAdministrativeAreaLevel1.String,
			row.CompCountry.String,
			row.CompEmail.String,
			row.CompSocProfURL.String,
			row.CompType.String,
			row.CompIndustry.String,
			row.CompCreatedOn.String,
			row.CompUpdatedOn.String,
			row.ContId.String,
			row.ContGender.String,
			row.ContFirstName.String,
			row.ContLastName.String,
			row.ContJobTitle.String,
			row.ContJobFunction.String,
			row.ContJobLevel.String,
			row.ContTelephone.String,
			row.ContStreetNumber.String,
			row.ContRoute.String,
			row.ContPostalCode.String,
			row.ContLocality.String,
			row.ContAdministrativeAreaLevel2.String,
			row.ContAdministrativeAreaLevel1.String,
			row.ContCountry.String,
			row.ContEmail.String,
			row.ContEmailStatus.String,
			row.ContEmailCreatedOn.String,
			row.ContSocProfURL.String,
			row.ContIndustry.String,
			row.ContCreatedOn.String,
			row.ContUpdatedOn.String,
		}
		csvWriter.Write(csvRow)
	}

	return err

}

// sendResultsByEmail sends the .zip archive by email to a person defined
// in env variable.
// Here we're using a nice little library for attachments.
func sendResultsByEmail() error {

	m := gomail.NewMessage()

	m.SetHeader("From", "admin@example.com")
	m.SetHeader("To", getUserEmail())
	m.SetHeader("Subject", "Database extraction done !")
	m.SetBody("text/html", "Please find enclosed the extracted results.")
	m.Attach(returnedArchiveName)

	d := gomail.NewPlainDialer("smtp.example.com", 587, "admin@example.com", "password")
	err := d.DialAndSend(m)

	return err

}

// ReturnCompaniesAndContacts loads companies and associated contacts from db
// based on user criteria and renders results in JSON response to frontend
func ReturnCompaniesAndContacts(w http.ResponseWriter, r *http.Request) {

	// Retrieve JSON data containing user inputs.
	// Cannot use r.PostFormValue here
	//(https://golang.org/pkg/net/http/#Request.PostFormValue)
	// because expects form encoded data while Axios can only
	// send json encoded data (impossible to change it in Axios for
	// the moment.)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err = CustErr(err, "Cannot read request body.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Println(string(body))

	// Store JSON data in a userInput struct
	var userInput UserInput
	err = json.Unmarshal(body, &userInput)
	if err != nil {
		err = CustErr(err, "Cannot unmarshall json.\nStopping here.")
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Apply cleaning and validation rules to all the user inputs.
	// userInput is not passed by value but by pointer because heavy struct in memory
	// http://goinbigdata.com/golang-pass-by-pointer-vs-pass-by-value/
	err = validateUserInput(&userInput)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cleanUserInput(&userInput)

	var returnedJson []byte

	// If user only ask a count we launch a special count sql request and only return the nb of rows.
	// If user ask for the full results we return everything either in json or by compressed csv by email
	// depending on the size.
	switch userInput.Step {

	case "count":

		// Initialize the count SQL statement that will be created incrementally
		var sqlStmtCnt strings.Builder

		// Build the SQL request
		// sqlStmt is not passed by value but by pointer because compulsory for a strings.Builder
		sqlArgs := buildSQLReq(&sqlStmtCnt, true, userInput)
		// Convert SQL statement to string
		sqlStmtCntStr := sqlStmtCnt.String()

		log.Println(sqlStmtCntStr)

		// Run the SQL query
		var countRes CountRes
		countRes.rowsNb, err = runCountSQLReq(sqlStmtCntStr, sqlArgs, w)
		if err != nil {
			return
		}

		// Turn struct into a proper JSON response:
		returnedJson, err = json.Marshal(countRes.rowsNb)
		if err != nil {
			err = CustErr(err, "Could not not marshall to JSON.\nStopping here.")
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

	case "full":

		// Initialize the full SQL statement that will be created incrementally
		var sqlStmtFull strings.Builder

		sqlArgs := buildSQLReq(&sqlStmtFull, false, userInput)
		sqlStmtFullStr := sqlStmtFull.String()

		log.Println(sqlStmtFullStr)

		compAndContRows, err := runFullSQLReq(sqlStmtFullStr, sqlArgs, w)
		if err != nil {
			return
		}

		// Get nb of rows returned by query
		rowsNb := len(compAndContRows)

		// If no result found, stop here
		if rowsNb == 0 {
			log.Println("No result found\nStopping here.")
			http.Error(w, "No result found", http.StatusNotFound)
			return
		}

		if rowsNb > 5000 { // Send results in a compressed csv by email because too big

			// Put results in a CSV file
			err = createCSV(compAndContRows)
			if err != nil {
				err = CustErr(err, "Could not create CSV.\nStopping here.")
				log.Println(err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			// Compress the CSV file to .zip
			err = compressCSV()
			if err != nil {
				err = CustErr(err, "Could not compress CSV.\nStopping here.")
				log.Println(err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			// Send .zip archive by email
			err = sendResultsByEmail()
			if err != nil {
				err = CustErr(err, "Could not send results by email.\nStopping here.")
				log.Println(err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			// Remove CSV
			err = os.Remove(returnedCSVName)
			if err != nil {
				err = CustErr(err, "Could not delete CSV.\nNOT stopping here.")
				log.Println(err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			// Remove .zip archive
			err = os.Remove(returnedArchiveName)
			if err != nil {
				err = CustErr(err, "Could not remove archive.\nNOT stopping here.")
				log.Println(err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}

			// Tell frontend that not returning a json but sent by email.
			http.Error(w, "The request returned too many lines so results have been sent by email.", http.StatusNoContent)

		} else { // Send results in json

			// Turn struct into a proper JSON response:
			returnedJson, err = json.Marshal(compAndContRows)
			if err != nil {
				err = CustErr(err, "Could not not marshall to JSON.\nStopping here.")
				log.Println(err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

		}
	}

	// Set a custom JSON header and send response to client:
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", returnedJson)

}
