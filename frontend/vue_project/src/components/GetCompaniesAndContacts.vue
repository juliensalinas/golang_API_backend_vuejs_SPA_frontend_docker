<template>
  <div>
    <v-card color="blue-grey lighten-5" class="ml-3 mr-3 mb-3 pb-3">
      <v-card-title primary-title>
        <h3 class="headline mb-0">Search companies and associated contacts based on various criteria</h3>
      </v-card-title>
      <v-container fluid>
        <!-- Error/warning display -->
        <v-alert type="error" class="mt-4" :value="showError">{{ errorMessage }}</v-alert>
        <v-alert type="warning" class="mt-4" :value="showWarning">{{ warningMessage }}</v-alert>
        <!-- Form containing all inputs -->
        <v-form v-show="showForm" v-model="formIsValid" ref="form">
          <v-card elevation-2 class="mb-5 pr-2 pl-2 pb-4">
            <v-card-title primary-title>
              <h3 class="title">Companies</h3>
            </v-card-title>
            <v-layout row>
              <v-flex xs6>
                <!-- Text inputs can be validated by clicking enter in addition to the Send button -->
                <v-text-field v-model="companyCity" :rules="[rules.isText]" @keyup.enter="countResults" label="Company City"></v-text-field>
              </v-flex>
              <v-flex xs6>
                <v-text-field v-model="companyPostCode" @keyup.enter="countResults" label="Company Postal Code"></v-text-field>
              </v-flex>
            </v-layout>
            <v-layout row>
              <v-flex xs6>
                <!-- Select inputs load their content dynamically by calling an external API. In case it takes a lot of time, we have hardcoded multiple default values -->
                <v-select label="Company Countries" :loading="countriesAreLoading" :items="countries" v-model="companyCountries" multiple chips hint="Select one or several countries" persistent-hint></v-select>
              </v-flex>
              <v-flex xs6>
                <v-select label="Company Industries" :loading="companyIndustriesAreLoading" :items="companyIndustries" v-model="companyIndustriesSelected" multiple chips hint="Select one or several industries" persistent-hint></v-select>
              </v-flex>
            </v-layout>
            <v-layout row>
              <v-flex xs6>
                <v-select label="Company Sizes" :loading="companySizesAreLoading" :items="companySizes" v-model="companySizesSelected" multiple chips hint="Select one or several sizes" persistent-hint></v-select>
              </v-flex>
              <v-flex xs6>
                <v-select label="Company Types" :loading="companyTypesAreLoading" :items="companyTypes" v-model="companyTypesSelected" multiple chips hint="Select one or several types" persistent-hint></v-select>
              </v-flex>
            </v-layout>
            <v-layout row>
              <v-flex xs6>
                <!-- Very simple radio group -->
                <v-radio-group v-model="companyHasPhone">
                  <v-radio :key="2" label="Company has a phone number" :value="2"></v-radio>
                  <v-radio :key="1" label="Company does not have a phone number" :value="1"></v-radio>
                  <v-radio :key="0" label="Does not matter" :value="0"></v-radio>
                </v-radio-group>
              </v-flex>
              <v-flex xs6>
                <v-radio-group v-model="companyHasEmail">
                  <v-radio :key="2" label="Company has an email" :value="2"></v-radio>
                  <v-radio :key="1" label="Company does not have an email" :value="1"></v-radio>
                  <v-radio :key="0" label="Does not matter" :value="0"></v-radio>
                </v-radio-group>
              </v-flex>
            </v-layout>
            <v-layout row>
              <!-- Cannot make a nice file input field with Vuetify so kept the simplest solution.  -->
              <!-- Could also make a fake button and hide the hugly button. -->
              <!-- File is not sent to server but loaded into a javascript array locally and then this array is sent via POST. -->
              <!-- File loading is done right after clicking the file (no need of an "upload button"). -->
              <v-flex xs6 class="mr-1" elevation-2>
                <p>Domain names to be <b>included</b></p>
                <p>Upload CSV file containing one column of domain names and no header</p>
                <form enctype="multipart/form-data">
                  <input ref="csv1" type="file" @change="getCompanyDomainsCSV" />
                </form>
              </v-flex>
              <v-flex xs6 class="ml-1" elevation-2>
                <p>Domain names to be <b>excluded</b></p>
                <p>Upload CSV file containing one column of domain names and no header</p>
                <form enctype="multipart/form-data">
                  <input ref="csv2" type="file" @change="getExcludedCompanyDomainsCSV" />
                </form>
              </v-flex>
            </v-layout>
          </v-card>
          <v-card elevation-2 class="mb-5 pr-2 pl-2 pb-4">
            <v-card-title primary-title>
              <h3 class="title">Contacts</h3>
            </v-card-title>
            <v-layout row>
              <v-flex xs6>
                <v-text-field v-model="contactPostCode" @keyup.enter="countResults" label="Contact Postal Code"></v-text-field>
              </v-flex>
              <v-flex xs6>
                <v-select label="Contact Countries" :loading="countriesAreLoading" :items="countries" v-model="contactCountries" multiple chips hint="Select one or several countries" persistent-hint></v-select>
              </v-flex>
            </v-layout>
            <v-layout row>
              <v-flex xs6>
                <v-select label="Contact Industries" :loading="contactIndustriesAreLoading" :items="contactIndustries" v-model="contactIndustriesSelected" multiple chips hint="Select one or several industries" persistent-hint></v-select>
              </v-flex>
              <v-flex xs6>
                <v-text-field v-model="contactJobTitle" :rules="[rules.isText]" @keyup.enter="countResults" label="Contact Job Title"></v-text-field>
              </v-flex>
            </v-layout>
            <v-layout row>
              <v-flex xs6>
                <v-select label="Contact Functions" :loading="contactFunctionsAreLoading" :items="contactFunctions" v-model="contactFunctionsSelected" multiple chips hint="Select one or several functions" persistent-hint></v-select>
              </v-flex>
              <v-flex xs6>
                <v-select label="Contact Job levels" :loading="contactLevelsAreLoading" :items="contactLevels" v-model="contactLevelsSelected" multiple chips hint="Select one or several job levels" persistent-hint></v-select>
              </v-flex>
            </v-layout>
            <v-layout row>
              <v-flex xs6>
                <v-radio-group v-model="contactHasEmail">
                  <v-radio :key="2" label="Contact has an email" :value="2"></v-radio>
                  <v-radio :key="1" label="Contact does not have an email" :value="1"></v-radio>
                  <v-radio :key="0" label="Does not matter" :value="0"></v-radio>
                </v-radio-group>
              </v-flex>
            </v-layout>
            <v-layout>
              <v-flex xs6 class="mr-1" elevation-2>
                <p>Remote accounts ids to be <b>included</b></p>
                <p>Upload CSV file containing one column of Remote account ids and no header</p>
                <form enctype="multipart/form-data">
                  <input ref="csv3" type="file" @change="getContactRemoteAccountsCSV" />
                </form>
              </v-flex>
              <v-flex xs6 class="ml-1" elevation-2>
                <p>Remote accounts ids to be <b>excluded</b></p>
                <p>Upload CSV file containing one column of Remote account ids and no header</p>
                <form enctype="multipart/form-data">
                  <input ref="csv4" type="file" @change="getExcludedContactRemoteAccountsCSV" />
                </form>
              </v-flex>
            </v-layout>
          </v-card>
          <v-layout row>
            <!-- Send data to server. -->
            <!-- If one of the inputs is not valid, the button is greyed. -->
            <!-- If no input is filled, the form should not be valid and the button shoudl be greyed but -->
            <!-- did not manage to do it... A validation is done on the server side for this special case anyway.-->
            <!-- Either choose to get a count, or get the full results of the request. -->
            <v-btn class="mx-auto" :disabled="!formIsValid" color="primary" @click="countResults">Count results</v-btn>
            <v-btn class="mx-auto" :disabled="!formIsValid" color="primary" @click="getFullResults">Get all results</v-btn>
            <!-- Clear all fields of the form -->
            <v-btn class="mx-auto" @click="clearForm">Clear</v-btn>
          </v-layout>
        </v-form>
        <!-- Display loader while data is loaded from server. -->
        <!-- v-show rather than v-if is important here because with v-if we never see the loader. I think the reason is that memory is used 100% for data table rendering and cannot render loader. -->
        <v-progress-circular v-show="showLoader" indeterminate :size="50" color="primary" class="mt-3"></v-progress-circular>
        <v-card v-if="showResultsRowsNb" class="mt-3 pl-2 pr-2 pt-2 pb-3">
          <!-- Count the number of results returned. -->
          <v-alert type="success" :value="showResultsRowsNb">
            Request successful.
            <br> Number of results: <b>{{ resultsRowsNb }}</b>
          </v-alert>
          <!-- Show the results in a data table with a search form and pagination filter. -->
          <!-- v-show rather than v-if is important here because with v-if the table takes a long time to render. -->
          <div v-show="showResults" class="mt-3 pr-2 pl-2">
            <v-card-title>
              <v-spacer></v-spacer>
              <v-text-field append-icon="search" label="Search" single-line hide-details v-model="resultsSearch"></v-text-field>
            </v-card-title>
            <v-data-table :headers="resultsHeaders" :items="resultsRows" :search="resultsSearch" class="elevation-1">
              <template slot="items" slot-scope="props">
                <td>{{ props.item.compId }}</td>
                <td>{{ props.item.compName }}</td>
                <td>{{ props.item.compDomain }}</td>
                <td>{{ props.item.compWebsite }}</td>
                <td>{{ props.item.compTelephone }}</td>
                <td>{{ props.item.compFaxNumber }}</td>
                <td>{{ props.item.compSize }}</td>
                <td>{{ props.item.compFounded }}</td>
                <td>{{ props.item.compStreetNumber }}</td>
                <td>{{ props.item.compRoute }}</td>
                <td>{{ props.item.compPostalCode }}</td>
                <td>{{ props.item.compLocality }}</td>
                <td>{{ props.item.compAdministrativeAreaLevel1 }}</td>
                <td>{{ props.item.compAdministrativeAreaLevel2 }}</td>
                <td>{{ props.item.compCountry }}</td>
                <td>{{ props.item.compEmail }}</td>
                <td>{{ props.item.compSocProfURL }}</td>
                <td>{{ props.item.compType }}</td>
                <td>{{ props.item.compIndustry }}</td>
                <td>{{ props.item.compTecontIdlephone }}</td>
                <td>{{ props.item.compCreatedOn }}</td>
                <td>{{ props.item.compUpdatedOn }}</td>
                <td>{{ props.item.contGender }}</td>
                <td>{{ props.item.contFirstName }}</td>
                <td>{{ props.item.contLastName }}</td>
                <td>{{ props.item.contJobTitle }}</td>
                <td>{{ props.item.contJobFunction }}</td>
                <td>{{ props.item.contJobLevel }}</td>
                <td>{{ props.item.contTelephone }}</td>
                <td>{{ props.item.contStreetNumber }}</td>
                <td>{{ props.item.contRoute }}</td>
                <td>{{ props.item.contPostalCode }}</td>
                <td>{{ props.item.contLocality }}</td>
                <td>{{ props.item.contAdministrativeAreaLevel1 }}</td>
                <td>{{ props.item.contAdministrativeAreaLevel2 }}</td>
                <td>{{ props.item.contCountry }}</td>
                <td>{{ props.item.contEmail }}</td>
                <td>{{ props.item.contEmailStatus }}</td>
                <td>{{ props.item.contEmailCreatedOn }}</td>
                <td>{{ props.item.contSocProfURL }}</td>
                <td>{{ props.item.contIndustry }}</td>
                <td>{{ props.item.contCreatedOn }}</td>
                <td>{{ props.item.contUpdatedOn }}</td>
              </template>
            </v-data-table>
          </div>
          <v-layout class="mt-2" row>
            <!-- Decide to show the results, a export as a CSV, or get the full results if we are only counting results untill now -->
            <v-btn v-if="showShowResultsBtn" class="mx-auto" color="primary" @click="showResults = true; showShowResultsBtn = false">Show Results</v-btn>
            <v-btn v-if="showGenerateCSV" class="mx-auto" color="primary" @click="generateCSV">Export results to CSV</v-btn>
            <v-btn v-if="showGetFullResultsBtn" class="mx-auto" color="primary" @click="getFullResults">Get all results</v-btn>
          </v-layout>
        </v-card>
        <!-- Decide to make a new search (does not clear the intial form) -->
        <v-card v-if="showNewSearchBtn" class="mt-3 pl-2 pr-2 pt-3 pb-3">
          <v-btn class="mx-auto" color="primary" @click="newSearch">New search</v-btn>
        </v-card>
      </v-container>
    </v-card>
  </div>
