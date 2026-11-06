package main

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Stabseite repräsentiert ein Feldende an einem Knoten
type Stabseite struct {
	Name       string
	Mu         float64
	M0         float64
	MVert      float64
	MOver      float64
	MTotal     float64
	OppositeId int
}

// Knoten gruppiert die Stabseiten
type Knoten struct {
	Name    string
	SeitenI []int
}

// BerechnungsErgebnis hält die Daten für die HTML-Ausgabe
type BerechnungsErgebnis struct {
	Fehler   string
	Laengen  string
	Resultat []Stabseite
}

// HTML-Template für die Benutzeroberfläche im Browser
const htmlTemplate = `
<!DOCTYPE html>
<html lang="de">
<head>
    <meta charset="UTF-8">
    <title>Cross-Verfahren Durchlaufträger</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 40px auto; padding: 20px; line-height: 1.6; }
        h2 { color: #333; }
        .form-group { margin-bottom: 15px; }
        label { display: block; font-weight: bold; margin-bottom: 5px; }
        input[type="text"] { width: 100%; padding: 8px; box-sizing: border-box; }
        button { background: #007BFF; color: white; border: none; padding: 10px 15px; cursor: pointer; font-size: 16px; width: 100%; }
        button:hover { background: #0056b3; }
        .error { color: red; font-weight: bold; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { border: 1px solid #ccc; padding: 10px; text-align: left; }
        th { background: #f4f4f4; }
    </style>
</head>
<body>
    <h2>Berechnung eines Durchlaufträgers (Cross-Verfahren)</h2>
    <p>Konstante Last <strong>q = 10.0 kN/m</strong></p>
    
    {{if .Fehler}}<p class="error">{{.Fehler}}</p>{{end}}

    <form method="POST">
        <div class="form-group">
            <label for="laengen">Feldlängen in Metern (mit Komma oder Leerzeichen getrennt):</label>
            <input type="text" id="laengen" name="laengen" value="{{.Laengen}}" placeholder="z.B. 4.5, 6.0, 5.2">
        </div>
        <button type="submit">Berechnen</button>
    </form>

    {{if .Resultat}}
    <h3>Ergebnis:</h3>
    <table>
        <tr>
            <th>Stabseite</th>
            <th>µ (Verteilwert)</th>
            <th>Stützmoment (kNm)</th>
        </tr>
        {{range .Resultat}}
        <tr>
            <td>{{.Name}}</td>
            <td>{{printf "%.2f" .Mu}}</td>
            <td>{{printf "%.2f" .MTotal}}</td>
        </tr>
        {{end}}
    </table>
    {{end}}
</body>
</html>
`

