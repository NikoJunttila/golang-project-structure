package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nikojunttila/community/internal/handlers"
)

func TestGetFooHandle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(handlers.GetFooHandler))
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 but got %d", resp.StatusCode)
	}
}