</template>
<script>
/* beautify preserve:start */
import {HTTP} from '../http-constants'
import axios from 'axios'

export default {
  name: 'GetCompaniesAndContacts',
  data () {
    return {
      companyCity: '',
      companyPostCode: '',
      companyCountries: [],
      // Countries list are common to company and contact
      countries: ['France', 'United States', 'Spain', 'Italy'],
      countriesAreLoading: true,
      companyIndustries: ['Pharmaceuticals', 'Machinery', 'Automotive'],
      companyIndustriesSelected: [],
      companyIndustriesAreLoading: true,
      companySizes: ['1-10 employees', '11-50 employees', '51-200 employees'],
      companySizesSelected: [],
      companySizesAreLoading: true,
      companyTypes: ['Public Company', 'Entreprise individuelle', 'Educational'],
      companyTypesSelected: [],
      companyTypesAreLoading: true,
      companyHasPhone: 0,
      companyHasEmail: 0,
      companyDomains: [],
      excludedCompanyDomains: [],
      contactCity: '',
      contactPostCode: '',
      contactCountries: [],
      contactIndustries: ['Pharmaceuticals', 'Machinery', 'Automotive'],
      contactIndustriesSelected: [],
      contactIndustriesAreLoading: true,
      contactJobTitle: '',
      contactFunctions: ['Support', 'Sales', 'Research'],
      contactFunctionsSelected: [],
      contactFunctionsAreLoading: true,
      contactLevels: ['Contributor', 'Director', 'Executive'],
      contactLevelsSelected: [],
      contactLevelsAreLoading: true,
      contactHasEmail: 0,
      contactRemoteAccounts: [],
      excludedContactRemoteAccounts: [],
      formIsValid: false,
      showForm: true,
      showResults: false,
      showShowResultsBtn: false,
      showResultsRowsNb: false,
      showLoader: false,
      showError: false,
      showWarning: false,
      showNewSearchBtn: false,
      showGenerateCSV: false,
      showGetFullResultsBtn: false,
      errorMessage: '',
      warningMessage: '',
      resultsRowsNb: 0,
      resultsRows: [],
      resultsSearch: '',
      step: 'count',
      // Text is the column name, and value is the name of data in resultsRows
      resultsHeaders: [
        { text: 'Company Id', value: 'compId' },
        { text: 'Company Name', value: 'compName' },
        { text: 'Company Domain', value: 'compDomain' },
        { text: 'Company Website', value: 'compWebsite' },
        { text: 'Company Telephone', value: 'compTelephone' },
        { text: 'Company Fax Number', value: 'compFaxNumber' },
        { text: 'Company Size', value: 'compSize' },
        { text: 'Company Founded', value: 'compFounded' },
        { text: 'Company Street Number', value: 'compStreetNumber' },
        { text: 'Company Route', value: 'compRoute' },
        { text: 'Company Postal Code', value: 'compPostalCode' },
        { text: 'Company Locality', value: 'compLocality' },
        { text: 'Company Admin Area Level 2', value: 'compAdministrativeAreaLevel2' },
        { text: 'Company Admin Area Level 1', value: 'compAdministrativeAreaLevel1' },
        { text: 'Company Country', value: 'compCountry' },
        { text: 'Company Email', value: 'compEmail' },
        { text: 'Company Social Profile URL', value: 'compSocProfURL' },
        { text: 'Company Type', value: 'compType' },
        { text: 'Company Industry', value: 'compIndustry' },
        { text: 'Company Creation Date', value: 'compCreatedOn' },
        { text: 'Company Update Date', value: 'compUpdatedOn' },
        { text: 'Contact Id', value: 'contId' },
        { text: 'Contact Gender', value: 'contGender' },
        { text: 'Contact First Name', value: 'contFirstName' },
        { text: 'Contact Last Name', value: 'contLastName' },
        { text: 'Contact Job Title', value: 'contJobTitle' },
        { text: 'Contact Job Function', value: 'contJobFunction' },
        { text: 'Contact Job Level', value: 'contJobLevel' },
        { text: 'Contact Telephone', value: 'contTelephone' },
        { text: 'Contact Street Number', value: 'contStreetNumber' },
        { text: 'Contact Route', value: 'contRoute' },
        { text: 'Contact Postal Code', value: 'contPostalCode' },
        { text: 'Contact Locality', value: 'contLocality' },
        { text: 'Contact Admin Area Level 2', value: 'contAdministrativeAreaLevel2' },
        { text: 'Contact Admin Area Level 1', value: 'contAdministrativeAreaLevel1' },
        { text: 'Contact Country', value: 'contCountry' },
        { text: 'Contact Email', value: 'contEmail' },
        { text: 'Contact Email Status', value: 'contEmailStatus' },
        { text: 'Contact Email Creation Date', value: 'contEmailCreatedOn' },
        { text: 'Contact Social Profile URL', value: 'contSocProfURL' },
        { text: 'Contact Industry', value: 'contIndustry' },
        { text: 'Contact Creation Date', value: 'contCreatedOn' },
        { text: 'Contact Update Date', value: 'contUpdatedOn' }
      ],
      // Form validation rules.
      // Should add a global validation rule so that if all fields are empty, send button is
      // greyed, but seems pretty complicated with Vuetify.
      // All frontend validation rules are also enforced in backend.
      // Content of input CSV could be checked in frontend but I'm affraid this is too heavy
      // so leave it to backend only.
      rules: {
        // isText checks that content is text
        isText: (value) => {
          // https://stackoverflow.com/questions/9716468/is-there-any-function-like-isnumeric-in-javascript-to-validate-numbers
          // Problem is that is doesn't work if start with a letter and add numbers
          // after
          if ((isNaN(parseFloat(value)) && !isFinite(value)) || value === '') {
            return true
          } else {
            return 'Please enter text'
          }
        }
      }
    }
  },
  // created is the way to launch things once everything else is loaded in the page
  created () {
    // Load data from API into input selects.
    // Axios.all/spread allows 2 levels of error handling:
    // - catching error for each request in axios.all: HTTP.get('/get-countries-list').catch(...)
    // - catching error globally after axios.spread: .then(axios.spread(...)).catch(...)
    // The first option allows us to display precise error messages depending on which request
    // raised an error, but this is non blocking so we still enter in axio.spread(...) despite the error
    // and some of the parameters (countriesResp, ...) will be undefined in axios.spread() so need to handle it.
    // In the second option, a global error is raised as soon as one of the requests fails at least, and
    // we do not enter axios.spread().
    axios.all([
      HTTP.get('/get-countries-list'),
      HTTP.get('/get-companies-industries-list'),
      HTTP.get('/get-companies-sizes-list'),
      HTTP.get('/get-companies-types-list'),
      HTTP.get('/get-contacts-industries-list'),
      HTTP.get('/get-contacts-functions-list'),
      HTTP.get('/get-contacts-levels-list')
    ])
    // If all requests succeed
    .then(axios.spread(function (
      // Each response comes from the get query above
      countriesResp,
      companyIndustriesResp,
      companySizesResp,
      companyTypesResp,
      contactIndustriesResp,
      contactFunctionsResp,
      contactLevelsResp
    ) {
      // Put countries retrieved from API into an array available to Vue.js
      this.countriesAreLoading = false
      this.countries = []
      for (let i = countriesResp.data.length - 1; i >= 0; i--) {
        this.countries.push(countriesResp.data[i].countryName)
      }
      // Remove France and put it at the top for convenience
      let indexOfFrance = this.countries.indexOf('France')
      this.countries.splice(indexOfFrance, 1)
      // Sort the data alphabetically for convenience
      this.countries.sort()
      this.countries.unshift('France')

      // Put company industries retrieved from API into an array available to Vue.js
      this.companyIndustriesAreLoading = false
      this.companyIndustries = []
      for (let i = companyIndustriesResp.data.length - 1; i >= 0; i--) {
        this.companyIndustries.push(companyIndustriesResp.data[i].industryName)
      }
      this.companyIndustries.sort()

      // Put company sizes retrieved from API into an array available to Vue.js
      this.companySizesAreLoading = false
      this.companySizes = []
      for (let i = companySizesResp.data.length - 1; i >= 0; i--) {
        this.companySizes.push(companySizesResp.data[i].sizeName)
      }
      this.companySizes.sort()

      // Put company types retrieved from API into an array available to Vue.js
      this.companyTypesAreLoading = false
      this.companyTypes = []
      for (let i = companyTypesResp.data.length - 1; i >= 0; i--) {
        this.companyTypes.push(companyTypesResp.data[i].typeName)
      }
      this.companyTypes.sort()

      // Put contact industries retrieved from API into an array available to Vue.js
      this.contactIndustriesAreLoading = false
      this.contactIndustries = []
      for (let i = contactIndustriesResp.data.length - 1; i >= 0; i--) {
        this.contactIndustries.push(contactIndustriesResp.data[i].industryName)
      }
      this.contactIndustries.sort()

      // Put contact functions retrieved from API into an array available to Vue.js
      this.contactFunctionsAreLoading = false
      this.contactFunctions = []
      for (let i = contactFunctionsResp.data.length - 1; i >= 0; i--) {
        this.contactFunctions.push(contactFunctionsResp.data[i].functionName)
      }
      this.contactFunctions.sort()

      // Put contact levels retrieved from API into an array available to Vue.js
      this.contactLevelsAreLoading = false
      this.contactLevels = []
      for (let i = contactLevelsResp.data.length - 1; i >= 0; i--) {
        this.contactLevels.push(contactLevelsResp.data[i].levelName)
      }
      this.contactLevels.sort()
    }
    // bind(this) is need in order to inject this of Vue.js (otherwise
    // this would be the axios instance)
    .bind(this)))
    // In case one of the get request failed, stop everything and tell the user
    .catch(e => {
      alert('Could not load the full input lists in form.')
      this.countriesAreLoading = false
      this.companyIndustriesAreLoading = false
      this.companySizesAreLoading = false
      this.companyTypesAreLoading = false
      this.contactIndustriesAreLoading = false
      this.contactFunctionsAreLoading = false
      this.contactLevelsAreLoading = false
    })
  },
  methods: {
    sendData () {
      // Send data in the JSON format through POST
      HTTP.post('/get-companies-and-contacts', {
        step: this.step,
        companyCity: this.companyCity,
        companyPostCode: this.companyPostCode,
        companyCountries: this.companyCountries,
        companyIndustries: this.companyIndustriesSelected,
        companySizes: this.companySizesSelected,
        companyTypes: this.companyTypesSelected,
        companyHasPhone: this.companyHasPhone,
        companyHasEmail: this.companyHasEmail,
        companyDomains: this.companyDomains,
        excludedCompanyDomains: this.excludedCompanyDomains,
        contactCity: this.contactCity,
        contactPostCode: this.contactPostCode,
        contactCountries: this.contactCountries,
        contactIndustries: this.contactIndustriesSelected,
        contactJobTitle: this.contactJobTitle,
        contactFunctions: this.contactFunctionsSelected,
        contactLevels: this.contactLevelsSelected,
        contactHasEmail: this.contactHasEmail,
        contactRemoteAccounts: this.contactRemoteAccounts,
        excludedContactRemoteAccounts: this.excludedContactRemoteAccounts
      })
      // If request succeeds, store results into this.resultsRows and set
      // a couple of presentation parameters
      .then(response => {
        if (this.step === 'count') {
          // Get the number of rows retrieved by getting info from API
          this.resultsRowsNb = response.data
          this.showGenerateCSV = false
          this.showGetFullResultsBtn = true
          this.showResultsRowsNb = true
        }
        if (this.step === 'full') {
          if (response.status === 204) {
            this.showShowResultsBtn = false
            this.showGenerateCSV = false
            this.showGetFullResultsBtn = false
            this.warningMessage = 'The request returned too many lines so results have been sent to you by email.'
            this.showWarning = true
            this.showResultsRowsNb = false
          } else if (response.status === 200) {
            this.showShowResultsBtn = true
            // Get the number of rows retrieved by counting lines in array
            this.resultsRows = response.data
            this.resultsRowsNb = this.resultsRows.length
            this.showGenerateCSV = true
            this.showGetFullResultsBtn = false
            this.showResultsRowsNb = true
          }
        }
        this.showLoader = false
        this.showResults = false
        this.showNewSearchBtn = true
      })
      // If error
      .catch(e => {
        // If the request was sent but API returned an error
        if (e.response) {
          if (e.response.status === 404) {
            this.warningMessage = 'No result found for this search'
            this.showWarning = true
          } else if (e.response.status === 500) {
            this.errorMessage = 'Backend server error. Please contact an admin.'
            this.showError = true
          } else if (e.response.status === 400) {
            // Here we also show the error message raised by API because gives us more information
            // about which input field was not properly formatted
            this.errorMessage = 'Some of your data are not properly formated.\n ' + e.response.data
            this.showError = true
          } else {
            this.errorMessage = e.response.data
            this.showError = true
          }
        // If the request could not be sent because of a network error for example
        } else if (e.request) {
          this.errorMessage = 'No response from server'
          this.showError = true
        // For any other kind of error
        } else {
          this.errorMessage = e.message
          this.showError = true
        }
        this.showShowResultsBtn = false
        this.showResultsRowsNb = false
        this.showLoader = false
        this.showForm = true
        this.showNewSearchBtn = true
        this.showGenerateCSV = false
      })
    },
    // count Results sends form data to API and gets rows nb without results
    countResults () {
      // Send the form only if every input is valid (but send button is greyed
      // in that case, so should not happen)
      if (this.$refs.form.validate()) {
        // Before sending data, initialize a couple of things
        this.step = 'count'
        this.showError = false
        this.showWarning = false
        this.showForm = false
        this.showLoader = true
        this.sendData()
      }
    },
    // GetFullResults sends form data to API and gets results
    getFullResults () {
      if (this.$refs.form.validate()) {
        this.step = 'full'
        this.showError = false
        this.showWarning = false
        this.showForm = false
        this.showLoader = true
        this.sendData()
      }
    },
    // clearForm clears all fields of the form
    clearForm () {
      this.companyCity = ''
      this.companyPostCode = ''
      this.companyCountries = []
      this.companyIndustriesSelected = []
      this.companySizesSelected = []
      this.companyTypesSelected = []
      this.companyHasPhone = 0
      this.companyHasEmail = 0
      this.companyDomains = []
      this.excludedCompanyDomains = []
      this.contactCity = ''
      this.contactPostCode = ''
      this.contactCountries = []
      this.contactIndustriesSelected = []
      this.contactJobTitle = ''
      this.contactFunctionsSelected = []
      this.contactLevelsSelected = []
      this.contactHasEmail = 0
      this.contactRemoteAccounts = []
      this.excludedContactRemoteAccounts = []
      // CSV data were put in arrays and those arrays were emptied above.
      // But need to remove the name of the file close to the file input button
      // (cosmetic issue only).
      const csv1 = this.$refs.csv1
      csv1.type = 'text'
      csv1.type = 'file'
      const csv2 = this.$refs.csv2
      csv2.type = 'text'
      csv2.type = 'file'
      const csv3 = this.$refs.csv3
      csv3.type = 'text'
      csv3.type = 'file'
      const csv4 = this.$refs.csv4
      csv4.type = 'text'
      csv4.type = 'file'
    },
    // newSearch comes back to form for a new search
    newSearch () {
      this.showForm = true
      this.showError = false
      this.showWarning = false
      this.showResults = false
      this.showNewSearchBtn = false
      this.showResultsRowsNb = false
      this.showShowResultsBtn = false
      this.showLoader = false
    },
    // getCompanyDomainsCSV loads CSV
    // https://stackoverflow.com/questions/34498596/vuejs-csv-filereader
    getCompanyDomainsCSV: function (e) {
      let files = e.target.files || e.dataTransfer.files
      if (!files.length) {
        alert('No CSV received!')
        return
      }
      this.convCompanyDomainsCSVToArray(files[0])
    },
    // convCompanyDomainsCSVToArray puts CSV data into an array managed by Vue.js
    convCompanyDomainsCSVToArray: function (file) {
      let reader = new FileReader()
      let vm = this
      reader.onload = (e) => {
        vm.fileinput = reader.result
        let csvData = vm.fileinput.split('\n')
        csvData.splice(-1, 1)
        this.companyDomains = csvData
      }
      reader.readAsText(file)
    },
    getExcludedCompanyDomainsCSV: function (e) {
      let files = e.target.files || e.dataTransfer.files
      if (!files.length) {
        alert('No CSV received!')
        return
      }
      this.convExcludedCompanyDomainsCSVToArray(files[0])
    },
    convExcludedCompanyDomainsCSVToArray: function (file) {
      let reader = new FileReader()
      let vm = this
      reader.onload = (e) => {
        vm.fileinput = reader.result
        let csvData = vm.fileinput.split('\n')
        csvData.splice(-1, 1)
        this.excludedCompanyDomains = csvData
      }
      reader.readAsText(file)
    },
    getContactRemoteAccountsCSV: function (e) {
      let files = e.target.files || e.dataTransfer.files
      if (!files.length) {
        alert('No CSV received!')
        return
      }
      this.convContactRemoteAccountsCSVToArray(files[0])
    },
    convContactRemoteAccountsCSVToArray: function (file) {
      let reader = new FileReader()
      let vm = this
      reader.onload = (e) => {
        vm.fileinput = reader.result
        let csvData = vm.fileinput.split('\n')
        csvData.splice(-1, 1)
        this.contactRemoteAccounts = csvData
      }
      reader.readAsText(file)
    },
    getExcludedContactRemoteAccountsCSV: function (e) {
      let files = e.target.files || e.dataTransfer.files
      if (!files.length) {
        alert('No CSV received!')
        return
      }
      this.convExcludedContactRemoteAccountsCSVToArray(files[0])
    },
    convExcludedContactRemoteAccountsCSVToArray: function (file) {
      let reader = new FileReader()
      let vm = this
      reader.onload = (e) => {
        vm.fileinput = reader.result
        let csvData = vm.fileinput.split('\n')
        csvData.splice(-1, 1)
        this.excludedContactRemoteAccounts = csvData
      }
      reader.readAsText(file)
    },
    // generateCSV puts this.resultsRows into a CSV and serves it to the user.
    // Delimiter is ";".
    generateCSV: function () {
      let csvArray = [
        'data:text/csv;charset=utf-8,' +
        'Company Id;' +
        'Company Name;' +
        'Company Domain;' +
        'Company Website;' +
        'Company Telephone;' +
        'Company Fax Number;' +
        'Company Size;' +
        'Company Founded;' +
        'Company Street Number;' +
        'Company Route;' +
        'Company Postal Code;' +
        'Company Locality;' +
        'Company Admin Area Level 1;' +
        'Company Admin Area Level 2;' +
        'Company Country;' +
        'Company Email;' +
        'Company Social Profile URL;' +
        'Company Type;' +
        'Company Industry;' +
        'Company Creation Date;' +
        'Company Update Date;' +
        'Contact Id;' +
        'Contact Gender;' +
        'Contact First Name;' +
        'Contact Last Name;' +
        'Contact Job Title;' +
        'Contact Job Function;' +
        'Contact Job Level;' +
        'Contact Telephone;' +
        'Contact Street Number;' +
        'Contact Route;' +
        'Contact Postal Code;' +
        'Contact Locality;' +
        'Contact Admin Area Level 1;' +
        'Contact Admin Area Level 2;' +
        'Contact Country;' +
        'Contact Email;' +
        'Contact Email Status;' +
        'Contact Email Creation Date;' +
        'Contact Social Profile URL;' +
        'Contact Industry;' +
        'Contact Creation Date;' +
        'Contact Update Date'
      ]
      this.resultsRows.forEach(function (row) {
        let csvRow = row['compId'] + ';' +
          row['compName'] + ';' +
          row['compDomain'] + ';' +
          row['compWebsite'] + ';' +
          row['compTelephone'] + ';' +
          row['compFaxNumber'] + ';' +
          row['compSize'] + ';' +
          row['compFounded'] + ';' +
          row['compStreetNumber'] + ';' +
          row['compRoute'] + ';' +
          row['compPostalCode'] + ';' +
          row['compLocality'] + ';' +
          row['compAdministrativeAreaLevel1'] + ';' +
          row['compAdministrativeAreaLevel2'] + ';' +
          row['compCountry'] + ';' +
          row['compEmail'] + ';' +
          row['compSocProfURL'] + ';' +
          row['compType'] + ';' +
          row['compIndustry'] + ';' +
          row['compCreatedOn'] + ';' +
          row['compUpdatedOn'] + ';' +
          row['contId'] + ';' +
          row['contGender'] + ';' +
          row['contFirstName'] + ';' +
          row['contLastName'] + ';' +
          row['contJobTitle'] + ';' +
          row['contJobFunction'] + ';' +
          row['contJobLevel'] + ';' +
          row['contTelephone'] + ';' +
          row['contStreetNumber'] + ';' +
          row['contRoute'] + ';' +
          row['contPostalCode'] + ';' +
          row['contLocality'] + ';' +
          row['contAdministrativeAreaLevel1'] + ';' +
          row['contAdministrativeAreaLevel2'] + ';' +
          row['contCountry'] + ';' +
          row['contEmail'] + ';' +
          row['contEmailStatus'] + ';' +
          row['contEmailCreatedOn'] + ';' +
          row['contSocProfURL'] + ';' +
          row['contIndustry'] + ';' +
          row['contCreatedOn'] + ';' +
          row['contUpdatedOn']
        csvArray.push(csvRow)
      })
      let csvContent = csvArray.join('\r\n')
      let encodedUri = encodeURI(csvContent)
      let link = document.createElement('a')
      link.setAttribute('href', encodedUri)
      link.setAttribute('download', 'companies_and_contacts_extracted.csv')
      document.body.appendChild(link)
      link.click()
    }
  }
}
/* beautify preserve:end */
</script>
