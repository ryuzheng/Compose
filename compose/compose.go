// Copyright (C) 2015  Matt Borgerson
// 
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// 
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// 
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
    "fmt"
    "html/template"
    "net/http"
    "os"
    "path/filepath"
    "regexp"
    // "strconv"
    // "strings"


    "github.com/zenazn/goji"
)

var templates *template.Template
var AdminTemplates *template.Template

// BuildTemplates builds all the required site templates.
func BuildTemplates() error {
    funcMap := template.FuncMap {
        "add": func(a, b int) int { return a+b },
        "sub": func(a, b int) int { return a-b },
    }

    files := []string{
        "index.html",
        "post.html",
    }

    for i, file := range files {
        files[i] = filepath.Join(config.TemplatesPath, file)
    }

    _templates := template.New("base")
    _templates.Delims("<%", "%>")
    _templates.Funcs(funcMap)
    _templates, err := _templates.ParseFiles(files...)
    if err != nil {
        return err
    }
    templates = _templates

    files = []string{
        "index.html", 
        "edit.html",
        "posts.html",
        "settings.html",
        "login.html",
    }

    for i, file := range files {
        files[i] = filepath.Join(config.AdminTemplatesPath, file)
    }

    _templates = template.New("base")
    _templates.Delims("<%", "%>")
    _templates.Funcs(funcMap)
    _templates, err = _templates.ParseFiles(files...)
    if err != nil {
        return err
    }
    AdminTemplates = _templates
    return nil
}

// main is the entry point. Loads the program resources and begins waiting for
// connections.
func main() {
    // Create a config file with the defaults
    if !FileExists(ConfigDefaultFilename) {
        config, _ := GetDefaultConfig()
        err := config.Save(ConfigDefaultFilename)
        if err != nil {
            fmt.Println("I tried to create default config file but failed. Check directory permissions.")
            os.Exit(1)
        }
        fmt.Println("The config file could not be found, so I created a config file at '", ConfigDefaultFilename, "'. Please ensure this file contains the correct values and relaunch.")
        os.Exit(0)
    }

    // Load config
    _, err := GetConfig()
    if err != nil {
        fmt.Println("Failed to load the config file:", err.Error())
        os.Exit(1)
    }

    // Build Templates
    err = BuildTemplates()
    if err != nil {
        fmt.Println("Failed to build templates:", err.Error())
        os.Exit(1)
    }

    // Connect to the database
    err = SetupDatabaseSession()
    if err != nil {
        panic(err)
    }
    defer CleanupDatabaseSession()

    // Setup the router
    goji.Handle("/api/*",                               GetApiHandler())

    goji.Get("/setup",                                  SetupHandler)
    goji.Get("/admin/partials/edit",                    MakeRestrictedHttpHandler(AdminEditHandler))
    goji.Get("/admin/partials/posts",                   MakeRestrictedHttpHandler(AdminPostsHandler))
    goji.Get("/admin/partials/settings",                MakeRestrictedHttpHandler(AdminSettingsHandler))
    goji.Get("/admin/assets/*",                         AdminAssetsHandler)
    goji.Get("/admin/*",                                MakeRestrictedHttpHandler(AdminHandler))
    
    goji.Get("/assets/*",                               http.StripPrefix("/assets/", http.FileServer(http.Dir(config.AssetsPath))).ServeHTTP)
    
    goji.Get("/login",                                  LoginHandler)
    goji.Post("/login",                                 LoginHandler)
    goji.Get("/logout",                                 LogoutHandler)
    
    goji.Post("/upload",                                MakeRestrictedHttpHandler(UploadHandler))
    goji.Get(regexp.MustCompile("^/(?P<page>[0-9]*)$"), IndexHandler)
 
    goji.Get("/:slug",                                  ViewHandler)
    goji.Get("/:slug/",                                 ViewHandlerRemoveTrailingSlash) // Courtesy redirect for SEO
    goji.Get("/:slug/:file",                            ViewFileHandler)

    // Begin serving
    goji.Serve()
}