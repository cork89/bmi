package main

import (
	"html/template"
	"net/http"
)

var tmpl map[string]*template.Template

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl["home"].ExecuteTemplate(w, "base", nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	tmpl = make(map[string]*template.Template)

	tmpl["home"] = template.Must(template.ParseFiles("static/home.html", "static/base.html"))

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "static/style.css")
	})
	http.HandleFunc("/", homeHandler)
	http.ListenAndServe(":8083", nil)
}
