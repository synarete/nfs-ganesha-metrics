// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	goruntime "runtime"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/go-logr/logr"
	"github.com/synarete/nfs-ganesha-metrics/internal/metrics"
)

var (
	// Version of the software at compile time.
	Version = "(unset)"
	// CommitID of the git revision used to compile the software.
	CommitID = "(unset)"
)

func init() {
	metrics.UpdateDefaultVersions(Version, CommitID)
}

func main() {
	log := zap.New(zap.UseDevMode(true))

	// Report global info early upon boot
	start(log)

	// Do minimal check for DBus readers
	poke(log)

	// Execute metrics server
	exec(log)
}

func start(log logr.Logger) {
	log.Info("Initializing nfsganeshametrics",
		"ProgramName", os.Args[0],
		"GoVersion", goruntime.Version(),
		"Version", Version,
		"CommitID", CommitID)

	log.Info("Self", "PodID", metrics.GetSelfPodID())

	log.Info("IDs", "uid", os.Getuid(), "gid", os.Getgid())

	log.Info("Env",
		"KUBECONFIG",
		os.Getenv("KUBECONFIG"),
		"DBUS_SESSION_BUS_ADDRESS",
		os.Getenv("DBUS_SESSION_BUS_ADDRESS"),
	)
}

func poke(log logr.Logger) {
	exportsReader := metrics.NewExportsDbusReader()
	if err := exportsReader.Setup(); err != nil {
		log.Error(err, "ExportsDbusReader.Setup")
		return
	}
	defer exportsReader.Close()

	clientsReader := metrics.NewClientsDbusReader()
	if err := exportsReader.Setup(); err != nil {
		log.Error(err, "ClientsDbusReader.Setup")
		return
	}
	defer clientsReader.Close()
}

func exec(log logr.Logger) {
	err := metrics.RunNfsgMetricsExporter(log)
	if err != nil {
		log.Error(err, "RunNfsgMetricsExporter")
		os.Exit(1)
	}
}
