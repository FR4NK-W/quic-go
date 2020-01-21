package qlog

import (
	"io"

	"github.com/francoispqt/gojay"
	"github.com/lucas-clemente/quic-go/internal/protocol"
)

// A Tracer records events to be exported to a qlog.
type Tracer interface {
	Export(io.Writer) error
}

type tracer struct {
	odcid       protocol.ConnectionID
	perspective protocol.Perspective

	events []event
}

var _ Tracer = &tracer{}

// NewTracer creates a new tracer to record a qlog.
func NewTracer(p protocol.Perspective, odcid protocol.ConnectionID) Tracer {
	return &tracer{
		perspective: p,
		odcid:       odcid,
	}
}

// Export writes a qlog.
func (t *tracer) Export(w io.Writer) error {
	enc := gojay.NewEncoder(w)
	enc.Encode(&topLevel{
		traces: traces{
			{
				VantagePoint: vantagePoint{Type: t.perspective},
				CommonFields: commonFields{ODCID: connectionID(t.odcid), GroupID: connectionID(t.odcid)},
				EventFields:  eventFields[:],
				Events:       t.events,
			},
		}})
	return nil
}
