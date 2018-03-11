<template>
  <div>
    <v-card class="pb-3">
      <v-card-title primary-title>
        <h3 class="headline mb-0">Retrieve data stored in the email_checked_by_john table</h3>
      </v-card-title>
      <v-container fluid>
        <v-layout row>
          <v-flex xs8>
            <v-text-field :rules="[rules.missionNumberIsInteger]" v-model="missionNumber" @keyup.enter="sendMissionNumber" label="Mission number"></v-text-field>
          </v-flex>
          <v-flex xs4>
            <v-btn v-if="activateMissionNumberButton" color="primary" @click="sendMissionNumber">Send</v-btn>
            <v-btn v-else light disabled>Send</v-btn>
          </v-flex>
        </v-layout>
      </v-container>
      <v-alert color="success" icon="check_circle" :value="showRowsNb" class="mb-4">Request successfull.
        <br> Number of results: <b>{{ rowsNb }}</b></v-alert>
      <p v-if="showShowResultsButton == true">
        <v-btn color="primary" @click="showResults = true; showShowResultsButton = false">Show Results</v-btn>
      </p>
      <div v-if="showGenerateCSV">
        <p>
          <v-btn color="primary" @click="generateCSV">Export results to CSV</v-btn>
        </p>
      </div>
      <v-progress-circular v-if="showLoader" indeterminate :size="50" color="primary"></v-progress-circular>
      <div v-if="showResults">
        <v-card-title>
          <v-spacer></v-spacer>
          <v-text-field append-icon="search" label="Search" single-line hide-details v-model="search"></v-text-field>
        </v-card-title>
        <v-data-table :headers="headers" :items="rows" :search="search" class="elevation-1">
          <template slot="items" slot-scope="props">
            <td>{{ props.item.id }}</td>
            <td>{{ props.item.missionnumber }}</td>
            <td>{{ props.item.firstname }}</td>
            <td>{{ props.item.lastname }}</td>
            <td>{{ props.item.emaildomain }}</td>
            <td>{{ props.item.email }}</td>
            <td>{{ props.item.contactfromc2lid }}</td>
            <td>{{ props.item.qevresult }}</td>
            <td>{{ props.item.qevreason }}</td>
            <td>{{ props.item.qevdisposable }}</td>
            <td>{{ props.item.qevacceptall }}</td>
            <td>{{ props.item.qevrole }}</td>
            <td>{{ props.item.qevfree }}</td>
            <td>{{ props.item.qevsafetosend }}</td>
            <td>{{ props.item.qevdidyoumean }}</td>
            <td>{{ props.item.qevsuccess }}</td>
            <td>{{ props.item.qevmessage }}</td>
            <td>{{ props.item.apicheckdatetime }}</td>
            <td>{{ props.item.manualemailsendingdatetime }}</td>
            <td>{{ props.item.manualemailerrorresponsedatetime }}</td>
            <td>{{ props.item.contactid }}</td>
          </template>
        </v-data-table>
      </div>
      <v-alert color="error" icon="warning" :value="showErrorMessage">Error: {{ errorMessage }}</v-alert>
    </v-card>
  </div>
</template>
<script>
/* beautify preserve:start */

import {HTTP} from '../http-constants'

