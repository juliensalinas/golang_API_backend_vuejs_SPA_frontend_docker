<template>
  <div id="app">
    <v-app id="inspire">
      <v-navigation-drawer fixed v-model="drawer" app>
        <v-list three-line>
          <v-list-tile @click="goToGetCompaniesAndContacts">
            <v-list-tile-content>
              <v-list-tile-title>Companies and Contacts</v-list-tile-title>
              <v-list-tile-sub-title>Get companies and associated contacts from the Remote database</v-list-tile-sub-title>
            </v-list-tile-content>
          </v-list-tile>
          <v-list-tile @click="goToGetEmailsCheckedByJohn">
            <v-list-tile-content>
              <v-list-tile-title>Emails Checked By John</v-list-tile-title>
              <v-list-tile-sub-title>
                Get emails located in the email_checked_by_john table
              </v-list-tile-sub-title>
            </v-list-tile-content>
          </v-list-tile>
        </v-list>
      </v-navigation-drawer>
      <v-toolbar color="blue darken-3" dark fixed app>
        <v-toolbar-side-icon @click.stop="drawer = !drawer"></v-toolbar-side-icon>
        <v-toolbar-title>Menu</v-toolbar-title>
      </v-toolbar>
      <v-content>
        <v-container grid-list-md text-xs-center>
          <v-layout row wrap>
            <v-flex xs12 class="mt-5">
              <img src="./assets/logo.png">
            </v-flex>
            <v-flex xs12 class="mt-5">
              <router-view/>
            </v-flex>
          </v-layout>
        </v-container>
      </v-content>
      <v-footer color="blue darken-3" app>
        <span class="white--text">My project</span>
      </v-footer>
    </v-app>
  </div>
</template>
<script>
/* beautify preserve:start */
export default {
  name: 'app',
  data: () => ({
    drawer: null
  }),
  methods: {
    goToGetEmailsCheckedByJohn: function () {
      this.$router.push({ name: 'GetEmailsCheckedByJohn' })
    },
    goToGetCompaniesAndContacts: function () {
      this.$router.push({ name: 'GetCompaniesAndContacts' })
    }
  },
  props: {
    source: String
  }
}
/* beautify preserve:end */
</script>
