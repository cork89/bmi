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
	Order2Cols   int
	Order3Cols   int
	Order4Cols   int
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

func prepareCards(rows []Row, numCols int) []CardData {
	cards := make([]CardData, 0)
	for _, r := range rows {
		if r.ImageLink != "" {
			var order int
			switch numCols {
			case 2:
				order = r.Order2Cols
			case 3:
				order = r.Order3Cols
			case 4:
				order = r.Order4Cols
			default:
				order = r.Order4Cols
			}

			cards = append(cards, CardData{
				Row:   r,
				Order: order,
			})
		}
	}

	// Sort cards by their pre-calculated order
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Order < cards[j].Order
	})

	return cards
}

func calculateRows(numCards, numCols int) int {
	if numCards == 0 {
		return 0
	}
	return (numCards + numCols - 1) / numCols
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	numCols := 4
	cards := prepareCards(rows, numCols)
	pageData := PageData{
		Cards:   cards,
		NumCols: numCols,
		NumRows: calculateRows(len(cards), numCols),
	}
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

		order2Cols, _ := strconv.Atoi(rec[8])
		order3Cols, _ := strconv.Atoi(rec[9])
		order4Cols, _ := strconv.Atoi(rec[10])

		rows = append(rows, Row{
			Country:      rec[0],
			Both:         both,
			NationalDish: rec[4],
			DishWiki:     rec[5],
			ImageLink:    rec[6],
			AspectRatio:  aspectRatio,
			Order2Cols:   order2Cols,
			Order3Cols:   order3Cols,
			Order4Cols:   order4Cols,
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

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "."+r.URL.Path)
	})

	http.HandleFunc("/", homeHandler)
	http.ListenAndServe(":8083", nil)
}
