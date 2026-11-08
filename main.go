package main

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os"
	"strconv"
)

type Stabseite struct {
	Name       string
	Mu         float64
	M0         float64
	MVert      float64
	MOver      float64
	MTotal     float64
	OppositeId int
}

type Knoten struct {
	Name    string
	SeitenI []int
}

type PageData struct {
	AnzahlFelder int
	LaengenInput string
	Berechnet    bool
	Fehler       string
	Seiten       []Stabseite
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="de">
<head>
    <meta charset="UTF-8">
    <title>Cross-Verfahren Berechnung</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 20px auto; padding: 20px; background: #fafafa; color: #333; }
        .card { background: white; padding: 25px; border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
        h1 { color: #111; border-bottom: 2px solid #eee; padding-bottom: 8px; margin-top: 0; font-size: 24px; }
        h2 { color: #111; margin-top: 20px; font-size: 20px; border-bottom: 1px solid #e2e8f0; padding-bottom: 4px; }
        h3 { color: #428787; margin-bottom: 12px; border-bottom: 1px solid #e2e8f0; padding-bottom: 4px; font-size: 18px; }
        .grid { display: grid; grid-template-columns: 1fr; gap: 20px; margin-bottom: 20px; }
        .box { background: #f8fafc; padding: 15px; border-radius: 6px; border: 1px solid #e2e8f0; }
        .form-group { margin-bottom: 15px; }
        label { display: block; font-size: 14px; font-weight: bold; color: #475569; margin-bottom: 6px; }
        input[type="number"], input[type="text"] { width: 100%; padding: 8px; border: 1px solid #cbd5e1; border-radius: 4px; font-size: 14px; box-sizing: border-box; }
        button { background: #428787; color: white; border: none; padding: 12px; cursor: pointer; font-size: 16px; border-radius: 4px; width: 100%; font-weight: bold; transition: background 0.2s; }
        button:hover { background: #336666; }
        .error { color: #b91c1c; font-weight: bold; background: #fee2e2; padding: 12px; border-radius: 4px; border: 1px solid #fca5a5; margin-bottom: 20px; }
        table { width: 100%; border-collapse: collapse; margin-top: 15px; margin-bottom: 20px; font-size: 14px; }
        th, td { border: 1px solid #e2e8f0; padding: 8px 12px; text-align: left; }
        th { background-color: #428787 !important; color: white !important; } /* Erzwingt das originale Blau gegen die website.css */
        tr:nth-child(even) { background-color: #f8fafc; }
        .r { text-align: right; }
        .hint { color: #64748b; font-size: 12px; margin-top: 4px; display: block; font-weight: normal; }
    </style>
</head>
<body>
    <div class="card">
        <h1>Durchlaufträger nach dem Cossverfahren</h1>
        <p style="color: #428787; font-style: italic; margin-top: -5px; margin-bottom: 20px; font-size: 14px;">Systembelastung: Konstante Gleichlast q = 10.0 kN/m</p>

        {{if .Fehler}}
            <div class="error">{{.Fehler}}</div>
        {{end}}

        <form method="POST">
    <div class="grid">
        <div class="box">
            <h3>Eingabeparameter Träger</h3>
            <div class="form-group">
                <label for="felder">Anzahl der Felder:</label>
                <!-- value wird nur befüllt, wenn bereits berechnet wurde; die 3 ist jetzt der graue Platzhalter -->
                <input type="number" id="felder" name="felder" placeholder="3" value="{{if .Berechnet}}{{.AnzahlFelder}}{{end}}" min="1" required>
            </div>
            <div class="form-group">
                <label for="laengen">Feldlängen in Metern (mit Komma getrennt):</label>
                <input type="text" id="laengen" name="laengen" value="{{.LaengenInput}}" placeholder="z.B. 4.5, 6.0, 4.5" required>
                <span class="hint">Gib die Längen der Felder der Reihe nach ein (z.B. 5, 4.5, 5)</span>
            </div>
            <button type="submit">Berechnung starten</button>

            {{if .Berechnet}}
                <h2 class="result-title">Berechnungsergebnisse</h2>
                <table>
                    <thead>
                        <tr>
                            <th>Stabseite</th>
                            <th class="r">µ</th>
                            <th class="r">Moment (kNm)</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .Seiten}}
                        <tr>
                            <td><strong>{{.Name}}</strong></td>
                            <td class="r">{{printf "%.2f" .Mu}}</td>
                            <td class="r">{{printf "%.2f" .MTotal}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            {{end}}
        </div>
    </div>
</form>

    </div>
</body>
</html>
`

func main() {
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Del("X-Frame-Options")
		data := PageData{}

		if r.Method == http.MethodPost {
			anzahlStr := r.FormValue("felder")
			laengenStr := r.FormValue("laengen")
			data.LaengenInput = laengenStr

			anzahl, err := strconv.Atoi(anzahlStr)
			if err != nil || anzahl < 1 {
				data.Fehler = "Ungültige Anzahl an Feldern."
				tmpl.Execute(w, data)
				return
			}
			data.AnzahlFelder = anzahl

			var laengen []float64
			var aktuellesZahl string

			for _, char := range laengenStr + "," {
				if char == ',' || char == ';' || char == ' ' {
					if aktuellesZahl != "" {
						val, err := strconv.ParseFloat(aktuellesZahl, 64)
						if err != nil || val <= 0 {
							data.Fehler = fmt.Sprintf("Ungültige Länge eingegeben: %s", aktuellesZahl)
							tmpl.Execute(w, data)
							return
						}
						laengen = append(laengen, val)
						aktuellesZahl = ""
					}
				} else {
					aktuellesZahl += string(char)
				}
			}

			if len(laengen) != anzahl {
				data.Fehler = fmt.Sprintf("Eingegebene Längen (%d) stimmen nicht mit der Feldanzahl (%d) überein.", len(laengen), anzahl)
				tmpl.Execute(w, data)
				return
			}

			const q = 10.0
			anzahlSeiten := anzahl * 2
			seiten := make([]Stabseite, anzahlSeiten)
			kSteifigkeit := make([]float64, anzahlSeiten)

			for i := 0; i < anzahl; i++ {
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
				} else if i == anzahl-1 {
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

			if anzahl == 1 {
				seiten[0].M0 = 0
				seiten[1].M0 = 0
			}

			anzahlKnoten := anzahl + 1
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

			for i := range seiten {
				roundedM := math.Round(seiten[i].MTotal*100) / 100
				if math.Abs(roundedM) == 0.0 {
					roundedM = 0.0
				}
				seiten[i].MTotal = roundedM
			}

			data.Berechnet = true
			data.Seiten = seiten
		}

		tmpl.Execute(w, data)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}
	fmt.Println("Server läuft online...")
	http.ListenAndServe("0.0.0.0:"+port, nil)
}
