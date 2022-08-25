// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	collectorsNamespace = "nfs_ganesha"
)

func (nme *nfsgMetricsExporter) register() error {
	cols := []prometheus.Collector{
		nme.newNfsgVersionsCollector(),
		nme.newNfsgExportsCollector(),
		nme.newNfsgClientsCollector(),
	}
	for _, c := range cols {
		if err := nme.reg.Register(c); err != nil {
			nme.log.Error(err, "failed to register collector")
			return err
		}
	}
	return nil
}

func collectorName(subsystem, name string) string {
	return prometheus.BuildFQName(collectorsNamespace, subsystem, name)
}

// nfsgCollector is common base type for all collectors
type nfsgCollector struct {
	// nolint:structcheck
	nme *nfsgMetricsExporter
	dsc []*prometheus.Desc
}

func (col *nfsgCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range col.dsc {
		ch <- d
	}
}

// nfsgVersionsCollector exports various versions informations
type nfsgVersionsCollector struct {
	nfsgCollector
	clnt *kclient
}

func (col *nfsgVersionsCollector) Collect(ch chan<- prometheus.Metric) {
	status := 0
	vers := GetVersions()
	if vers.Version != "" {
		status = 1
	}
	ch <- prometheus.MustNewConstMetric(
		col.dsc[0],
		prometheus.GaugeValue,
		float64(status),
		vers.Version,
		vers.CommitID,
	)
}

func (nme *nfsgMetricsExporter) newNfsgVersionsCollector() prometheus.Collector {
	col := &nfsgVersionsCollector{}
	col.nme = nme
	col.clnt, _ = newKClient()
	col.dsc = []*prometheus.Desc{
		prometheus.NewDesc(
			collectorName("metrics", "status"),
			"Current metrics-collector status and versions",
			[]string{
				"version",
				"commitid",
			}, nil),
	}
	return col
}

// nfsgExportsCollector extern information on NFS-Ganesha exports as
// Premetheus metrics
type nfsgExportsCollector struct {
	nfsgCollector
}

func (col *nfsgExportsCollector) Collect(ch chan<- prometheus.Metric) {
	reader := NewExportsDbusReader()
	if err := reader.Setup(); err != nil {
		col.nme.log.Error(err, "Collect exports stats")
		return
	}
	defer reader.Close()

	_, exports, err := reader.GetExports()
	if err != nil {
		col.nme.log.Error(err, "GetExports")
		return
	}
	ch <- prometheus.MustNewConstMetric(
		col.dsc[0],
		prometheus.GaugeValue,
		float64(len(exports)))

	for _, export := range exports {
		exportID := uint16(export.ExportID)
		stats, ok, err := reader.GetTotalOPS(exportID)
		if err != nil || !ok {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			col.dsc[1],
			prometheus.GaugeValue,
			float64(stats.OPS.NFSv3),
			strconv.Itoa(int(exportID)),
			export.Path)
		ch <- prometheus.MustNewConstMetric(
			col.dsc[2],
			prometheus.GaugeValue,
			float64(stats.OPS.NFSv40),
			strconv.Itoa(int(exportID)),
			export.Path)
		ch <- prometheus.MustNewConstMetric(
			col.dsc[3],
			prometheus.GaugeValue,
			float64(stats.OPS.NFSv41),
			strconv.Itoa(int(exportID)),
			export.Path)
		ch <- prometheus.MustNewConstMetric(
			col.dsc[4],
			prometheus.GaugeValue,
			float64(stats.OPS.NFSv42),
			strconv.Itoa(int(exportID)),
			export.Path)
	}
}

func (nme *nfsgMetricsExporter) newNfsgExportsCollector() prometheus.Collector {
	col := &nfsgExportsCollector{}
	col.nme = nme
	col.dsc = []*prometheus.Desc{
		prometheus.NewDesc(
			collectorName("export", "count"),
			"Total number of NFS exports", []string{}, nil),
		prometheus.NewDesc(
			collectorName("export", "ops_nfsv3"),
			"NFSv3 operations",
			[]string{"exportid", "path"}, nil),
		prometheus.NewDesc(
			collectorName("export", "ops_nfsv40"),
			"NFSv4.0 operations",
			[]string{"exportid", "path"}, nil),
		prometheus.NewDesc(
			collectorName("export", "ops_nfsv41"),
			"NFSv4.1 operations",
			[]string{"exportid", "path"}, nil),
		prometheus.NewDesc(
			collectorName("export", "ops_nfsv42"),
			"NFSv4.2 operations",
			[]string{"exportid", "path"}, nil),
	}
	return col
}

