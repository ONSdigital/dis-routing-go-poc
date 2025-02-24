package service_test

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dis-routing-go-poc/service"
)

const (
	numRequestsMultiple = 100
	delayMultiple       = 100

	numRequestsStaggered = 100
	delayStaggered       = 10

	numConfigRedirects = 1000
)

// TestMain runs before any tests and applies globally for all tests in the package.
func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestService_Run(t *testing.T) {
	svc, err := startServiceInstance()
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Shutdown()

	// Wait a little bit of time for server to finish starting up
	time.Sleep(time.Millisecond * 100)

	// Add some routes and redirects
	err = addRedirect(svc.AdminPort, "/redir1", "http://localhost/redirected1")
	if err != nil {
		t.Fatal(err)
	}
	err = addRoute(svc.AdminPort, "/route1", fmt.Sprintf("http://localhost:%d", svc.UpstreamPort))
	if err != nil {
		t.Fatal(err)
	}

	nrClient := getRouterClient(svc.RouterPort)

	t.Run("redirect returns", func(t *testing.T) {
		result, err := nrClient.Send("/redir1", nil)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if result.Status != http.StatusTemporaryRedirect {
			t.Errorf("expected status code %v, got %v", http.StatusTemporaryRedirect, result.Status)
		}

		if result.Location != "http://localhost/redirected1" {
			t.Errorf("expected Location header \"http://localhost/redirected1\", got %v", result.Location)
		}
	})

	t.Run("route returns", func(t *testing.T) {
		result, err := nrClient.Send("/route1", nil)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if result.Status != http.StatusOK {
			t.Errorf("expected status code %v, got %v", http.StatusOK, result.Status)
		}

		if result.Location != "" {
			t.Errorf("expected no Location header, got %v", result.Location)
		}
	})

}

func TestService_Run_Reroute(t *testing.T) {
	svc, err := startServiceInstance()
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Shutdown()

	// Wait a little bit of time for server to finish starting up
	time.Sleep(time.Millisecond * 100)

	// Add some route and redirects
	err = addRoute(svc.AdminPort, "/route1", fmt.Sprintf("http://localhost:%d", svc.UpstreamPort))
	if err != nil {
		t.Fatal(err)
	}

	nrClient := getRouterClient(svc.RouterPort)

	t.Run("multiple_redirects", func(t *testing.T) {

		// Fire off some requests with a delay
		preresults := make([]*result, numRequestsMultiple)
		prewg := &sync.WaitGroup{}
		for i := 0; i < numRequestsMultiple; i++ {
			prewg.Add(1)
			go func() {
				defer prewg.Done()
				r, err := nrClient.Send("/route1", &sendOptions{UpstreamDelay: delayMultiple})
				if err != nil {
					t.Errorf("expected no error, got %v, instance %d", err, i)
				}
				preresults[i] = r
			}()
		}

		// Sleep until half way through sleep time of pre requests
		time.Sleep((delayMultiple / 2) * time.Millisecond)

		// Add another route
		err = addRoute(svc.AdminPort, "/route2", fmt.Sprintf("http://localhost:%d", svc.UpstreamPort))
		if err != nil {
			t.Fatal(err)
		}

		// Fire off some more requests with a delay
		postresults := make([]*result, numRequestsMultiple)
		postwg := &sync.WaitGroup{}
		for i := 0; i < numRequestsMultiple; i++ {
			postwg.Add(1)
			go func() {
				defer postwg.Done()
				r, err := nrClient.Send("/route1", &sendOptions{UpstreamDelay: delayMultiple})
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				postresults[i] = r
			}()
		}

		// Collect and test the results

		// Test pre requests
		prewg.Wait()
		for i := 0; i < numRequestsMultiple; i++ {
			r := preresults[i]
			if r == nil {
				t.Errorf("expected a valid r, index %d", i)
				continue
			}

			if r.Status != http.StatusOK {
				t.Errorf("expected status code %v, got %v, index %d", http.StatusOK, r.Status, i)
			}

			if r.RouterVersion != "2" {
				t.Errorf("expected router version 2, got %v, index %d", r.RouterVersion, i)
			}

		}

		// Test post-results
		postwg.Wait()
		for i := 0; i < numRequestsMultiple; i++ {
			r := postresults[i]
			if r == nil {
				t.Errorf("expected a valid result, index %d", i)
				continue
			}

			if r.Status != http.StatusOK {
				t.Errorf("expected status code %v, got %v, index %d", http.StatusOK, r.Status, i)
			}

			if r.RouterVersion != "3" {
				t.Errorf("expected router version 3, got %v, index %d", r.RouterVersion, i)
			}
		}

	})

}