func main() {
	tmpl := template.Must(template.New("cross").Parse(htmlTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			tmpl.Execute(w, BerechnungsErgebnis{})
			return
		}

		rawLaengen := r.FormValue("laengen")
		// Ersetze Kommas durch Leerzeichen und teile den String auf
		cleanString := strings.ReplaceAll(rawLaengen, ",", " ")
		fields := strings.Fields(cleanString)

		if len(fields) == 0 {
			tmpl.Execute(w, BerechnungsErgebnis{Fehler: "Bitte gib mindestens eine Feldlänge ein.", Laengen: rawLaengen})
			return
		}

		var laengen []float64
		for _, f := range fields {
			val, err := strconv.ParseFloat(f, 64)
			if err != nil || val <= 0 {
				tmpl.Execute(w, BerechnungsErgebnis{Fehler: "Ungültige Feldlänge eingegeben. Nur positive Zahlen erlaubt.", Laengen: rawLaengen})
				return
			}
			laengen = append(laengen, val)
		}

		anzahlFelder := len(laengen)
		const q = 10.0
		anzahlSeiten := anzahlFelder * 2
		seiten := make([]Stabseite, anzahlSeiten)
		kSteifigkeit := make([]float64, anzahlSeiten)

		// HIER FOLGT DEINE ORIGINALE MATHEMATISCHE LOGIK
		for i := 0; i < anzahlFelder; i++ {
			l := laengen[i]
			leftIdx := i * 2
			rightIdx := i*2 + 1

			charLinks := string(rune('A' + i))
			charRechts := string(rune('A' + i + 1))
			seiten[leftIdx].Name = fmt.Sprintf("%s-rechts", charLinks)
			seiten[rightIdx].Name = fmt.Sprintf("%s-links", charRechts)

			seiten[leftIdx].OppositeId = rightIdx
			seiten[rightIdx].OppositeId = leftIdx

			if i == 0 {
				kSteifigkeit[leftIdx] = 0.75 * (1.0 / l)
				kSteifigkeit[rightIdx] = 0.75 * (1.0 / l)
				seiten[leftIdx].M0 = 0.0
				seiten[rightIdx].M0 = (q * l * l) / 8.0
			} else if i == anzahlFelder-1 {
				kSteifigkeit[leftIdx] = 0.75 * (1.0 / l)
				kSteifigkeit[rightIdx] = 0.75 * (1.0 / l)
				seiten[leftIdx].M0 = -(q * l * l) / 8.0
				seiten[rightIdx].M0 = 0.0
			} else {
				kSteifigkeit[leftIdx] = 1.0 / l
				kSteifigkeit[rightIdx] = 1.0 / l
				seiten[leftIdx].M0 = -(q * l * l) / 12.0
				seiten[rightIdx].M0 = (q * l * l) / 12.0
			}
		}

		if anzahlFelder == 1 {
			seiten[0].M0 = 0
			seiten[1].M0 = 0
		}

		anzahlKnoten := anzahlFelder + 1
		knoten := make([]Knoten, anzahlKnoten)

		for i := 0; i < anzahlKnoten; i++ {
			knoten[i].Name = string(rune('A' + i))
			if i == 0 {
				knoten[i].SeitenI = []int{0}
				seiten[0].Mu = 1.0
			} else if i == anzahlKnoten-1 {
				knoten[i].SeitenI = []int{anzahlSeiten - 1}
				seiten[anzahlSeiten-1].Mu = 1.0
			} else {
				idxLinks := (i * 2) - 1
				idxRechts := i * 2
				knoten[i].SeitenI = []int{idxLinks, idxRechts}
				sumK := kSteifigkeit[idxLinks] + kSteifigkeit[idxRechts]
				seiten[idxLinks].Mu = kSteifigkeit[idxLinks] / sumK
				seiten[idxRechts].Mu = kSteifigkeit[idxRechts] / sumK
			}
		}

		for i := range seiten {
			seiten[i].MTotal = seiten[i].M0
		}

		maxIterationen := 20
		epsilon := 0.001

		for iter := 1; iter <= maxIterationen; iter++ {
			groessteAenderung := 0.0
			for _, k := range knoten {
				knotenSumme := 0.0
				for _, idx := range k.SeitenI {
					if iter == 1 {
						knotenSumme += seiten[idx].M0
					} else {
						knotenSumme += seiten[idx].MOver
					}
				}
				for _, idx := range k.SeitenI {
					seiten[idx].MVert = -knotenSumme * seiten[idx].Mu
					seiten[idx].MTotal += seiten[idx].MVert
					if math.Abs(seiten[idx].MVert) > groessteAenderung {
						groessteAenderung = math.Abs(seiten[idx].MVert)
					}
				}
			}
			if iter > 1 && groessteAenderung < epsilon {
				break
			}
			for i := range seiten {
				seiten[i].MOver = 0
			}
			for i := range seiten {
				oppIdx := seiten[i].OppositeId
				if seiten[oppIdx].Mu != 1.0 {
					seiten[oppIdx].MOver = seiten[i].MVert * 0.5
				}
			}
		}

		// Runden für die Anzeige im Web
		for i := range seiten {
			roundedM := math.Round(seiten[i].MTotal*100) / 100
			if math.Abs(roundedM) == 0.0 {
				roundedM = 0.0
			}
			seiten[i].MTotal = roundedM
		}

		tmpl.Execute(w, BerechnungsErgebnis{Laengen: rawLaengen, Resultat: seiten})
	})

	// Cloud-Port auslesen
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server läuft online auf Port " + port)
	http.ListenAndServe(":"+port, nil)
}
