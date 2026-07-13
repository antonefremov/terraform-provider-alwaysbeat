package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateCheckSendsInputAndParsesResponse(t *testing.T) {
	var gotBody CheckInput
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/checks" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("X-DMF-Token"); got != "dmf_test" {
			t.Errorf("token header = %q; want dmf_test", got)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Check{
			CheckID: "abc", Name: gotBody.Name, Status: "new",
			ScheduleKind: "interval", IntervalS: 3600, TZ: "UTC", GraceS: 300,
			PingURL: "https://ingest.example/ping/abc",
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "dmf_test")
	out, err := c.CreateCheck(context.Background(), CheckInput{
		Name:     "job",
		Schedule: Schedule{Kind: "interval", Interval: "1h", TZ: "UTC"},
		Grace:    "5m",
	})
	if err != nil {
		t.Fatalf("CreateCheck: %v", err)
	}
	if gotBody.Schedule.Interval != "1h" || gotBody.Grace != "5m" {
		t.Errorf("request body not sent as durations: %+v", gotBody)
	}
	if out.CheckID != "abc" || out.PingURL == "" || out.GraceS != 300 {
		t.Errorf("response not parsed: %+v", out)
	}
}

func TestGetCheckNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	_, err := New(srv.URL, "k").GetCheck(context.Background(), "missing")
	if !isErrNotFound(err) {
		t.Fatalf("GetCheck(missing) = %v; want ErrNotFound", err)
	}
}

func TestDeleteCheckToleratesMissing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	if err := New(srv.URL, "k").DeleteCheck(context.Background(), "gone"); err != nil {
		t.Fatalf("DeleteCheck on 404 should be nil, got %v", err)
	}
}

func TestErrorMessageSurfaced(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"plan \"free\" allows 20 checks"}`))
	}))
	defer srv.Close()
	_, err := New(srv.URL, "k").CreateCheck(context.Background(), CheckInput{Name: "x"})
	if err == nil || !containsStr(err.Error(), "plan \"free\" allows 20 checks") {
		t.Fatalf("error message not surfaced: %v", err)
	}
}

func isErrNotFound(err error) bool { return err == ErrNotFound }

func containsStr(hay, needle string) bool {
	return len(hay) >= len(needle) && (hay == needle || indexOf(hay, needle) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
