package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VoolFI71/url-shortener/internal/shortener"
	"github.com/VoolFI71/url-shortener/internal/storage/memory"
)

type fixedGenerator struct {
	code string
}

func (g fixedGenerator) Generate() (string, error) {
	return g.code, nil
}

func TestCreateAndRedirect(t *testing.T) {
	const code = "Abc123_Xyz"
	handler := newTestHandler(code)

	response := postURL(t, handler, `{"url":"https://example.com/article"}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want %d", response.Code, http.StatusCreated)
	}

	var body createURLResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ShortURL != "http://short.test/"+code {
		t.Fatalf("POST response = %+v", body)
	}

	request := httptest.NewRequest(http.MethodGet, "/"+code, nil)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusFound {
		t.Fatalf("GET status = %d, want %d", response.Code, http.StatusFound)
	}
	if location := response.Header().Get("Location"); location != "https://example.com/article" {
		t.Fatalf("Location = %q, want %q", location, "https://example.com/article")
	}
}

func TestCreateReturnsExistingLink(t *testing.T) {
	handler := newTestHandler("Abc123_Xyz")
	body := `{"url":"https://example.com"}`

	first := postURL(t, handler, body)
	second := postURL(t, handler, body)

	if first.Code != http.StatusCreated {
		t.Fatalf("first POST status = %d, want %d", first.Code, http.StatusCreated)
	}
	if second.Code != http.StatusOK {
		t.Fatalf("second POST status = %d, want %d", second.Code, http.StatusOK)
	}
}

func TestCreateRejectsInvalidRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed JSON", body: `{"url":`},
		{name: "invalid URL", body: `{"url":"example.com"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := postURL(t, newTestHandler("Abc123_Xyz"), test.body)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestRedirectReturnsNotFound(t *testing.T) {
	handler := newTestHandler("Abc123_Xyz")
	request := httptest.NewRequest(http.MethodGet, "/NotFound01", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func newTestHandler(code string) http.Handler {
	service := shortener.New(memory.New(), fixedGenerator{code: code})
	return New(service, "http://short.test")
}

func postURL(t *testing.T, handler http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
