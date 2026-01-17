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
	AspectRatio  float64
}

// type CardData struct {
// 	Row
// 	GridRow int
// 	GridCol int
// 	Order   int
// }

// type PageData struct {
// 	Cards   []CardData
// 	NumCols int
// 	NumRows int
// }

type CardData struct {
	Row
	Order int
}

type PageData struct {
	Cards   []CardData
	NumCols int
	NumRows int
}

var tmpl map[string]*template.Template
var rows []Row

func prepareCards(rows []Row) []CardData {
	cards := make([]CardData, 0)
	order := 1
	for _, r := range rows {
		if r.ImageLink != "" {
			cards = append(cards, CardData{
				Row:   r,
				Order: order,
			})
			order++
		}
	}
	return cards
}

func reorderForRowLayout(cards []CardData, numCols int) PageData {
	n := len(cards)
	if n == 0 {
		return PageData{
			Cards:   cards,
			NumCols: numCols,
			NumRows: 0,
		}
	}

	numRows := (n + numCols - 1) / numCols
	reordered := make([]CardData, 0, n)

	for col := 0; col < numCols; col++ {
		for row := 0; row < numRows; row++ {
			srcIndex := row*numCols + col
			if srcIndex < n {
				reordered = append(reordered, cards[srcIndex])
			}
		}
	}

	return PageData{
		Cards:   reordered,
		NumCols: numCols,
		NumRows: numRows,
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// pageData := calculateDiagonalLayout(rows, 4)
	// pageData.Cards = reorderForRowLayout(pageData.Cards, pageData.NumCols)
	cards := prepareCards(rows)
	pageData := reorderForRowLayout(cards, 4)
	// pageData := PageData{
	// 	Cards: cards,
	// }
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

		aspectRatio, err := strconv.ParseFloat(rec[7], 64)
		if err != nil {
			aspectRatio = 0
		}

		rows = append(rows, Row{
			Country:      rec[0],
			Both:         both,
			NationalDish: rec[4],
			DishWiki:     rec[5],
			ImageLink:    rec[6],
			AspectRatio:  aspectRatio,
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
