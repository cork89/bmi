package main

import (
	"bytes"
	"embed"
	"encoding/csv"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	colCountry     = 0
	colBMI         = 1
	colDish        = 4
	colDishWiki    = 5
	colImageLink   = 6
	colAspectRatio = 7
	colOrder2      = 8
	colOrder3      = 9
	colOrder4      = 10
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

type CardData struct {
	Row
	Order       int
	ImageSource string
}

type PageData struct {
	Cards   []CardData
	NumCols int
	NumRows int
}

type App struct {
	Tmpl         map[string]*template.Template
	Rows         []Row
	ContentCache map[string][]byte
}

func (app *App) preRenderContent() {
	for _, numCols := range []int{1, 2, 3, 4} {
		for _, reversed := range []bool{false, true} {
			key := fmt.Sprintf("cols:%d:rev:%t", numCols, reversed)

			cards := prepareCards(app.Rows, numCols, reversed)
			pageData := PageData{
				Cards:   cards,
				NumCols: numCols,
				NumRows: calculateRows(len(cards), numCols),
			}

			var buf bytes.Buffer
			if err := app.Tmpl["content"].ExecuteTemplate(&buf, "content", pageData); err != nil {
				log.Printf("failed to pre-render %s: %v", key, err)
				continue
			}

			app.ContentCache[key] = buf.Bytes()
		}
	}
}

func prepareCards(rows []Row, numCols int, reversed bool) []CardData {
	cards := make([]CardData, 0, len(rows))
	imgSource := os.Getenv("IMG_SOURCE")

	for i, r := range rows {
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
				order = i
			}

			cards = append(cards, CardData{
				Row:         r,
				Order:       order,
				ImageSource: imgSource,
			})
		}
	}

	sort.Slice(cards, func(i, j int) bool {
		if reversed {
			return cards[i].Order > cards[j].Order
		}
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

func (app *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	err := app.Tmpl["home"].ExecuteTemplate(w, "base", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *App) sourcesHandler(w http.ResponseWriter, r *http.Request) {
	err := app.Tmpl["sources"].ExecuteTemplate(w, "base", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getNumCols(r *http.Request) int {
	screenSizeHeader := r.Header.Get("X-Screen-Width")
	screenSize, err := strconv.Atoi(screenSizeHeader)

	if err != nil {
		screenSize = 1
	}

	numCols := 1

	if screenSize > 1400 {
		numCols = 4
	} else if screenSize > 1000 {
		numCols = 3
	} else if screenSize > 600 {
		numCols = 2
	}
	return numCols
}

func (app *App) contentHandler(w http.ResponseWriter, r *http.Request) {
	reversed := r.URL.Query().Get("sort") == "rev"
	numCols := getNumCols(r)

	key := fmt.Sprintf("cols:%d:rev:%t", numCols, reversed)

	if content, ok := app.ContentCache[key]; ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
		return
	}

	http.Error(w, "content not found", http.StatusInternalServerError)
}

func (app *App) loadCSV() error {
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
			// skip the column headers
			continue
		}

		both, err := strconv.ParseFloat(rec[colBMI], 64)
		if err != nil {
			log.Println("failed to parse BMI")
			continue
		}

		aspectRatio, err := strconv.ParseFloat(rec[colAspectRatio], 64)
		if err != nil {
			aspectRatio = 0
		}

		order2Cols, err := strconv.Atoi(rec[colOrder2])
		if err != nil {
			// log.Println("failed to parse second column order. defaulting to 0.")
			order2Cols = 0
		}
		order3Cols, err := strconv.Atoi(rec[colOrder3])
		if err != nil {
			// log.Println("failed to parse third column order. defaulting to 0.")
			order3Cols = 0
		}
		order4Cols, err := strconv.Atoi(rec[colOrder4])
		if err != nil {
			// log.Println("failed to parse fourth column order. defaulting to 0.")
			order4Cols = 0
		}

		app.Rows = append(app.Rows, Row{
			Country:      rec[colCountry],
			Both:         both,
			NationalDish: rec[colDish],
			DishWiki:     rec[colDishWiki],
			ImageLink:    rec[colImageLink],
			AspectRatio:  aspectRatio,
			Order2Cols:   order2Cols,
			Order3Cols:   order3Cols,
			Order4Cols:   order4Cols,
		})
	}

	return nil
}

func main() {
	app := App{
		Tmpl:         make(map[string]*template.Template),
		Rows:         make([]Row, 0),
		ContentCache: make(map[string][]byte),
	}

	err := app.loadCSV()
	if err != nil {
		panic(err)
	}

	err = godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	app.Tmpl["home"] = template.Must(
		template.ParseFiles("static/home.html", "static/base.html"),
	)
	app.Tmpl["content"] = template.Must(template.ParseFiles("static/content.html"))
	app.Tmpl["sources"] = template.Must(template.ParseFiles("static/sources.html", "static/base.html"))
	app.preRenderContent()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", app.homeHandler)
	http.HandleFunc("/content", app.contentHandler)
	http.HandleFunc("/sources", app.sourcesHandler)

	if err := http.ListenAndServe(":8083", nil); err != nil {
		log.Fatal(err)
	}
}
