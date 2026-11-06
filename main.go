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
        h3 { color: #0284c7; margin-bottom: 12px; border-bottom: 1px solid #e2e8f0; padding-bottom: 4px; font-size: 18px; }
        .grid { display: grid; grid-template-columns: 1fr; gap: 20px; margin-bottom: 20px; }
        .box { background: #f8fafc; padding: 15px; border-radius: 6px; border: 1px solid #e2e8f0; }
        .form-group { margin-bottom: 15px; }
        label { display: block; font-size: 14px; font-weight: bold; color: #475569; margin-bottom: 6px; }
        input[type="number"], input[type="text"] { width: 100%; padding: 8px; border: 1px solid #cbd5e1; border-radius: 4px; font-size: 14px; box-sizing: border-box; }
        button { background: #0284c7; color: white; border: none; padding: 12px; cursor: pointer; font-size: 16px; border-radius: 4px; width: 100%; font-weight: bold; transition: background 0.2s; }
        button:hover { background: #0369a1; }
        .error { color: #b91c1c; font-weight: bold; background: #fee2e2; padding: 12px; border-radius: 4px; border: 1px solid #fca5a5; margin-bottom: 20px; }
        table { width: 100%; border-collapse: collapse; margin-top: 15px; margin-bottom: 20px; font-size: 14px; }
        th, td { border: 1px solid #e2e8f0; padding: 8px 12px; text-align: left; }
        th { background-color: #0f172a; color: white; }
        tr:nth-child(even) { background-color: #f8fafc; }
        .r { text-align: right; }
        .hint { color: #64748b; font-size: 12px; margin-top: 4px; display: block; font-weight: normal; }
    </style>
</head>
<body>
    <div class="card">
        <h1>Durchlaufträger-Berechnung (Cross-Verfahren)</h1>
        <p style="color: #64748b; font-style: italic; margin-top: -5px; margin-bottom: 20px; font-size: 14px;">Systembelastung: Konstante Gleichlast q = 10.0 kN/m</p>

        {{if .Fehler}}
            <div class="error">{{.Fehler}}</div>
        {{end}}

        <form method="POST">
            <div class="grid">
                <div class="box">
                    <h3>Eingabeparameter Träger</h3>
                    <div class="form-group">
                        <label for="felder">Anzahl der Felder:</label>
                        <input type="number" id="felder" name="felder" value="{{if .AnzahlFelder}}{{.AnzahlFelder}}{{else}}3{{end}}" min="1" required>
                    </div>
                    <div class="form-group">
                        <label for="laengen">Feldlängen in Metern (mit Komma getrennt):</label>
                        <input type="text" id="laengen" name="laengen" value="{{.LaengenInput}}", placeholder="z.B. 4.5, 6.0, 4.5" required>
                        <span class="hint">Gib die Längen der Felder der Reihe nach ein (z.B. 5, 4.5, 5)</span>
                    </div>
                </div>
            </div>
            <button type="submit">Berechnung starten</button>
        </form>

        {{if .Berechnet}}
            <h2>Berechnungsergebnisse</h2>
            <table>
                <thead>
                    <tr>
                        <th>Stabseite</th>
                        <th class="r">Verteilungszahl (µ)</th>
                        <th class="r">Stützmoment (kNm)</th>
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
</body>
</html>
`
