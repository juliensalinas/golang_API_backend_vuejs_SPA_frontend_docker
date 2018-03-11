import Vue from 'vue'
import Router from 'vue-router'

import GetEmailsCheckedByJohn from '@/components/GetEmailsCheckedByJohn'
import GetCompaniesAndContacts from '@/components/GetCompaniesAndContacts'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/get-emails-checked-by-john',
      name: 'GetEmailsCheckedByJohn',
      component: GetEmailsCheckedByJohn
    },
    {
      path: '/get-companies',
      name: 'GetCompaniesAndContacts',
      component: GetCompaniesAndContacts
    }
  ]
})
