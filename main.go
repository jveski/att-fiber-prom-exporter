package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	var (
		addr    = flag.String("addr", ":8080", "address to listen on")
		url     = flag.String("url", "http://192.168.1.254", "base URL of the modem")
		timeout = flag.Duration("timeout", time.Second*10, "timeout for requests to the modem")
	)
	flag.Parse()

	client := &http.Client{Timeout: *timeout}
	mux := newHandler(client, *url)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func newHandler(client *http.Client, url string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics, err := poll(r.Context(), client, url)
		if err != nil {
			log.Printf("error while polling modem: %s", err)
			w.WriteHeader(500)
			return
		}
		lines := []string{
			"# HELP att_modem_network_receive_bytes_total The total number of transmitted bytes.",
			"# TYPE att_modem_network_receive_bytes_total counter",
			`att_modem_network_receive_bytes_total{} ` + strconv.FormatInt(metrics.RX, 10),

			"# HELP att_modem_network_transmit_bytes_total The total number of received bytes.",
			"# TYPE att_modem_network_transmit_bytes_total counter",
			`att_modem_network_transmit_bytes_total{} ` + strconv.FormatInt(metrics.TX, 10),
			"",
		}
		io.WriteString(w, strings.Join(lines, "\n"))
	})

	return mux
}

type metrics struct {
	RX, TX int64
}

var re = regexp.MustCompile(`<th scope="row" width=".*">(Receive Bytes|Transmit Bytes)</th>\s*<td>(\d+)</td>`)

func poll(ctx context.Context, client *http.Client, url string) (*metrics, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/cgi-bin/broadbandstatistics.ha", url), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	m := &metrics{}
	for _, match := range re.FindAllSubmatch(raw, -1) {
		if len(match) < 3 {
			continue
		}
		value, _ := strconv.ParseInt(string(match[2]), 10, 0)
		switch string(match[1]) {
		case "Receive Bytes":
			m.RX = value
		case "Transmit Bytes":
			m.TX = value
		}
	}

	return m, nil
}