export default {
  name: 'GetEmailsCheckedByJohn',
  data () {
    return {
      missionNumber: '',
      rules: {
        missionNumberIsInteger: (value) => {
          if (isNaN(value) || value === '') {
            this.activateMissionNumberButton = false
            return 'Please enter an integer'
          } else {
            this.activateMissionNumberButton = true
            return true
          }
        }
      },
      activateMissionNumberButton: false,
      showLoader: false,
      showResults: false,
      showShowResultsButton: false,
      showRowsNb: false,
      rowsNb: '',
      rows: [],
      showGenerateCSV: false,
      errors: [],
      errorMessage: '',
      showErrorMessage: false,
      search: '',
      headers: [
          { text: 'Id', value: 'id' },
          { text: 'Mission Number', value: 'missionnumber' },
          { text: 'First Name', value: 'firstname' },
          { text: 'Last Name', value: 'lastname' },
          { text: 'Email Domain', value: 'emaildomain' },
          { text: 'Email', value: 'email' },
          { text: 'Contact From C2L Id', value: 'contactfromc2lid' },
          { text: 'QEV Result', value: 'qevresult' },
          { text: 'QEV Reason', value: 'qevreason' },
          { text: 'QEV Disposable', value: 'qevdisposable' },
          { text: 'QEV AcceptAll', value: 'qevacceptall' },
          { text: 'QEV Role', value: 'qevrole' },
          { text: 'QEV Free', value: 'qevfree' },
          { text: 'QEV Safe To Send', value: 'qevsafetosend' },
          { text: 'QEV Did You Mean', value: 'qevdidyoumean' },
          { text: 'QEV Success', value: 'qevsuccess' },
          { text: 'QEV Message', value: 'qevmessage' },
          { text: 'API Check Date Time', value: 'apicheckdatetime' },
          { text: 'Manual Email Sending Datetime', value: 'manualemailsendingdatetime' },
          { text: 'Manual Email Error Response Datetime', value: 'manualemailerrorresponsedatetime' },
          { text: 'Contact Id', value: 'contactid' }
      ]
    }
  },
  methods: {
    // sendMissionNumber contacts the backend API and gets
    // JSON data based on a mission number.
    // If no mission number matches the request, a 404 page is
    // returned so we catch an error here.
    sendMissionNumber: function () {
      if (this.activateMissionNumberButton) {
        this.showResults = false
        this.showShowResultsButton = false
        this.showRowsNb = false
        this.showGenerateCSV = false
        this.showLoader = true
        HTTP.get(`get-emails-checked-by-john/mission-number/` + this.missionNumber)
        .then(response => {
          // JSON responses are automatically parsed.
          this.rows = response.data
          this.rowsNb = this.rows.length
          this.showLoader = false
          this.showShowResultsButton = true
          this.showRowsNb = true
          this.showGenerateCSV = true
          this.errors = []
          this.errorMessage = ''
          this.showErrorMessage = false
        })
        .catch(e => {
          this.errors = e
          this.errorMessage = e.response.data
          this.showErrorMessage = true
          this.showLoader = false
          this.showResults = false
          this.showShowResultsButton = false
          this.showRowsNb = false
          this.showGenerateCSV = false
        })
      }
    },
    // generateCSV creates a CSV from data in this.rows
    // and makes it downloadable.
    // Found the CSV generation code here:
    // https://stackoverflow.com/questions/14964035/how-to-export-javascript-array-info-to-csv-on-client-side
    generateCSV: function () {
      // Set filetype + csv headers in the same line:
      var csvArray = ['data:text/csv;charset=utf-8,Id;MissionNumber;FirstName;LastName;EmailDomain;Email;ContactFromC2LId;QEVResult;QEVReason;QEVDisposable;QEVAcceptAll;QEVRole;QEVFree;QEVSafeToSend;QEVDidYouMean;QEVSuccess;QEVMessage;APICheckDateTime;ManualEmailSendingDatetime;ManualEmailErrorResponseDatetime;ContactId']
      // Create a new line in CSV for every row:
      this.rows.forEach(function (row) {
        var csvRow = row['id'] + ';' +
         row['missionnumber'] + ';' +
         row['firstname'] + ';' +
         row['lastname'] + ';' +
         row['emaildomain'] + ';' +
         row['email'] + ';' +
         row['contactfromc2lid'] + ';' +
         row['qevresult'] + ';' +
         row['qevreason'] + ';' +
         row['qevdisposable'] + ';' +
         row['qevacceptall'] + ';' +
         row['qevrole'] + ';' +
         row['qevfree'] + ';' +
         row['qevsafetosend'] + ';' +
         row['qevdidyoumean'] + ';' +
         row['qevsuccess'] + ';' +
         row['qevmessage'] + ';' +
         row['apicheckdatetime'] + ';' +
         row['manualemailsendingdatetime'] + ';' +
         row['manualemailerrorresponsedatetime'] + ';' +
         row['contactid']
        csvArray.push(csvRow)
      })
      // Add \r\n at the end of each line:
      var csvContent = csvArray.join('\r\n')
      // Create the browser compatible CSV file:
      var encodedUri = encodeURI(csvContent)
      var link = document.createElement('a')
      link.setAttribute('href', encodedUri)
      link.setAttribute('download', 'extract_from_email_checked_by_john.csv')
      document.body.appendChild(link)
      link.click()
    }
  }
}

/* beautify preserve:end */
</script>
