// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"errors"
	"os"
	"reflect"
	"strconv"

	dbus "github.com/godbus/dbus/v5"
	"golang.org/x/sys/unix"
)

const (
	nfsGaneshaDbusServicePrefix     = "org.ganesha.nfsd"
	nfsGaneshaDbusExportMgrPrefix   = "org.ganesha.nfsd.exportmgr"
	nfsGaneshaDbusExportStatsPrefix = "org.ganesha.nfsd.exportstats"
	nfsGaneshaExportInterface       = "/org/ganesha/nfsd/ExportMgr"
	nfsGaneshaDbusClientMgrPrefix   = "org.ganesha.nfsd.clientmgr"
	nfsGaneshaDbusClientStatsPrefix = "org.ganesha.nfsd.clientstats"
	nfsGaneshaClientInterface       = "/org/ganesha/nfsd/ClientMgr"
)

// DbusReader
type DbusReader struct {
	dbusServicePrefix string
	dbusStatsPrefix   string
	dbusMgrPrefix     string
	dbusInterfacePath string
	dbusConn          *dbus.Conn
	dbusObject        dbus.BusObject
}

func (dr *DbusReader) Setup() error {
	conn, err := dbus.SystemBusPrivate()
	if err != nil {
		return err
	}

	methods := []dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}
	err = conn.Auth(methods)
	if err != nil {
		conn.Close()
		return err
	}

	err = conn.Hello()
	if err != nil {
		conn.Close()
		return err
	}

	dr.dbusConn = conn
	dr.dbusObject = conn.Object(
		dr.dbusServicePrefix,
		dbus.ObjectPath(dr.dbusInterfacePath),
	)
	return nil
}

func (dr *DbusReader) Close() {
	if dr.dbusConn != nil {
		dr.dbusConn.Close()
		dr.dbusConn = nil
	}
}

func (dr *DbusReader) makeDbusCall(method string) (*dbus.Call, error) {
	call := dr.dbusObject.Call(method, 0)
	err := call.Err
	if err != nil {
		return nil, err
	}
	return call, nil
}

func (dr *DbusReader) makeDbusCallWith(
	method string, args ...interface{}) (*dbus.Call, bool, error) {
	call := dr.dbusObject.Call(method, 0, args...)
	err := call.Err
	if err != nil {
		return nil, false, err
	}
	if len(call.Body) < 1 {
		return nil, false, errors.New("illegal reply: " + method)
	}
	status, ok := call.Body[0].(bool)
	if !ok {
		return nil, false, errors.New("illegal reply status: " + method)
	}
	_, ok = call.Body[1].(string)
	if !ok {
		return nil, false, errors.New("illegal reply errstr: " + method)
	}
	return call, status, nil
}

func (dr *DbusReader) statsMethod(name string) string {
	return dr.dbusStatsPrefix + "." + name
}

func (dr *DbusReader) mgrMethod(name string) string {
	return dr.dbusMgrPrefix + "." + name
}

// ExportsDbusReader
type ExportsDbusReader struct {
	DbusReader
}

// NewExportsDbusReader
func NewExportsDbusReader() *ExportsDbusReader {
	return &ExportsDbusReader{
		DbusReader{
			dbusServicePrefix: nfsGaneshaDbusServicePrefix,
			dbusStatsPrefix:   nfsGaneshaDbusExportStatsPrefix,
			dbusMgrPrefix:     nfsGaneshaDbusExportMgrPrefix,
			dbusInterfacePath: nfsGaneshaExportInterface,
		},
	}
}

func (exdr *ExportsDbusReader) GetExports() (
	unix.Timespec, []Export, error) {
	var exports []Export
	utime := unix.Timespec{}
	method := exdr.mgrMethod("ShowExports")
	call, err := exdr.makeDbusCall(method)
	if err != nil {
		return utime, exports, err
	}
	err = call.Store(&utime, &exports)
	if err != nil {
		return utime, exports, err
	}
	return utime, exports, nil
}

