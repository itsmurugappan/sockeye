package http

import (
	"fmt"
	"net"
	nethttp "net/http"
	"net/url"
	"strings"
	"time"
)

// Option is the function signature required to be considered an http.Option.
type Option func(*Protocol) error

// WithTarget sets the outbound recipient of cloudevents when using an HTTP
// request.
func WithTarget(targetUrl string) Option {
	return func(p *Protocol) error {
		if p == nil {
			return fmt.Errorf("http protocol option can not set nil protocol")
		}
		targetUrl = strings.TrimSpace(targetUrl)
		if targetUrl != "" {
			var err error
			var target *url.URL
			target, err = url.Parse(targetUrl)
			if err != nil {
				return fmt.Errorf("http target option failed to parse target url: %s", err.Error())
			}

			p.Target = target

			if p.RequestTemplate == nil {
				p.RequestTemplate = &nethttp.Request{
					Method: nethttp.MethodPost,
				}
			}
			p.RequestTemplate.URL = target

			return nil
		}
		return fmt.Errorf("http target option was empty string")
	}
}

// WithHeader sets an additional default outbound header for all cloudevents
// when using an HTTP request.
func WithHeader(key, value string) Option {
	return func(p *Protocol) error {
		if p == nil {
			return fmt.Errorf("http header option can not set nil transport")
		}
		key = strings.TrimSpace(key)
		if key != "" {
			if p.RequestTemplate == nil {
				p.RequestTemplate = &nethttp.Request{
					Method: nethttp.MethodPost,
				}
			}
			if p.RequestTemplate.Header == nil {
				p.RequestTemplate.Header = nethttp.Header{}
			}
			p.RequestTemplate.Header.Add(key, value)
			return nil
		}
		return fmt.Errorf("http header option was empty string")
	}
}

// WithShutdownTimeout sets the shutdown timeout when the http server is being shutdown.
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(t *Protocol) error {
		if t == nil {
			return fmt.Errorf("http shutdown timeout option can not set nil transport")
		}
		t.ShutdownTimeout = &timeout
		return nil
	}
}

func checkListen(t *Protocol, prefix string) error {
	switch {
	case t.Port != nil:
		return fmt.Errorf("%v port already set", prefix)
	case t.listener != nil:
		return fmt.Errorf("%v listener already set", prefix)
	}
	return nil
}

// WithPort sets the listening port for StartReceiver.
// Only one of WithListener  or WithPort is allowed.
func WithPort(port int) Option {
	return func(t *Protocol) error {
		if t == nil {
			return fmt.Errorf("http port option can not set nil transport")
		}
		if port < 0 || port > 65535 {
			return fmt.Errorf("http port option was given an invalid port: %d", port)
		}
		if err := checkListen(t, "http port option"); err != nil {
			return err
		}
		t.setPort(port)
		return nil
	}
}

// WithListener sets the listener for StartReceiver.
// Only one of WithListener or WithPort is allowed.
func WithListener(l net.Listener) Option {
	return func(t *Protocol) error {
		if t == nil {
			return fmt.Errorf("http listener option can not set nil transport")
		}
		if err := checkListen(t, "http port option"); err != nil {
			return err
		}
		t.listener = l
		_, err := t.listen()
		return err
	}
}

// WithPath sets the path to receive cloudevents on for HTTP transports.
func WithPath(path string) Option {
	return func(t *Protocol) error {
		if t == nil {
			return fmt.Errorf("http path option can not set nil transport")
		}
		path = strings.TrimSpace(path)
		if len(path) == 0 {
			return fmt.Errorf("http path option was given an invalid path: %q", path)
		}
		t.Path = path
		return nil
	}
}

//
// Middleware is a function that takes an existing http.Handler and wraps it in middleware,
// returning the wrapped http.Handler.
type Middleware func(next nethttp.Handler) nethttp.Handler

// WithMiddleware adds an HTTP middleware to the transport. It may be specified multiple times.
// Middleware is applied to everything before it. For example
// `NewClient(WithMiddleware(foo), WithMiddleware(bar))` would result in `bar(foo(original))`.
func WithMiddleware(middleware Middleware) Option {
	return func(t *Protocol) error {
		if t == nil {
			return fmt.Errorf("http middleware option can not set nil transport")
		}
		t.middleware = append(t.middleware, middleware)
		return nil
	}
}

// WithHTTPTransport sets the HTTP client transport.
func WithHTTPTransport(httpTransport nethttp.RoundTripper) Option {
	return func(t *Protocol) error {
		t.transport = httpTransport
		return nil
	}
}
