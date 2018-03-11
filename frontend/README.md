# Purpose

Frontend of the app which is a VueJS/Axios/Vuetify app sending info to a Go backend.

# Security

Password protected thanks to a .htpasswd file.

# Sources

[VueJS template](https://vuejs-templates.github.io/webpack/structure.html)

# Project setup

1. `npm install -g vue-cli`
1. `vue init webpack vue_project`
1. `cd vue_project`
1. `npm install`
1. `npm install --save axios`
1. `npm run dev`
1. `npm run build`

# Local development

`npm run dev`

Hot reloading did not work out of the box because of a permission issue.
So I had to use the solution mentioned [in this thread](https://github.com/vuejs-templates/webpack/issues/1054) :

`echo 100000 | sudo tee /proc/sys/fs/inotify/max_user_watches` 

(initial value was 8192).

# Deployment

## Server

`docker run --net my_network --ip 172.50.0.11 -p 9000:9000 -d --name frontend_v1_container myaccount/myrepo:frontend_v1`