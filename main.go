package main

import (
	"html/template"
	"math"
	"net/http"
	"strconv"
	"strings"
)

// Stabseite repräsentiert ein Feldende an einem Knoten
type Stabseite struct {
	Name       string
	Mu         float64 // Verteilungszahl
	M0         float64 // Volleinspannmoment (Startwert)
	MRest      float64 // Unverteiltes Restmoment in der aktuellen Iteration
	MTotal     float64 // Endgültiges, aufsummiertes Stützmoment
	OppositeId int     // Index des gegenüberliegenden Endes desselben Feldes
}

// Knoten gruppiert die Stabseiten, die an einem Auflager zusammentreffen
type Knoten struct {
	Name    string
	SeitenI []int
}

// TemplateData steuert die Datenweitergabe an das HTML-Formular
type TemplateData struct {
	AnzahlFelder int
	LaengenInput string
	Seiten       []Stabseite
	Fehler       string
	Berechnet    bool
}

// HTML-Template direkt aus Ihrer Vorlage übernommen
const htmlTemplate = `<!DOCTYPE html>
<html lang="de">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Berechnung Durchlaufträger</title>
    <style>
        body {
            font-family: 'Roboto', sans-serif;
            max-width: 800px;
            margin: 20px auto;
            padding: 20px;
            background: #000000;
            color: #333;
        }

        .card {
            background: white;
            padding: 25px;
            border-radius: 5px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
        }

        h1 {
            color: #111;
            border-bottom: 2px solid #eee;
            padding-bottom: 8px;
            margin-top: 0;
            font-size: 24px;
        }

        h2 {
            color: #111;
            margin-top: 20px;
            font-size: 20px;
            border-bottom: 1px solid #e2e8f0;
            padding-bottom: 4px;
        }

        h3 {
            color: #428787;
            margin-bottom: 12px;
            border-bottom: 1px solid #e2e8f0;
            padding-bottom: 4px;
            font-size: 18px;
        }

        .grid {
            display: grid;
            grid-template-columns: 1fr;
            gap: 20px;
            margin-bottom: 20px;
        }

        .box {
            background: #f8fafc;
            padding: 15px;
            border-radius: 5px;
            border: 1px solid #e2e8f0;
        }

        .form-group {
            margin-bottom: 15px;
        }

        label {
            display: block;
            font-size: 14px;
            font-weight: bold;
            color: #475569;
            margin-bottom: 6px;
        }

        input[type="number"],
        input[type="text"] {
            width: 100%;
            padding: 8px;
            border: 1px solid #cbd5e1;
            font-size: 16px;
            box-sizing: border-box;
        }

        button {
            background: #428787;
            color: white;
            border: none;
            padding: 12px;
            cursor: pointer;
            font-size: 16px;
            border-radius: 4px;
            width: 100%;
            font-weight: bold;
            transition: background 0.2s;
        }

        button:hover {
            background: #336666;
        }

        .error {
            color: #b91c1c;
            font-weight: bold;
            background: #fee2e2;
            padding: 12px;
            border-radius: 4px;
            border: 1px solid #fca5a5;
            margin-bottom: 20px;
        }

        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 15px;
            margin-bottom: 20px;
            font-size: 14px;
        }

        th,
        td {
            border: 1px solid #e2e8f0;
            padding: 8px 12px;
            text-align: left;
        }

        th {
            background-color: #428787 !important;
            color: white !important;
        }

        tr:nth-child(even) {
            background-color: #f8fafc;
        }

        .r {
            text-align: right;
        }

        .hint {
            color: #64748b;
            font-size: 12px;
            margin-top: 4px;
            display: block;
            font-weight: normal;
        }

    </style>
</head>


<body>
    <div class="card">
        <h1>Berechnung Durchlaufträger</h1>
        <p style="color: #428787; font-style: italic; margin-top: -5px; margin-bottom: 20px; font-size: 14px;">Systembelastung: Konstante Gleichlast q = 10.0 kN/m</p>

        {{if .Fehler}}
        <div class="error">{{.Fehler}}</div>
        {{end}}

        <form method="POST" id="traegerForm">
            <div class="grid">
                <div class="box">
                    <h3>Eingabeparameter Träger</h3>
                    <div class="form-group">
                        <label for="felder">Anzahl der Felder:</label>
                        <input type="number" id="felder" name="felder" placeholder="3" value="{{if .Berechnet}}{{.AnzahlFelder}}{{end}}" min="1" required>
                    </div>
                    <div class="form-group">
                        <label for="laengen">Feldlängen in Metern (mit Komma getrennt):</label>
                        <input type="text" id="laengen" name="laengen" value="{{.LaengenInput}}" placeholder="z.B. 4.5, 6.0, 4.5" required>
                        <span class="hint">Gib die Längen der Felder der Reihe nach ein (z.B. 5, 4.5, 5)</span>
                    </div>
                    
            <div><h6>&copy; Ing. Burkhard Kaar - Studienprojekt imperative Programmierung, Berechnung eines Durchlaufträgers nach dem Cross-Verfahren</h6></div>

                    <button type="submit" id="submitBtn" disabled>Berechnung starten</button>

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

        <style>
            /* Standard-Design für den aktiven Button */
            button {
                background: #428787;
                color: white;
                border: none;
                padding: 12px;
                cursor: pointer;
                font-size: 16px;
                border-radius: 4px;
                width: 100%;
                font-weight: bold;
                transition: all 0.2s;
            }

            button:hover:not(:disabled) {
                background: #336666;
            }

            /* Automatisches Ausgrauen, wenn der Button deaktiviert ist */
            button:disabled {
                background: #cbd5e1;
                /* Ein helles Grau */
                color: #94a3b8;
                /* Dunklerer Text für Lesbarkeit im Grau */
                cursor: not-allowed;
                /* Zeigt ein "Verboten"-Symbol beim Drüberfahren */
                opacity: 0.7;
            }

        </style>

        <!-- Dieser kleine JavaScript-Block steuert die Freischaltung live beim Tippen -->
        <script>
            const form = document.getElementById('traegerForm');
            const submitBtn = document.getElementById('submitBtn');

            function pruefeFormular() {
                // checkValidity() prüft automatisch alle Felder mit "required"
                if (form.checkValidity()) {
                    submitBtn.removeAttribute('disabled');
                } else {
                    submitBtn.setAttribute('disabled', 'true');
                }
            }

            // Lauscht auf jede Eingabe in den Formularfeldern
            form.addEventListener('input', pruefeFormular);

            // Einmal beim Laden ausführen (falls nach Berechnung bereits Werte drinstehen)
            pruefeFormular();
            
            document.addEventListener('DOMContentLoaded', function() {
    var form = document.getElementById('stromForm');
    if (form) {
        form.addEventListener('input', pruefeFormular);
    }
    pruefeFormular();

    // DIESER TEIL SCHIEBT DAS FORMULAR NACH OBEN, WENN DIE ERGEBNISSE DA SIND:
    // Der Server fügt bei erfolgreicher Berechnung das <h2>-Element ein.
    // Sobald dieses Element im Browser existiert, scrollt die Seite automatisch dorthin.
    var ergebnisUeberschrift = document.querySelector('h2');
    if (ergebnisUeberschrift) {
        ergebnisUeberschrift.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
});

        </script>

    </div>
</body>

</html>`

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))

	if r.Method != http.MethodPost {
		tmpl.Execute(w, TemplateData{})
		return
	}

	// 1. Formulardaten auslesen
	felderStr := r.FormValue("felder")
	laengenStr := r.FormValue("laengen")

	anzahlFelder, err := strconv.Atoi(felderStr)
	if err != nil || anzahlFelder < 1 {
		tmpl.Execute(w, TemplateData{Fehler: "Ungültige Anzahl an Feldern.", LaengenInput: laengenStr})
		return
	}

	// String-Verarbeitung der Feldlängen (Kommas trennen, Leerzeichen entfernen)
	teile := strings.Split(laengenStr, ",")
	var laengen []float64
	for _, teil := range teile {
		t := strings.TrimSpace(teil)
		if t == "" {
			continue
		}
		val, err := strconv.ParseFloat(t, 64)
		if err != nil || val <= 0 {
			tmpl.Execute(w, TemplateData{Fehler: "Ungültiges Format bei den Feldlängen. Bitte Zahlen durch Komma trennen (z.B. 4, 6, 4).", LaengenInput: laengenStr})
			return
		}
		laengen = append(laengen, val)
	}

	if len(laengen) != anzahlFelder {
		tmpl.Execute(w, TemplateData{Fehler: "Die Anzahl der eingegebenen Feldlängen stimmt nicht mit der Anzahl der Felder überein.", LaengenInput: laengenStr})
		return
	}

	// 2. Mathematische Berechnung über das korrigierte Cross-Verfahren
	const q = 10.0 // Belastung kN/m
	anzahlSeiten := anzahlFelder * 2
	seiten := make([]Stabseite, anzahlSeiten)
	kSteifigkeit := make([]float64, anzahlSeiten)

	// Festeinspannmomente und elastische Steifigkeiten generieren
	for i := 0; i < anzahlFelder; i++ {
		l := laengen[i]
		leftIdx := i * 2
		rightIdx := i*2 + 1

		charLinks := string(rune('A' + i))
		charRechts := string(rune('A' + i + 1))
		seiten[leftIdx].Name = charLinks + "-rechts"
		seiten[rightIdx].Name = charRechts + "-links"

		seiten[leftIdx].OppositeId = rightIdx
		seiten[rightIdx].OppositeId = leftIdx

		// Volleinspannung als Referenzsystem
		kSteifigkeit[leftIdx] = 1.0 / l
		kSteifigkeit[rightIdx] = 1.0 / l

		seiten[leftIdx].M0 = -(q * l * l) / 12.0
		seiten[rightIdx].M0 = (q * l * l) / 12.0
	}

	if anzahlFelder == 1 {
		seiten[0].M0 = 0
		seiten[1].M0 = 0
	}

	// Knoten und Verteilwerte µ generieren
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

	// Startwerte setzen
	for i := range seiten {
		seiten[i].MTotal = seiten[i].M0
		seiten[i].MRest = seiten[i].M0
	}

	// Iterativer Ausgleichsprozess
	maxIterationen := 40
	epsilon := 0.0001
	mAuszugleichen := make([]float64, anzahlSeiten)

	for iter := 1; iter <= maxIterationen; iter++ {
		// Schritt A: Knotenausgleich ermitteln
		for _, k := range knoten {
			knotenRestSumme := 0.0
			for _, idx := range k.SeitenI {
				knotenRestSumme += seiten[idx].MRest
			}

			for _, idx := range k.SeitenI {
				mAusgleich := -knotenRestSumme * seiten[idx].Mu
				seiten[idx].MTotal += mAusgleich
				mAuszugleichen[idx] = mAusgleich
			}
		}

		// Konvergenz prüfen
		groessteAenderung := 0.0
		for _, m := range mAuszugleichen {
			if math.Abs(m) > groessteAenderung {
				groessteAenderung = math.Abs(m)
			}
		}
		if groessteAenderung < epsilon {
			break
		}

		// SCHRITT B: Fortleitung
		for i := range seiten {
			seiten[i].MRest = 0.0
		}
		for i := range seiten {
			oppIdx := seiten[i].OppositeId
			fortleitung := mAuszugleichen[i] * 0.5

			seiten[oppIdx].MRest = fortleitung
			seiten[oppIdx].MTotal += fortleitung
		}

	}

	// Bereinigung für die Anzeige (-0.00 korrigieren)
	for i := range seiten {
		seiten[i].MTotal = math.Round(seiten[i].MTotal*100) / 100
		if math.Abs(seiten[i].MTotal) == 0.0 {
			seiten[i].MTotal = 0.0
		}
	}

	// 3. Daten zurück ans HTML senden
	data := TemplateData{
		AnzahlFelder: anzahlFelder,
		LaengenInput: laengenStr,
		Seiten:       seiten,
		Berechnet:    true,
	}
	tmpl.Execute(w, data)
}

func main() {
	println("Webserver gestartet unter: http://localhost:8080")
	http.HandleFunc("/", handleIndex)
	http.ListenAndServe(":8080", nil)
}
