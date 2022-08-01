// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	goruntime "runtime"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

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
	log.Info("Initializing nfsganeshametrics",
		"ProgramName", os.Args[0],
		"GoVersion", goruntime.Version(),
		"Version", Version,
		"CommitID", CommitID)

	log.Info("Self", "PodID", metrics.GetSelfPodID())

	err := metrics.RunNfsgMetricsExporter(log)
	if err != nil {
		os.Exit(1)
	}
}
