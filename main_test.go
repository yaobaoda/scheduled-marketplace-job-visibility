package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunHandlerSurfacesHandoffFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"data":{},"error":null,"metadata":{}}`))
	}))
	defer server.Close()
	c := &InfraiClient{Key: "test", HTTP: server.Client(), Base: server.URL}
	req := httptest.NewRequest(http.MethodPost, "/jobs/run", strings.NewReader(`{"seller_asset":"sku-7","buyer_update":"packed","order_id":"o-9","handoff_ok":false}`))
	res := httptest.NewRecorder()
	runHandler(c).ServeHTTP(res, req)
	if !strings.Contains(res.Body.String(), `"status":"failed"`) {
		t.Fatalf("response=%s", res.Body.String())
	}
}

func TestRunHandlerSurfacesCaptureError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"ok":false,"error":{"code":"INVALID_REQUEST"}}`))
	}))
	defer server.Close()
	c := &InfraiClient{Key: "test", HTTP: server.Client(), Base: server.URL}
	req := httptest.NewRequest(http.MethodPost, "/jobs/run", strings.NewReader(`{"seller_asset":"sku-7","buyer_update":"packed","order_id":"o-9","handoff_ok":false}`))
	res := httptest.NewRecorder()
	runHandler(c).ServeHTTP(res, req)
	if res.Code != http.StatusBadGateway || !strings.Contains(res.Body.String(), "error capture failed") {
		t.Fatalf("status=%d response=%s", res.Code, res.Body.String())
	}
}
