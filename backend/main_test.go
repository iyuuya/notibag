package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestServer() *gin.Engine {
	repo := NewInMemoryNotificationRepository()
	service := NewNotificationService(repo)
	wsManager := NewWSManager(service)
	return setupRouter(NewNotificationHandler(service, wsManager))
}

func TestHealthCheck(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status field = %q, want ok", body["status"])
	}
}

func TestCreateAndListNotifications(t *testing.T) {
	srv := newTestServer()

	payload := []byte(`{"title":"hello","message":"world","type":"success"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/notifications", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var created Notification
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Title != "hello" || created.Message != "world" || created.Type != "success" || created.Read {
		t.Fatalf("created = %+v", created)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d", rec.Code)
	}

	var list NotificationsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Notifications) != 3 {
		t.Fatalf("unread count = %d, want 3 (2 seeded + 1 created)", len(list.Notifications))
	}
	if list.Notifications[0].ID != created.ID {
		t.Fatalf("newest id = %s, want %s", list.Notifications[0].ID, created.ID)
	}
}

func TestCreateNotificationRejectsMissingFields(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/api/notifications", bytes.NewReader([]byte(`{"title":"only title"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestMarkAsReadAndClear(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodPut, "/api/notifications/1/read", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("mark read status = %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	var list NotificationsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	for _, n := range list.Notifications {
		if n.ID == "1" {
			t.Fatal("read notification still listed as unread")
		}
	}

	req = httptest.NewRequest(http.MethodPut, "/api/notifications/missing/read", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing mark read status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/notifications", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("clear status = %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/notifications/all", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Notifications) != 0 {
		t.Fatalf("after clear count = %d, want 0", len(list.Notifications))
	}
}

func TestCORSPreflight(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodOptions, "/api/health", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("allow origin = %q, want *", got)
	}
}