func (exdr *ExportsDbusReader) GetTotalOPS(
	exportID uint16) (*OperationsStats, bool, error) {
	call, status, err := exdr.makeExportStatsDbusCall("GetTotalOPS", exportID)
	if err != nil {
		return nil, status, err
	}
	out := OperationsStats{}
	if !status {
		err = call.Store(&out.Status, &out.Error)
		return &out, status, err
	}
	if len(call.Body) < 4 ||
		!isSlice(call.Body[2]) || !isSlice(call.Body[3]) {
		_ = call.Store(&out.Status, &out.Error)
		return &out, status, errors.New("protocol error")
	}
	out.Status, _ = call.Body[0].(bool)
	out.Error, _ = call.Body[1].(string)
	out.OPS = parseOPs(call.Body[3])
	return &out, true, nil
}

func (exdr *ExportsDbusReader) GetGlobalOPS(
	exportID uint16) (*OperationsStats, bool, error) {
	call, status, err := exdr.makeExportStatsDbusCall("GetGlobalOPS", exportID)
	if err != nil {
		return nil, status, err
	}
	out := OperationsStats{}
	if !status {
		err = call.Store(&out.Status, &out.Error)
		return &out, status, err
	}
	if len(call.Body) < 4 ||
		!isSlice(call.Body[2]) || !isSlice(call.Body[3]) {
		_ = call.Store(&out.Status, &out.Error)
		return &out, status, errors.New("protocol error")
	}
	out.Status, _ = call.Body[0].(bool)
	out.Error, _ = call.Body[1].(string)
	out.OPS = parseOPs(call.Body[3])
	return &out, true, nil
}

func parseOPs(v interface{}) OperationCount {
	ops := OperationCount{}
	dat := reflect.ValueOf(v)
	for i := 1; i < dat.Len(); i += 2 {
		key, hasKey := asString(dat.Index(i - 1))
		val, hasVal := asUint64(dat.Index(i))
		if hasKey && hasVal {
			switch {
			case key == "NFSv3":
				ops.NFSv3 = val
			case key == "NFSv40":
				ops.NFSv40 = val
			case key == "NFSv41":
				ops.NFSv41 = val
			case key == "NFSv42":
				ops.NFSv42 = val
			case key == "MNTv1":
				ops.MNTv1 = val
			case key == "MNTv3":
				ops.MNTv3 = val
			case key == "NLMv4":
				ops.NLMv4 = val
			case key == "RQUOTA":
				ops.RQUOTA = val
			case key == "Plan9":
				ops.Plan9 = val
			}
		}
	}
	return ops
}

func (exdr *ExportsDbusReader) makeExportStatsDbusCall(
	name string, exportID uint16) (*dbus.Call, bool, error) {
	method := exdr.statsMethod(name)
	return exdr.makeDbusCallWith(method, exportID)
}

// ClientsDbusReader
type ClientsDbusReader struct {
	DbusReader
}

// NewClientsDbusReader
func NewClientsDbusReader() *ClientsDbusReader {
	return &ClientsDbusReader{
		DbusReader{
			dbusServicePrefix: nfsGaneshaDbusServicePrefix,
			dbusStatsPrefix:   nfsGaneshaDbusClientStatsPrefix,
			dbusMgrPrefix:     nfsGaneshaDbusClientMgrPrefix,
			dbusInterfacePath: nfsGaneshaClientInterface,
		},
	}
}

func (cldr *ClientsDbusReader) GetClients() (unix.Timespec, []Client, error) {
	var clients []Client
	utime := unix.Timespec{}
	method := cldr.mgrMethod("ShowClients")
	call, err := cldr.makeDbusCall(method)
	if err != nil {
		return utime, clients, err
	}
	err = call.Store(&utime, &clients)
	if err != nil {
		return utime, clients, err
	}
	return utime, clients, nil
}

