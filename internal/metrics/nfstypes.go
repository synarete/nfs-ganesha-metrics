// SPDX-License-Identifier: Apache-2.0

package metrics

import "golang.org/x/sys/unix"

// Client Structure of the output of ShowClients dbus call
type Client struct {
	Client   string
	NFSv3    bool
	MNTv3    bool
	NLMv4    bool
	RQUOTA   bool
	NFSv40   bool
	NFSv41   bool
	NFSv42   bool
	Plan9    bool
	LastTime unix.Timespec
}

// Export Structure of the output of ShowExports dbus call
type Export struct {
	ExportID uint32
	Path     string
	NFSv3    bool
	MNTv3    bool
	NLMv4    bool
	RQUOTA   bool
	NFSv40   bool
	NFSv41   bool
	NFSv42   bool
	Plan9    bool
	LastTime unix.Timespec
}

// OperationStats
type OperationStats struct {
	Total  uint64
	Errors uint64
}

// ReplyHeader
type ReplyHeader struct {
	Status bool
	Error  string
	Time   unix.Timespec
}

// OperationCount
type OperationCount struct {
	NFSv3  uint64
	MNTv1  uint64
	MNTv3  uint64
	NLMv4  uint64
	RQUOTA uint64
	NFSv40 uint64
	NFSv41 uint64
	NFSv42 uint64
	Plan9  uint64
}

// OperationsStats
type OperationsStats struct {
	ReplyHeader
	OPS OperationCount
}

// IOCounts
type IOCounts struct {
	Total       uint64
	Errors      uint64
	Transferred uint64
}

// IOStats
type IOStats struct {
	Read   IOCounts
	Write  IOCounts
	Other  IOCounts
	Layout IOCounts
}

// ClientIOStats
type ClientIOStats struct {
	NFSv3  IOStats
	NFSv40 IOStats
	NFSv41 IOStats
	NFSv42 IOStats
}

// ClientIOs
type ClientIOs struct {
	ReplyHeader
	ClientIOStats
}
