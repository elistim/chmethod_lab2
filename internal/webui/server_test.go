package webui

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"chmethod_lab2/internal/numerics"
)

func TestHTTP(t *testing.T) {
	h := Handler()
	input, _ := json.Marshal(numerics.DefaultInterpolation())
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/", "", 200}, {"GET", "/app.js", "", 200}, {"GET", "/style.css", "", 200}, {"GET", "/api/defaults", "", 200},
		{"POST", "/api/interpolate", string(input), 200}, {"POST", "/api/interpolate", "{}", 400},
		{"POST", "/api/integrate", `{"a":-0.75,"b":0.75,"n":8,"order":8,"tolerance":0.000001,"method":"gauss"}`, 200},
		{"POST", "/api/integrate", "{}{}", 400}, {"POST", "/api/integrate", `{"unknown":1}`, 400}, {"GET", "/missing", "", 404},
	} {
		t.Run(tc.method+tc.path+tc.body[:min(len(tc.body), 12)], func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("status %d: %s", w.Code, w.Body.String())
			}
		})
	}
}
