package qlog

import (
	"io"
	"time"

	"github.com/francoispqt/gojay"
	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/wire"
)

// A Tracer records events to be exported to a qlog.
type Tracer interface {
	Export(io.Writer) error
	SentPacket(time.Time, *wire.ExtendedHeader, []wire.Frame)
	ReceivedPacket(time.Time, *wire.ExtendedHeader, []wire.Frame)
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

func (t *tracer) SentPacket(time time.Time, hdr *wire.ExtendedHeader, frames []wire.Frame) {
	fs := make([]frame, len(frames))
	for i, f := range frames {
		fs[i] = *transformFrame(f)
	}
	t.events = append(t.events, event{
		Time: time,
		eventDetails: eventPacketSent{
			PacketType: getPacketType(hdr),
			Header:     *transformHeader(hdr),
			Frames:     fs,
		},
	})
}

func (t *tracer) ReceivedPacket(time time.Time, hdr *wire.ExtendedHeader, frames []wire.Frame) {
	fs := make([]frame, len(frames))
	for i, f := range frames {
		fs[i] = *transformFrame(f)
	}
	t.events = append(t.events, event{
		Time: time,
		eventDetails: eventPacketReceived{
			PacketType: getPacketType(hdr),
			Header:     *transformHeader(hdr),
			Frames:     fs,
		},
	})
}
