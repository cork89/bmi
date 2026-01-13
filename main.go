package main

import (
	"bytes"
	"embed"
	"encoding/csv"
	"html/template"
	"math"
	"net/http"
	"strconv"
)

//go:embed countries.csv
var csvData embed.FS

type SlideData struct {
	Image       string
	Description string
}

type Row struct {
	Country      string
	Both         float64
	NationalDish string
	DishWiki     string
	ImageLink    string
}

var tmpl map[string]*template.Template

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl["home"].ExecuteTemplate(w, "base", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func slideHandler(w http.ResponseWriter, r *http.Request) {
	v, err := strconv.ParseFloat(r.URL.Query().Get("n"), 64)
	if err != nil {
		http.Error(w, "bad value", http.StatusBadRequest)
		return
	}

	row := closestRow(v)
	if row == nil {
		http.Error(w, "no data", http.StatusNotFound)
		return
	}

	err = tmpl["slide"].Execute(w, row)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var rows []Row

func loadCSV() error {
	data, err := csvData.ReadFile("countries.csv")
	if err != nil {
		return err
	}

	r := csv.NewReader(bytes.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	for i, rec := range records {
		if i == 0 {
			continue // header
		}

		both, err := strconv.ParseFloat(rec[1], 64)
		if err != nil {
			continue
		}

		rows = append(rows, Row{
			Country:      rec[0],
			Both:         both,
			NationalDish: rec[4],
			DishWiki:     rec[5],
			ImageLink:    rec[6],
		})
	}

	return nil
}

func closestRow(value float64) *Row {
	var best *Row
	bestDiff := math.MaxFloat64

	for i := range rows {
		diff := math.Abs(rows[i].Both - value)
		if diff < bestDiff {
			bestDiff = diff
			best = &rows[i]
		}
	}

	return best
}

func main() {
	err := loadCSV()
	if err != nil {
		panic(err)
	}

	tmpl = make(map[string]*template.Template)

	tmpl["home"] = template.Must(
		template.ParseFiles("static/home.html", "static/base.html"),
	)
	tmpl["slide"] = template.Must(
		template.ParseFiles("static/slide.html"),
	)

	http.HandleFunc("/slide", slideHandler)

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "."+r.URL.Path)
	})

	http.HandleFunc("/", homeHandler)
	http.ListenAndServe(":8083", nil)
}