// nfsgClientsCollector extern information on NFS-Ganesha clients as
// Premetheus metrics
type nfsgClientsCollector struct {
	nfsgCollector
}

func (col *nfsgClientsCollector) Collect(ch chan<- prometheus.Metric) {
	reader := NewClientsDbusReader()
	if err := reader.Setup(); err != nil {
		col.nme.log.Error(err, "Collect clients stats")
		return
	}
	defer reader.Close()

	_, clients, err := reader.GetClients()
	if err != nil {
		col.nme.log.Error(err, "GetClients")
		return
	}
	ch <- prometheus.MustNewConstMetric(
		col.dsc[0],
		prometheus.GaugeValue,
		float64(len(clients)))

	for _, client := range clients {
		ipaddr := client.Client
		ios, ok, err := reader.GetClientIOs(ipaddr)
		if err != nil || !ok {
			continue
		}
		if client.NFSv3 {
			ch <- prometheus.MustNewConstMetric(
				col.dsc[1], prometheus.GaugeValue,
				float64(ios.NFSv3.Read.Total), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[2], prometheus.GaugeValue,
				float64(ios.NFSv3.Read.Errors), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[3], prometheus.GaugeValue,
				float64(ios.NFSv3.Read.Transferred), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[4], prometheus.GaugeValue,
				float64(ios.NFSv3.Write.Total), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[5], prometheus.GaugeValue,
				float64(ios.NFSv3.Write.Errors), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[6], prometheus.GaugeValue,
				float64(ios.NFSv3.Write.Transferred), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[7], prometheus.GaugeValue,
				float64(ios.NFSv3.Other.Total), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[8], prometheus.GaugeValue,
				float64(ios.NFSv3.Other.Errors), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[9], prometheus.GaugeValue,
				float64(ios.NFSv3.Other.Transferred), ipaddr)
		}

		if client.NFSv40 {
			ch <- prometheus.MustNewConstMetric(
				col.dsc[10], prometheus.GaugeValue,
				float64(ios.NFSv40.Read.Total), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[11], prometheus.GaugeValue,
				float64(ios.NFSv40.Read.Errors), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[12], prometheus.GaugeValue,
				float64(ios.NFSv40.Read.Transferred), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[13], prometheus.GaugeValue,
				float64(ios.NFSv40.Write.Total), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[14], prometheus.GaugeValue,
				float64(ios.NFSv40.Write.Errors), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[15], prometheus.GaugeValue,
				float64(ios.NFSv40.Write.Transferred), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[16], prometheus.GaugeValue,
				float64(ios.NFSv40.Other.Total), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[17], prometheus.GaugeValue,
				float64(ios.NFSv40.Other.Errors), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[18], prometheus.GaugeValue,
				float64(ios.NFSv40.Other.Transferred), ipaddr)
		}

		if client.NFSv41 {
			ch <- prometheus.MustNewConstMetric(
				col.dsc[19], prometheus.GaugeValue,
				float64(ios.NFSv41.Read.Total), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[20], prometheus.GaugeValue,
				float64(ios.NFSv41.Read.Errors), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[21], prometheus.GaugeValue,
				float64(ios.NFSv41.Read.Transferred), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[22], prometheus.GaugeValue,
				float64(ios.NFSv41.Write.Total), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[23], prometheus.GaugeValue,
				float64(ios.NFSv41.Write.Errors), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[24], prometheus.GaugeValue,
				float64(ios.NFSv41.Write.Transferred), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[25], prometheus.GaugeValue,
				float64(ios.NFSv41.Other.Total), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[26], prometheus.GaugeValue,
				float64(ios.NFSv41.Other.Errors), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[27], prometheus.GaugeValue,
				float64(ios.NFSv41.Other.Transferred), ipaddr)
		}

		if client.NFSv42 {
			ch <- prometheus.MustNewConstMetric(
				col.dsc[28], prometheus.GaugeValue,
				float64(ios.NFSv42.Read.Total), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[29], prometheus.GaugeValue,
				float64(ios.NFSv42.Read.Errors), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[30], prometheus.GaugeValue,
				float64(ios.NFSv42.Read.Transferred), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[31], prometheus.GaugeValue,
				float64(ios.NFSv42.Write.Total), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[32], prometheus.GaugeValue,
				float64(ios.NFSv42.Write.Errors), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[33], prometheus.GaugeValue,
				float64(ios.NFSv42.Write.Transferred), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[34], prometheus.GaugeValue,
				float64(ios.NFSv42.Other.Total), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[35], prometheus.GaugeValue,
				float64(ios.NFSv42.Other.Errors), ipaddr)

			ch <- prometheus.MustNewConstMetric(
				col.dsc[36], prometheus.GaugeValue,
				float64(ios.NFSv42.Other.Transferred), ipaddr)
		}
	}
}

