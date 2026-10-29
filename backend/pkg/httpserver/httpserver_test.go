package httpserver

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestServerBaseContextCancelsRequests(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	if err := listener.Close(); err != nil {
		t.Fatalf("listener.Close: %v", err)
	}

	baseCtx, cancelBase := context.WithCancel(context.Background())
	defer cancelBase()

	started := make(chan struct{})
	canceled := make(chan struct{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)

		<-r.Context().Done()

		close(canceled)
	})

	server := New(Config{Port: strconv.Itoa(port)}, handler, baseCtx)

	go func() {
		_ = server.Start()
	}()

	waitForPort(t, port)

	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		"http://127.0.0.1:"+strconv.Itoa(port)+"/stream", nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext: %v", err)
	}

	go func() {
		_, _ = client.Do(req)
	}()

	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("handler was not called")
	}

	cancelBase()

	select {
	case <-canceled:
	case <-time.After(3 * time.Second):
		t.Fatal("request context was not canceled with the base context")
	}
}

func waitForPort(t *testing.T, port int) {
	t.Helper()

	address := "127.0.0.1:" + strconv.Itoa(port)
	deadline := time.Now().Add(3 * time.Second)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()

			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("server did not start on %s", address)
}
