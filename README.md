# DCMan
A datacenter management tool

DCMan is composed of 3 primary layers:
1. A SQL database ([SQLite](http://www.sqlite.org) embedded database)
2. A web server written in [Go](https://www.golang.org) that exposes data via REST and provides authentication, etc.
3. The presentation layer which is a single paged app using [vuejs](https://vuejs.org)

## The Database
The database is defined by 3 DDL files (in the sql directory):
1. schema.sql -- all table definitions
2. views.sql -- sql views that compose complex data renderings from tables
3. triggers.sql -- actions that happen upon updates to the database

By breaking the structure out in this manner, updates can be applied to database behavior in a non-destructive manner when updating a view or trigger. Because triggers are defined primarily against views, if one reloads the view file then the trigger file must be reapplied.

Because data is most often consumed in view form and views are not directly modifiable, triggers translate changes to views to their underlying tables.

Applying the files to the database is simply a matter of running the command:
    sqlite3 data.db < sql/views.sql

## The Web Server
The server uses the internal Go net/http library and other than serving static files (supporting JS/CSS), wraps most data access in a REST style interface. All web paths are in the handlers.go file.

It uses a config file to control facets of operation (port to bind to, Okta secrets, etc)

## The Interface
The web interface is a "single paged application" (SPA) that comprises two custom files: spa.html and spa.js
The UI uses Bootstrap CSS for layout presentation.

spa.html holds all the templates that vue.js uses to render each page.
spa.js contains the application logic to populate and control page behavior.