func TestService_Run_Reroute_Staggered(t *testing.T) {
	svc, err := startServiceInstance()
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Shutdown()

	// Wait a little bit of time for server to finish starting up
	time.Sleep(time.Millisecond * 100)

	// Add some route and redirects
	err = addRoute(svc.AdminPort, "/route1", fmt.Sprintf("http://localhost:%d", svc.UpstreamPort))
	if err != nil {
		t.Fatal(err)
	}

	nrClient := getRouterClient(svc.RouterPort)

	t.Run("multiple redirects with staggered start", func(t *testing.T) {

		// Fire off some requests with a delay
		results := make([]*result, numRequestsStaggered)
		wg := &sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < numRequestsStaggered; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					r, err := nrClient.Send("/route1", &sendOptions{RouterPreDelay: delayStaggered})
					if err != nil {
						t.Errorf("expected no error, got %v", err)
					}
					results[i] = r
				}()
				time.Sleep(delayStaggered * time.Millisecond)
			}
		}()

		// Sleep until halfway through sleep time of  requests
		time.Sleep((delayStaggered * numRequestsStaggered / 2) * time.Millisecond)

		// Add another route
		err = addRoute(svc.AdminPort, "/route2", fmt.Sprintf("http://localhost:%d", svc.UpstreamPort))
		if err != nil {
			t.Fatal(err)
		}

		// Collect and test the results
		wg.Wait()

		routervVersions := make(map[string]int)
		for i := 0; i < numRequestsStaggered; i++ {
			r := results[i]
			if r == nil {
				t.Errorf("expected a valid r, index %d", i)
				continue
			}

			if r.Status != http.StatusOK {
				t.Errorf("expected status code %v, got %v, index %d", http.StatusOK, r.Status, i)
			}

			routervVersions[r.RouterVersion]++
		}

		for version, count := range routervVersions {
			v, _ := strconv.Atoi(version)
			if v < 2 || v > 3 {
				t.Errorf("expected router version 1 or 2, got %v", v)
			}
			if count < 0.4*numRequestsStaggered || count > 0.6*numRequestsStaggered {
				t.Errorf("expected between 40%% and 60%% for version %s, got %v", version, count)
			}
		}
	})
}

