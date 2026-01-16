package main

import (
	"bytes"
	"embed"
	"encoding/csv"
	"html/template"
	"math"
	"net/http"
	"sort"
	"strconv"
)

//go:embed countries.csv
var csvData embed.FS

type Row struct {
	Country      string
	Both         float64
	NationalDish string
	DishWiki     string
	ImageLink    string
}

type CardData struct {
	Row
	GridRow int
	GridCol int
	Order   int
}

type PageData struct {
	Cards   []CardData
	NumCols int
	NumRows int
}

var tmpl map[string]*template.Template
var rows []Row

func calculateDiagonalLayout(rows []Row, numCols int) PageData {
	numItems := 0
	for _, r := range rows {
		if r.ImageLink != "" {
			numItems++
		}
	}

	numRows := (numItems + numCols - 1) / numCols

	cards := make([]CardData, 0, numItems)
	positions := make([]struct{ row, col int }, 0, numItems)

	// Generate diagonal positions
	for d := 0; d < numRows+numCols-1 && len(positions) < numItems; d++ {
		var row, col int
		if d < numRows {
			row = d
			col = 0
		} else {
			row = numRows - 1
			col = d - numRows + 1
		}

		for row >= 0 && col < numCols && len(positions) < numItems {
			positions = append(positions, struct{ row, col int }{row + 1, col + 1})
			row--
			col++
		}
	}

	// Assign positions to rows with images
	posIndex := 0
	orderIndex := 1
	for _, r := range rows {
		if r.ImageLink != "" && posIndex < len(positions) {
			cards = append(cards, CardData{
				Row:     r,
				GridRow: positions[posIndex].row,
				GridCol: positions[posIndex].col,
				Order:   orderIndex,
			})
			posIndex++
			orderIndex++
		}
	}

	return PageData{
		Cards:   cards,
		NumCols: numCols,
		NumRows: numRows,
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	pageData := calculateDiagonalLayout(rows, 4)
	err := tmpl["home"].ExecuteTemplate(w, "base", pageData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loadCSV() error {
	data, err := csvData.ReadFile("countries.csv")
	if err != nil {
		return err
	}

	reader := csv.NewReader(bytes.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for i, rec := range records {
		if i == 0 {
			continue
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

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Both > rows[j].Both
	})

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

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "."+r.URL.Path)
	})

	http.HandleFunc("/", homeHandler)
	http.ListenAndServe(":8083", nil)
}
