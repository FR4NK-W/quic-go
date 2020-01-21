package qlog

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/wire"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tracer", func() {
	var tracer Tracer

	BeforeEach(func() {
		tracer = NewTracer(protocol.PerspectiveServer, protocol.ConnectionID{0xde, 0xad, 0xbe, 0xef})
	})

	It("exports a trace that has the right metadata", func() {
		buf := &bytes.Buffer{}
		Expect(tracer.Export(buf)).To(Succeed())

		m := make(map[string]interface{})
		Expect(json.Unmarshal(buf.Bytes(), &m)).To(Succeed())
		Expect(m).To(HaveKeyWithValue("qlog_version", "draft-02-wip"))
		Expect(m).To(HaveKey("title"))
		Expect(m).To(HaveKey("traces"))
		traces := m["traces"].([]interface{})
		Expect(traces).To(HaveLen(1))
		trace := traces[0].(map[string]interface{})
		Expect(trace).To(HaveKey(("common_fields")))
		commonFields := trace["common_fields"].(map[string]interface{})
		Expect(commonFields).To(HaveKeyWithValue("ODCID", "deadbeef"))
		Expect(commonFields).To(HaveKeyWithValue("group_id", "deadbeef"))
		Expect(trace).To(HaveKey("event_fields"))
		for i, ef := range trace["event_fields"].([]interface{}) {
			Expect(ef.(string)).To(Equal(eventFields[i]))
		}
		Expect(trace).To(HaveKey("vantage_point"))
		vantagePoint := trace["vantage_point"].(map[string]interface{})
		Expect(vantagePoint).To(HaveKeyWithValue("type", "server"))
	})

	Context("Events", func() {
		exportAndParse := func() (time.Time, string /* category */, string /* event */, map[string]interface{}) {
			buf := &bytes.Buffer{}
			Expect(tracer.Export(buf)).To(Succeed())

			m := make(map[string]interface{})
			Expect(json.Unmarshal(buf.Bytes(), &m)).To(Succeed())
			Expect(m).To(HaveKey("traces"))
			traces := m["traces"].([]interface{})
			Expect(traces).To(HaveLen(1))
			trace := traces[0].(map[string]interface{})
			Expect(trace).To(HaveKey("events"))
			events := trace["events"].([]interface{})
			Expect(events).To(HaveLen(1))
			ev := events[0].([]interface{})
			Expect(ev).To(HaveLen(4))
			return time.Unix(0, int64(1e6*ev[0].(float64))), ev[1].(string), ev[2].(string), ev[3].(map[string]interface{})
		}

		It("records a sent packet", func() {
			now := time.Now()
			tracer.SentPacket(
				now,
				&wire.ExtendedHeader{
					Header: wire.Header{
						IsLongHeader:     true,
						Type:             protocol.PacketTypeHandshake,
						DestConnectionID: protocol.ConnectionID{1, 2, 3, 4, 5, 6, 7, 8},
						SrcConnectionID:  protocol.ConnectionID{4, 3, 2, 1},
						Version:          protocol.VersionTLS,
					},
					PacketNumber: 1337,
				},
				[]wire.Frame{
					&wire.MaxStreamDataFrame{StreamID: 42, ByteOffset: 987},
					&wire.StreamFrame{StreamID: 123, Offset: 1234, Data: []byte("foobar"), FinBit: true},
				},
			)
			t, category, eventName, ev := exportAndParse()
			Expect(t).To(BeTemporally("~", now, time.Millisecond))
			Expect(category).To(Equal("transport"))
			Expect(eventName).To(Equal("packet_sent"))
			Expect(ev).To(HaveKeyWithValue("packet_type", "handshake"))
			Expect(ev).To(HaveKey("header"))
			Expect(ev).To(HaveKey("frames"))
			Expect(ev["frames"].([]interface{})).To(HaveLen(2))
		})
	})
})
