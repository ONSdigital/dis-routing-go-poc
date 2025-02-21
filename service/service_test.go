package service_test

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dis-routing-go-poc/service"
)

func TestService_Run(t *testing.T) {
	svc, err := startServiceInstance()
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Shutdown()

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

	// Add some route and redirects
	err = addRoute(svc.AdminPort, "/route1", fmt.Sprintf("http://localhost:%d", svc.UpstreamPort))
	if err != nil {
		t.Fatal(err)
	}

	nrClient := getRouterClient(svc.RouterPort)

	const numRequests = 100
	const delay = 100

	t.Run("multiple_redirects", func(t *testing.T) {

		// Fire off some requests with a delay
		preresults := make([]*result, numRequests)
		prewg := &sync.WaitGroup{}
		for i := 0; i < numRequests; i++ {
			prewg.Add(1)
			go func() {
				defer prewg.Done()
				r, err := nrClient.Send("/route1", &sendOptions{UpstreamDelay: delay})
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				preresults[i] = r
			}()
		}

		// Sleep until half way through sleep time of pre requests
		time.Sleep((delay / 2) * time.Millisecond)

		// Add another route
		err = addRoute(svc.AdminPort, "/route2", fmt.Sprintf("http://localhost:%d", svc.UpstreamPort))
		if err != nil {
			t.Fatal(err)
		}

		// Fire off some more requests with a delay
		postresults := make([]*result, numRequests)
		postwg := &sync.WaitGroup{}
		for i := 0; i < numRequests; i++ {
			postwg.Add(1)
			go func() {
				defer postwg.Done()
				r, err := nrClient.Send("/route1", &sendOptions{UpstreamDelay: delay})
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				postresults[i] = r
			}()
		}

		// Collect and test the results

		// Test pre requests
		prewg.Wait()
		for i := 0; i < numRequests; i++ {
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
		for i := 0; i < numRequests; i++ {
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

	// Add some route and redirects
	err = addRoute(svc.AdminPort, "/route1", fmt.Sprintf("http://localhost:%d", svc.UpstreamPort))
	if err != nil {
		t.Fatal(err)
	}

	nrClient := getRouterClient(svc.RouterPort)

	const numRequests = 100
	const delay = 10

	t.Run("multiple redirects with staggered start", func(t *testing.T) {

		// Fire off some requests with a delay
		results := make([]*result, numRequests)
		wg := &sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < numRequests; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					r, err := nrClient.Send("/route1", &sendOptions{RouterPreDelay: delay})
					if err != nil {
						t.Errorf("expected no error, got %v", err)
					}
					results[i] = r
				}()
				time.Sleep(delay * time.Millisecond)
			}
		}()

		// Sleep until halfway through sleep time of  requests
		time.Sleep((delay * numRequests / 2) * time.Millisecond)

		// Add another route
		err = addRoute(svc.AdminPort, "/route2", fmt.Sprintf("http://localhost:%d", svc.UpstreamPort))
		if err != nil {
			t.Fatal(err)
		}

		// Collect and test the results
		wg.Wait()

		routervVersions := make(map[string]int)
		for i := 0; i < numRequests; i++ {
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
			if count < 40 || count > 60 {
				t.Errorf("expected between 40 and 60 for version %s, got %v", version, count)
			}
		}
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
	payload := fmt.Sprintf(`{"path":"%s","redirect":"%s","type":"temp"}`, path, dest)
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
	payload := fmt.Sprintf(`{"path":"%s","host":"%s"}`, path, host)
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/routes", port), `application/json`, strings.NewReader(payload))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return nil
}
