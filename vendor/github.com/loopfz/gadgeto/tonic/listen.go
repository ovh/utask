package tonic

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var defaultOpts = []ListenOptFunc{
	ListenAddr(":8080"),
	CatchSignals(os.Interrupt, syscall.SIGTERM),
	ShutdownTimeout(10 * time.Second),
	ReadHeaderTimeout(5 * time.Second),
	WriteTimeout(30 * time.Second),
	KeepAliveTimeout(90 * time.Second),
}

func ListenAndServe(handler http.Handler, errorHandler func(error), opt ...ListenOptFunc) {

	srv := &http.Server{Handler: handler}

	listenOpt := &ListenOpt{Server: srv}

	for _, o := range defaultOpts {
		err := o(listenOpt)
		if err != nil {
			if errorHandler != nil {
				errorHandler(err)
			}
			return
		}
	}

	for _, o := range opt {
		err := o(listenOpt)
		if err != nil {
			if errorHandler != nil {
				errorHandler(err)
			}
			return
		}
	}

	stop := make(chan struct{})

	go func() {
		var err error
		if srv.TLSConfig != nil && len(srv.TLSConfig.Certificates) > 0 {
			// ListenAndServeTLS without cert files lets listenOpts set srv.TLSConfig.Certificates
			err = srv.ListenAndServeTLS("", "")
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed && errorHandler != nil {
			errorHandler(err)
		}
		close(stop)
	}()

	sig := make(chan os.Signal)

	if len(listenOpt.Signals) > 0 {
		signal.Notify(sig, listenOpt.Signals...)
	}

	select {
	case <-sig:
		ctx, cancel := context.WithTimeout(context.Background(), listenOpt.ShutdownTimeout)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil && errorHandler != nil {
			errorHandler(err)
		}

	case <-stop:
		break

	}
}

type ListenOpt struct {
	Server          *http.Server
	Signals         []os.Signal
	ShutdownTimeout time.Duration
}

type ListenOptFunc func(*ListenOpt) error

func CatchSignals(sig ...os.Signal) ListenOptFunc {
	return func(opt *ListenOpt) error {
		opt.Signals = sig
		return nil
	}
}

func ListenAddr(addr string) ListenOptFunc {
	return func(opt *ListenOpt) error {
		opt.Server.Addr = addr
		return nil
	}
}

func ReadTimeout(t time.Duration) ListenOptFunc {
	return func(opt *ListenOpt) error {
		opt.Server.ReadTimeout = t
		return nil
	}
}

func ReadHeaderTimeout(t time.Duration) ListenOptFunc {
	return func(opt *ListenOpt) error {
		opt.Server.ReadHeaderTimeout = t
		return nil
	}
}

func WriteTimeout(t time.Duration) ListenOptFunc {
	return func(opt *ListenOpt) error {
		opt.Server.WriteTimeout = t
		return nil
	}
}

func KeepAliveTimeout(t time.Duration) ListenOptFunc {
	return func(opt *ListenOpt) error {
		opt.Server.IdleTimeout = t
		return nil
	}
}

func ShutdownTimeout(t time.Duration) ListenOptFunc {
	return func(opt *ListenOpt) error {
		opt.ShutdownTimeout = t
		return nil
	}
}
