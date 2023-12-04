package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
	modem := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/broadbandstatistics.ha" {
			t.Errorf("unexpected request path: %s", r.URL.Path)
		}

		file, err := os.Open("fixtures/partial.html")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		io.Copy(w, file)
	}))
	t.Cleanup(modem.Close)

	client := &http.Client{Timeout: time.Second * 5}
	handler := newHandler(client, modem.URL)
	svr := httptest.NewServer(handler)
	t.Cleanup(svr.Close)

	resp, err := http.Get(svr.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "att_modem_network_receive_bytes_total{} 16415238888") ||
		!strings.Contains(string(body), "att_modem_network_transmit_bytes_total{} 841324175") {
		t.Errorf("unexpected body: %s", body)
	}
}