func TestService_RunBigConfig(t *testing.T) {
	svc, err := startServiceInstance()
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Shutdown()

	// Wait a little bit of time for server to finish starting up
	time.Sleep(time.Millisecond * 100)

	// Add some routes and redirects
	err = addRedirect(svc.AdminPort, "/redir1", "http://localhost/redirected1")
	if err != nil {
		t.Fatal(err)
	}
	err = addRoute(svc.AdminPort, "/route1", fmt.Sprintf("http://localhost:%d", svc.UpstreamPort))
	if err != nil {
		t.Fatal(err)
	}

	nrClient := getRouterClient(svc.RouterPort)

	redirs := make(map[string]string)

	t.Run("add lots of redirects", func(t *testing.T) {
		fmt.Printf("BigConfig num requests = %d\n", numConfigRedirects)

		redirects := make([]redirect, 0)
		for i := 0; i < numConfigRedirects; i++ {
			path := fmt.Sprintf("/redir%d", i)
			dest := fmt.Sprintf("http://localhost/redirected%d", i)
			redirects = append(redirects, redirect{path: path, dest: dest})
			redirs[path] = dest

		}
		err = addRedirects(svc.AdminPort, redirects)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Printf("BigConfig added requests = %d\n", numConfigRedirects)

		start := time.Now()
		for path, dest := range redirs {
			result, err := nrClient.Send(path, nil)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if result.Status != http.StatusTemporaryRedirect {
				t.Errorf("expected status code %v, got %v", http.StatusTemporaryRedirect, result.Status)
			}

			if result.Location != dest {
				t.Errorf("expected Location header \"%s\", got %v", dest, result.Location)
			}
		}

		end := time.Now()
		fmt.Printf("BigConfig duration for %d reqs = %d ms\n", numConfigRedirects, end.Sub(start).Milliseconds())
	})

}

func startServiceInstance() (*service.Service, error) {
	// FIXME Get some random ports (not ideal as they could be reused inbetween identifying and starting the server).

	rl, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}
	al, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}
	ul, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}

	svc := &service.Service{
		RouterPort:   rl.Addr().(*net.TCPAddr).Port,
		UpstreamPort: ul.Addr().(*net.TCPAddr).Port,
		AdminPort:    al.Addr().(*net.TCPAddr).Port,
	}
	err = rl.Close()
	if err != nil {
		return nil, err
	}
	err = al.Close()
	if err != nil {
		return nil, err
	}
	err = ul.Close()
	if err != nil {
		return nil, err
	}

	if err := svc.Run(); err != nil {
		return nil, err
	}
	return svc, nil
}

type httpClient struct {
	routerPort int
	client     http.Client
}

func getRouterClient(port int) *httpClient {
	// HTTP client that doesn't follow redirects
	return &httpClient{port, http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}}
}

type sendOptions struct {
	UpstreamDelay  int
	RouterPreDelay int
}
type result struct {
	Status        int
	RouterVersion string
	Location      string
}

func (c *httpClient) Send(path string, opts *sendOptions) (*result, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d%s", c.routerPort, path), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("expected no error, got %v", err)
	}
	if opts != nil {
		if opts.UpstreamDelay > 0 {
			req.Header.Set("x-upstream-delay", strconv.Itoa(opts.UpstreamDelay))
		}
		if opts.RouterPreDelay > 0 {
			req.Header.Set("x-router-pre-delay", strconv.Itoa(opts.RouterPreDelay))
		}
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("expected no error, got %v", err)
	}

	result := &result{
		Status:        resp.StatusCode,
		Location:      resp.Header.Get("Location"),
		RouterVersion: resp.Header.Get("X-Router-Version"),
	}

	return result, nil
}

func addRedirect(port int, path, dest string) error {
	return addRedirects(port, []redirect{{path: path, dest: dest}})
}

type redirect struct {
	path string
	dest string
}

func addRedirects(port int, rs []redirect) error {
	items := []string{}
	for _, r := range rs {
		item := fmt.Sprintf(`{"path":"%s","redirect":"%s","type":"temp"}`, r.path, r.dest)
		items = append(items, item)
	}
	payload := "[" + strings.Join(items, ",") + "]"
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/redirects", port), `application/json`, strings.NewReader(payload))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return nil
}

func addRoute(port int, path, host string) error {
	return addRoutes(port, []route{{path: path, host: host}})
}

type route struct {
	path string
	host string
}

func addRoutes(port int, rs []route) error {
	items := []string{}
	for _, r := range rs {
		item := fmt.Sprintf(`{"path":"%s","host":"%s"}`, r.path, r.host)
		items = append(items, item)
	}
	payload := "[" + strings.Join(items, ",") + "]"
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/routes", port), `application/json`, strings.NewReader(payload))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return nil
}
