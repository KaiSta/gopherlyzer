// Package oracler answers questions about channel operations in Go source
// files.
//
// Earlier versions of gopherlyzer answered these questions by shelling out to
// `guru peers`. That query was removed from golang.org/x/tools (Go issue
// #59676) and the last release that still implemented it no longer builds with
// a current toolchain, so the query is answered in-process on top of
// go/packages and go/types; see resolve.go.
package oracler

import (
	"strconv"
	"strings"
)

// OracleAPI resolves channel operations of a single source file.
type OracleAPI struct {
	FilePath string
}

// PeersMode mirrors the JSON envelope of guru's peers query.
type PeersMode struct {
	Mode        string `json:"mode"`
	PeerEntries Peers  `json:"peers"`
}

// Peers mirrors the JSON payload of guru's peers query. Only Allocs is
// populated: the make(chan ...) call sites the queried channel may originate
// from. Sends and Receives are kept so the shape stays familiar.
type Peers struct {
	Pos      string   `json:"pos"`
	Type     string   `json:"type"`
	Allocs   []string `json:"allocs"`
	Sends    []string `json:"sends"`
	Receives []string `json:"receives"`
}

// New returns an oracle for the given source file.
func New(filePath string) *OracleAPI {
	return &OracleAPI{FilePath: filePath}
}

// GetPeers resolves the channel operation at byte offset charpos to the
// make(chan ...) sites that created its channel. Entries in Allocs are
// formatted "file:line:column"; parse them with LineNColumn.
//
// If the operation or its channel cannot be resolved, Allocs is empty; the
// caller treats that as "no channel operation here".
func (o *OracleAPI) GetPeers(charpos int) PeersMode {
	peers := Peers{Pos: strconv.Itoa(charpos)}

	idx, err := indexFor(o.FilePath)
	if err != nil {
		// Already reported once per file by indexFor.
		return PeersMode{Mode: "peers", PeerEntries: peers}
	}

	operand, ok := idx.channelOperandAt(o.FilePath, charpos)
	if !ok {
		return PeersMode{Mode: "peers", PeerEntries: peers}
	}

	peers.Allocs = idx.resolve(operand)
	return PeersMode{Mode: "peers", PeerEntries: peers}
}

// LineNColumn extracts the line and column from a "file:line:column" position.
// Everything before the final two colon-separated fields is ignored, so
// absolute paths and Windows drive letters are handled.
func (o *OracleAPI) LineNColumn(text string) (line, column int64) {
	parts := strings.Split(text, ":")
	if len(parts) < 2 {
		return 0, 0
	}
	line, _ = strconv.ParseInt(parts[len(parts)-2], 10, 32)
	column, _ = strconv.ParseInt(parts[len(parts)-1], 10, 32)
	return line, column
}
