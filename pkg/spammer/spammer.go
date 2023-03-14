package spammer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/projectdiscovery/ratelimit"
	"github.com/sirupsen/logrus"
)

type SpamOptions struct {
	Paths   []string
	Hosts   []string
	Threads uint
	Rate    uint
}

func Spam(ctx context.Context, opts *SpamOptions) error {
	internalCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	rl := ratelimit.New(internalCtx, opts.Rate, time.Second)
	wg := &sync.WaitGroup{}
	inputs := make(chan string, opts.Threads*2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(inputs)
		nextPath := 0
		nextHost := 0
		for {
			select {
			case <-internalCtx.Done():
				return
			default:
				if nextHost >= len(opts.Hosts) {
					nextPath++
					nextHost = 0
				}
				if nextPath >= len(opts.Paths) {
					logrus.Debug("finished generating inputs")
					return
				}
				path := opts.Paths[nextPath]
				host := opts.Hosts[nextHost]
				nextHost++
				u, err := normalizeHostAndAddPath(host, path)
				if err != nil {
					logrus.Debugf("error normalizing host and path: %s", err)
					continue
				}
				inputs <- u
			}
		}
	}()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	for i := 0; i < int(opts.Threads); i++ {
		wg.Add(1)
		logrus.Debugf("starting thread %d", i+1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-internalCtx.Done():
					logrus.Debug("thread finished")
					return
				case u, ok := <-inputs:
					if !ok {
						return
					}
					rl.Take()
					logrus.Debugf("sending request to %q", u)
					res, err := client.Get(u)
					if err != nil {
						logrus.Errorf("error sending request: %s", u)
						continue
					}
					if res.StatusCode != http.StatusNotFound && res.StatusCode < 500 {
						fmt.Printf("%s [%d]\n", u, res.StatusCode)
					} else {
						logrus.Debugf("%s [%d]", u, res.StatusCode)
					}
				}
			}
		}()
	}

	wg.Wait()
	logrus.Debug("all threads finished")
	return nil
}

func normalizeHostAndAddPath(host string, path string) (string, error) {
	u, err := url.Parse(host)
	if err != nil {
		return "", err
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	}

	new := u.JoinPath(path)
	return new.String(), nil
}
