// Command et0-fao56 computes the FAO-56 Penman-Monteith reference
// evapotranspiration ET0 and the crop evapotranspiration ETc = Kc*ET0.
//
// CLI:
//
//	et0-fao56 -et0 example/arid-day.json          ET0 and its two terms
//	et0-fao56 -et0 example/arid-day.json -calm     same day next to a no-wind reference
//	et0-fao56 -etc example/arid-day.json          ET0 and ETc
//	et0-fao56 -http :8080                          Web console
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"

	"et0-fao56/internal/crop"
	"et0-fao56/internal/penman"
)

//go:embed web
var webFS embed.FS

//go:embed example/arid-day.json
var aridDayJSON []byte

const maxRequestBytes = 1 << 20

func main() {
	httpAddr := flag.String("http", "", "serve the Web console on this address (e.g. :8080)")
	et0File := flag.String("et0", "", "compute ET0 for a document and print the report")
	etcFile := flag.String("etc", "", "compute ET0 and ETc for a document and print the report")
	calm := flag.String("calm", "", "with -et0: also print the no-wind reference (any non-empty value turns it on)")
	windSpeed := flag.Float64("wind", -1, "override the 2 m wind speed in m/s (negative keeps the document value)")
	flag.Parse()

	switch {
	case *httpAddr != "":
		if err := serveHTTP(*httpAddr); err != nil {
			log.Fatal(err)
		}
	case *et0File != "":
		if err := runET0(*et0File, *calm != "", *windSpeed); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case *etcFile != "":
		if err := runETc(*etcFile, *windSpeed); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	default:
		flag.Usage()
		os.Exit(2)
	}
}

func runET0(path string, withCalm bool, windOverride float64) error {
	doc, err := crop.LoadFile(path)
	if err != nil {
		return err
	}
	if doc.Weather == nil {
		return fmt.Errorf("%s: the document has no weather block", path)
	}
	weather := *doc.Weather
	if windOverride >= 0 {
		weather = weather.WithWindSpeed(windOverride)
	}
	scale, err := doc.Scale()
	if err != nil {
		return err
	}
	if withCalm {
		actual, calmResult, err := penman.CalmComparison(weather, scale)
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, actual.String())
		fmt.Fprint(os.Stdout, "\n", penman.CompareCalm(actual, calmResult))
		return nil
	}
	res, err := penman.Compute(weather, scale)
	if err != nil {
		return err
	}
	fmt.Fprint(os.Stdout, res.String())
	return nil
}

func runETc(path string, windOverride float64) error {
	doc, err := crop.LoadFile(path)
	if err != nil {
		return err
	}
	if windOverride >= 0 {
		if doc.Weather == nil {
			return fmt.Errorf("%s: cannot override the wind speed without a weather block", path)
		}
		weather := doc.Weather.WithWindSpeed(windOverride)
		doc.Weather = &weather
		doc.ET0 = nil
	}
	out, err := doc.EvaluateCrop()
	if err != nil {
		return err
	}
	if out.Reference != nil {
		fmt.Fprint(os.Stdout, out.Reference.String(), "\n")
	}
	fmt.Fprint(os.Stdout, out.Crop.String())
	return nil
}

type et0Response struct {
	Site      string             `json:"site,omitempty"`
	Result    *penman.Result     `json:"result,omitempty"`
	WindSweep []penman.WindPoint `json:"windSweep,omitempty"`
	Error     string             `json:"error,omitempty"`
}

type etcResponse struct {
	Site      string             `json:"site,omitempty"`
	Scale     penman.TimeScale   `json:"scale,omitempty"`
	ET0       float64            `json:"et0"`
	ET0Source string             `json:"et0Source,omitempty"`
	Reference *penman.Result     `json:"reference,omitempty"`
	Crop      *crop.Result       `json:"crop,omitempty"`
	WindSweep []penman.WindPoint `json:"windSweep,omitempty"`
	Error     string             `json:"error,omitempty"`
}

func handleET0(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, err := readBody(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, et0Response{Error: err.Error()})
		return
	}
	doc, err := crop.LoadBytes(body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, et0Response{Error: err.Error()})
		return
	}
	if doc.Weather == nil {
		writeJSON(w, http.StatusBadRequest, et0Response{Error: "/api/et0 needs the weather block"})
		return
	}
	out, err := doc.Evaluate()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, et0Response{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, et0Response{
		Site:      out.Site,
		Result:    out.Reference,
		WindSweep: out.WindSweep,
	})
}

func handleETc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, err := readBody(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, etcResponse{Error: err.Error()})
		return
	}
	doc, err := crop.LoadBytes(body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, etcResponse{Error: err.Error()})
		return
	}
	out, err := doc.EvaluateCrop()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, etcResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, etcResponse{
		Site:      out.Site,
		Scale:     out.Scale,
		ET0:       out.ET0,
		ET0Source: out.ET0Source,
		Reference: out.Reference,
		Crop:      out.Crop,
		WindSweep: out.WindSweep,
	})
}

func handleExample(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, et0Response{Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, json.RawMessage(aridDayJSON))
}

func serveHTTP(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/et0", handleET0)
	mux.HandleFunc("/api/etc", handleETc)
	mux.HandleFunc("/api/example", handleExample)
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		return err
	}
	mux.Handle("/", http.FileServer(http.FS(sub)))
	log.Printf("et0-fao56 web console on http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
}

func readBody(r *http.Request) ([]byte, error) {
	if r.Method != http.MethodPost {
		return nil, fmt.Errorf("method %s not allowed, use POST", r.Method)
	}
	if r.Body == nil {
		return nil, fmt.Errorf("empty request body")
	}
	defer r.Body.Close()
	if r.ContentLength > maxRequestBytes {
		return nil, fmt.Errorf("request body too large")
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if int64(len(data)) > maxRequestBytes {
		return nil, fmt.Errorf("request body too large")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty request body")
	}
	return data, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}
