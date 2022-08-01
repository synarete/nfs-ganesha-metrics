// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"fmt"
	"net"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// DefaultMetricsPort is the default port used to export prometheus metrics
	DefaultMetricsPort = int(8080)
	// DefaultMetricsPath is the default HTTP path to export prometheus metrics
	DefaultMetricsPath = "/metrics"
)

type nfsgMetricsExporter struct {
	log  logr.Logger
	reg  *prometheus.Registry
	mux  *http.ServeMux
	port int
}

func newNfsgMetricsExporter(log logr.Logger, port int) *nfsgMetricsExporter {
	return &nfsgMetricsExporter{
		log:  log,
		reg:  prometheus.NewRegistry(),
		mux:  http.NewServeMux(),
		port: port,
	}
}

func (nme *nfsgMetricsExporter) init() error {
	nme.log.Info("register collectors")
	return nme.register()
}

func (nme *nfsgMetricsExporter) serve() error {
	addr := fmt.Sprintf(":%d", nme.port)
	nme.log.Info("serve metrics", "addr", addr)

	handler := promhttp.HandlerFor(nme.reg, promhttp.HandlerOpts{})
	nme.mux.Handle(DefaultMetricsPath, handler)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		nme.log.Error(err, "failed to listen", "addr", addr)
		return err
	}
	defer listener.Close()

	if err := http.Serve(listener, nme.mux); err != nil {
		nme.log.Error(err, "HTTP server failure", "addr", addr)
		return err
	}
	return nil
}

// RunNfsgMetricsExporter executes an HTTP server and exports NFS-Ganesha
// stats as Prometheus metrics.
func RunNfsgMetricsExporter(log logr.Logger) error {
	nme := newNfsgMetricsExporter(log, DefaultMetricsPort)
	err := nme.init()
	if err != nil {
		return err
	}
	return nme.serve()
}
