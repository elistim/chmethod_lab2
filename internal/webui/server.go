package webui

import (
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"

	"chmethod_lab2/internal/numerics"
)

//go:embed static/*
var assets embed.FS

func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/defaults", func(w http.ResponseWriter, r *http.Request) { send(w, http.StatusOK, numerics.DefaultInterpolation()) })
	mux.HandleFunc("POST /api/interpolate", func(w http.ResponseWriter, r *http.Request) {
		var in numerics.InterpolationInput
		if !decode(w, r, &in) {
			return
		}
		out, err := numerics.Interpolate(in)
		if err != nil {
			send(w, 400, map[string]string{"error": err.Error()})
			return
		}
		send(w, 200, out)
	})
	mux.HandleFunc("POST /api/integrate", func(w http.ResponseWriter, r *http.Request) {
		var in numerics.IntegrationInput
		if !decode(w, r, &in) {
			return
		}
		out, err := numerics.Integrate(in)
		if err != nil {
			send(w, 400, map[string]string{"error": err.Error()})
			return
		}
		send(w, 200, out)
	})
	static, _ := fs.Sub(assets, "static")
	mux.Handle("GET /", http.FileServer(http.FS(static)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; object-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		mux.ServeHTTP(w, r)
	})
}
func send(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "Ошибка сериализации результата", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}
func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 128<<10)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		send(w, 400, map[string]string{"error": "Некорректный JSON или слишком большой запрос"})
		return false
	}
	if err := d.Decode(new(any)); err != io.EOF {
		send(w, 400, map[string]string{"error": "Ожидается один JSON-объект"})
		return false
	}
	return true
}