func (cldr *ClientsDbusReader) GetClientIOs(
	ipaddr string) (*ClientIOs, bool, error) {
	out := ClientIOs{}
	call, status, err := cldr.makeClientStatsDbusCall("GetClientIOops", ipaddr)
	if err != nil {
		return nil, false, err
	}
	if !status {
		err = call.Store(&out.Status, &out.Error)
		return &out, false, err
	}
	if len(call.Body) < 3 {
		_ = call.Store(&out.Status, &out.Error)
		return &out, status, errors.New("protocol error")
	}
	out.ClientIOStats = parseClientIOs(call.Body[3:])
	return &out, true, nil
}

func parseClientIOs(v []interface{}) ClientIOStats {
	ios := ClientIOStats{}
	len := len(v)
	idx := 0
	cnt := 0
	// NFSv3
	if idx >= len {
		return ios
	}
	val, hasVal := asBool(reflect.ValueOf(v[idx]))
	idx += 1
	if !hasVal {
		return ios
	}
	if val {
		ios.NFSv3, cnt = ParseIOStats(v[idx:])
		idx += cnt
	}
	// NFSv40
	if idx >= len {
		return ios
	}
	val, hasVal = asBool(reflect.ValueOf(v[idx]))
	idx += 1
	if !hasVal {
		return ios
	}
	if val {
		ios.NFSv40, cnt = ParseIOStats(v[idx:])
		idx += cnt
	}
	// NFSv41
	if idx >= len {
		return ios
	}
	val, hasVal = asBool(reflect.ValueOf(v[idx]))
	idx += 1
	if !hasVal {
		return ios
	}
	if val {
		ios.NFSv41, cnt = ParseIOStats(v[idx:])
		idx += cnt
	}
	// NFSv42
	if idx >= len {
		return ios
	}
	val, hasVal = asBool(reflect.ValueOf(v[idx]))
	idx += 1
	if !hasVal {
		return ios
	}
	if val {
		ios.NFSv42, cnt = ParseIOStats(v[idx:])
		idx += cnt
	}
	return ios
}

func ParseIOStats(v []interface{}) (IOStats, int) {
	cnt := 0
	cur := 0
	ret := IOStats{}
	len := len(v)
	if len >= 1 {
		ret.Read, cur = ParseIOCounts(v[0])
		if cur > 0 {
			cnt += 1
		}
	}
	if len >= 2 {
		ret.Write, cur = ParseIOCounts(v[1])
		if cur > 0 {
			cnt += 1
		}
	}
	if len >= 3 {
		ret.Other, cur = ParseIOCounts(v[2])
		if cur > 0 {
			cnt += 1
		}
	}
	if len >= 4 {
		ret.Layout, cur = ParseIOCounts(v[3])
		if cur > 0 {
			cnt += 1
		}
	}
	return ret, cnt
}

func ParseIOCounts(in interface{}) (IOCounts, int) {
	cnt := 0
	ret := IOCounts{}
	if !isSlice(in) {
		return ret, cnt
	}
	dat := reflect.ValueOf(in)
	v, ok := dat.Interface().([]interface{})
	if !ok {
		return ret, cnt
	}
	len := len(v)
	if len > 0 {
		ret.Total, _ = asUint64(dat.Index(0))
		cnt += 1
	}
	if len > 1 {
		ret.Errors, _ = asUint64(dat.Index(1))
		cnt += 1
	}
	if len > 2 {
		ret.Transferred, _ = asUint64(dat.Index(2))
		cnt += 1
	}
	return ret, cnt
}

func (cldr *ClientsDbusReader) makeClientStatsDbusCall(
	name string, ipaddr string) (*dbus.Call, bool, error) {
	method := cldr.statsMethod(name)
	return cldr.makeDbusCallWith(method, ipaddr)
}

func asString(v reflect.Value) (string, bool) {
	s, ok := v.Interface().(string)
	if !ok {
		return "", false
	}
	return s, true
}

func asUint64(v reflect.Value) (uint64, bool) {
	u, ok := v.Interface().(uint64)
	if !ok {
		return 0, false
	}
	return u, true
}

func asBool(v reflect.Value) (bool, bool) {
	b, ok := v.Interface().(bool)
	if !ok {
		return false, false
	}
	return b, true
}

func isSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}
