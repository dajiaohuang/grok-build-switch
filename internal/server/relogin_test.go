package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"grok_switch/internal/cpamint"
	"grok_switch/internal/grokpool"
	"grok_switch/internal/registrar"
)

func TestRefreshCookieRejectsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	pool, err := grokpool.NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	registrarService, err := registrar.NewService(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer registrarService.Close()

	s := &Server{GrokPool: pool, Registrar: registrarService}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/grok-pool/refresh-cookie", strings.NewReader("{"))
	s.handleGrokPoolRefreshCookie(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCpaMintCancelRejectsInvalidJSON(t *testing.T) {
	pool, err := grokpool.NewManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	s := &Server{CpaMint: cpamint.NewService(), GrokPool: pool}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/cpa-mint", strings.NewReader("{"))
	s.handleCpaMint(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}
