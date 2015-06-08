package manta

import (
	"github.com/dotabuff/manta/dota"
)

// Internal parser for callback OnCDemoPacket, responsible for extracting
// multiple inner packets from a single CDemoPacket. This is the main payload
// for most data in the replay.
func (p *Parser) onCDemoPacket(m *dota.CDemoPacket) error {
	r := newDemoPacketReader(m.GetData())

	for r.hasNext() {
		t, buf := r.readNext()
		if err := p.CallByPacketType(t, buf); err != nil {
			return err
		}
	}

	return nil
}

// Creates a new demoPacketHandler, used to read data in the expected format.
func newDemoPacketReader(buf []byte) *demoPacketReader {
	return &demoPacketReader{newReader(buf)}
}

// Reads a series of inner packets from a CDemoPacket buffer
type demoPacketReader struct {
	r *reader
}

// Determines whether or not another packet is available.
// XXX TODO: this seems wrong, we may be skipping the last packet or some
// other value at the end of the buffer.
func (r *demoPacketReader) hasNext() bool {
	return r.r.rem_bits() > 10
}

// Reads the next packet, returning a type and inner buffer.
// XXX TODO: detail our knowledge of the structure of this packet.
func (r *demoPacketReader) readNext() (int32, []byte) {
	t := r.r.read_bits(6)

	if header := t >> 4; header != 0 {
		bits := int(header*4 + (((2 - header) >> 31) & 16))
		t = (t & 15) | (r.r.read_bits(bits) << 4)
	}

	size := r.r.read_var_uint32()
	buf := r.r.read_bytes(int(size))

	return int32(t), buf
}

func (p *Parser) onCDemoFullPacket(m *dota.CDemoFullPacket) error {
	if m.Packet != nil {
		if err := p.onCDemoPacket(m.GetPacket()); err != nil {
			return err
		}
	}

	if m.StringTable != nil {
		if err := p.onCDemoStringTables(m.GetStringTable()); err != nil {
			return err
		}
	}

	return nil
}