func (nme *nfsgMetricsExporter) newNfsgClientsCollector() prometheus.Collector {
	col := &nfsgClientsCollector{}
	col.nme = nme
	col.dsc = []*prometheus.Desc{
		prometheus.NewDesc(
			collectorName("client", "count"),
			"Total number of NFS clients", []string{}, nil),

		// NFSv3
		prometheus.NewDesc(
			collectorName("client", "nfsv3_read_total"),
			"NFSv3 READ total", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv3_read_errors"),
			"NFSv3 READ errors", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv3_read_transferred"),
			"NFSv3 READ transferred", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv3_write_total"),
			"NFSv3 WRITE total", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv3_write_errors"),
			"NFSv3 WRITE errors", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv3_write_transferred"),
			"NFSv3 WRITE transferred", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv3_other_total"),
			"NFSv3 OTHER total", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv3_other_errors"),
			"NFSv3 OTHER errors", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv3_other_transferred"),
			"NFSv3 OTHER transferred", []string{"ipaddr"}, nil),

		// NFSv40
		prometheus.NewDesc(
			collectorName("client", "nfsv40_read_total"),
			"NFSv40 READ total", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv40_read_errors"),
			"NFSv40 READ errors", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv40_read_transferred"),
			"NFSv40 READ transferred", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv40_write_total"),
			"NFSv40 WRITE total", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv40_write_errors"),
			"NFSv40 WRITE errors", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv40_write_transferred"),
			"NFSv40 WRITE transferred", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv40_other_total"),
			"NFSv40 OTHER total", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv40_other_errors"),
			"NFSv40 OTHER errors", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv40_other_transferred"),
			"NFSv40 OTHER transferred", []string{"ipaddr"}, nil),

		// NFSv41
		prometheus.NewDesc(
			collectorName("client", "nfsv41_read_total"),
			"NFSv41 READ total", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv41_read_errors"),
			"NFSv41 READ errors", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv41_read_transferred"),
			"NFSv41 READ transferred", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv41_write_total"),
			"NFSv41 WRITE total", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv41_write_errors"),
			"NFSv41 WRITE errors", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv41_write_transferred"),
			"NFSv41 WRITE transferred", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv41_other_total"),
			"NFSv41 OTHER total", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv41_other_errors"),
			"NFSv41 OTHER errors", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv41_other_transferred"),
			"NFSv41 OTHER transferred", []string{"ipaddr"}, nil),

		// NFSv42
		prometheus.NewDesc(
			collectorName("client", "nfsv42_read_total"),
			"NFSv42 READ total", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv42_read_errors"),
			"NFSv42 READ errors", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv42_read_transferred"),
			"NFSv42 READ transferred", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv42_write_total"),
			"NFSv42 WRITE total", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv42_write_errors"),
			"NFSv42 WRITE errors", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv42_write_transferred"),
			"NFSv42 WRITE transferred", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv42_other_total"),
			"NFSv42 OTHER total", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv42_other_errors"),
			"NFSv42 OTHER errors", []string{"ipaddr"}, nil),
		prometheus.NewDesc(
			collectorName("client", "nfsv42_other_transferred"),
			"NFSv42 OTHER transferred", []string{"ipaddr"}, nil),
	}
	return col
}
