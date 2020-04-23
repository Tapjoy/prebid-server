package vendorconsent

import (
	"encoding/binary"
	"fmt"
)

func parseRangeSection(data consentMetadata) (*rangeSection, error) {

	// This makes an int from bits 230-241
	if len(data) < 31 {
		return nil, fmt.Errorf("vendor consent strings using RangeSections require at least 31 bytes. Got %d", len(data))
	}
	numEntries := parseNumEntries(data)

	// Parse out the "exceptions" here.
	currentOffset := uint(242)
	consents := make([]rangeConsent, numEntries)
	for i := uint16(0); i < numEntries; i++ {
		thisConsent, bitsConsumed, err := parseRangeConsent(data, currentOffset)
		if err != nil {
			return nil, err
		}
		consents[i] = thisConsent
		currentOffset = currentOffset + bitsConsumed
	}

	return &rangeSection{
		consentMetadata: data,
		consents:        consents,
	}, nil
}

// parse the value of NumEntries, assuming this consent string uses a RangeEntry
func parseNumEntries(data []byte) uint16 {
	// This should isolate the bits [000000xx, xxxxxxxx, xx000000] to get bits 230-241 as an int
	leftByte := ((data[28] & 0x03) << 2) | (data[29] >> 6)
	rightByte := (data[29] << 2) | (data[30] >> 6)

	return binary.BigEndian.Uint16([]byte{leftByte, rightByte})
}

// RangeSection Exception implemnetations

// parseRangeConsents parses a RangeSection starting from the initial bit.
// It returns the exception, as well as the number of bits consumed by the parsing.
func parseRangeConsent(data consentMetadata, initialBit uint) (rangeConsent, uint, error) {
	// Fixes #10
	if uint(len(data)) <= initialBit/8 {
		return nil, 0, fmt.Errorf("bit %d was supposed to start a new RangeEntry, but the consent string was only %d bytes long", initialBit, len(data))
	}
	// If the first bit is set, it's a Range of IDs
	if isSet(data, initialBit) {
		start, err := parseUInt16(data, initialBit+1)
		if err != nil {
			return nil, 0, err
		}
		end, err := parseUInt16(data, initialBit+17)
		if err != nil {
			return nil, 0, err
		}
		if start == 0 {
			return nil, 0, fmt.Errorf("bit %d range entry exclusion starts at 0, but the min vendor ID is 1", initialBit)
		}
		if end > data.MaxVendorID() {
			return nil, 0, fmt.Errorf("bit %d range entry exclusion ends at %d, but the max vendor ID is %d", initialBit, end, data.MaxVendorID())
		}
		if end <= start {
			return nil, 0, fmt.Errorf("bit %d range entry excludes vendors [%d, %d]. The start should be less than the end", initialBit, start, end)
		}
		return rangeVendorConsent{
			startID: start,
			endID:   end,
		}, uint(33), nil
	}

	vendorID, err := parseUInt16(data, initialBit+1)
	if err != nil {
		return nil, 0, err
	}
	if vendorID == 0 || vendorID > data.MaxVendorID() {
		return nil, 0, fmt.Errorf("bit %d range entry excludes vendor %d, but only vendors [1, %d] are valid", initialBit, vendorID, data.MaxVendorID())
	}

	return singleVendorConsent(vendorID), 17, nil
}

// parseUInt16  parses a 16-bit integer from the data array, starting at the given index
func parseUInt16(data []byte, bitStartIndex uint) (uint16, error) {
	startByte := bitStartIndex / 8
	bitStartOffset := bitStartIndex % 8
	if bitStartOffset == 0 {
		if uint(len(data)) < (startByte + 2) {
			return 0, fmt.Errorf("rangeSection expected a 16-bit vendorID to start at bit %d, but the consent string was only %d bytes long", bitStartIndex, len(data))
		}
		return binary.BigEndian.Uint16(data[startByte : startByte+2]), nil
	}
	if uint(len(data)) < (startByte + 3) {
		return 0, fmt.Errorf("rangeSection expected a 16-bit vendorID to start at bit %d, but the consent string was only %d bytes long", bitStartIndex, len(data))
	}

	shiftComplement := 8 - bitStartOffset

	// Take the rightmost bits of the left byte, and the leftmost bits of the middle byte
	leftByte := (data[startByte] & (0xff >> bitStartOffset)) << bitStartOffset
	leftByte = leftByte | (data[startByte+1] >> shiftComplement)

	// Take the rightmost bits of the middle byte, and the leftmost bits of the right byte
	rightByte := data[startByte+2] & (0xff << shiftComplement)
	rightByte = (rightByte >> shiftComplement) | (data[startByte+1] << bitStartOffset)

	return binary.BigEndian.Uint16([]byte{leftByte, rightByte}), nil
}

// A RangeConsents encodes consents that have been registered.
type rangeSection struct {
	consentMetadata
	consents []rangeConsent
}

// VendorConsents implementation
func (p rangeSection) VendorConsent(id uint16) bool {
	if id < 1 || id > p.MaxVendorID() {
		return false
	}

	for i := 0; i < len(p.consents); i++ {
		if p.consents[i].Contains(id) {
			return true
		}
	}
	return false
}

// A RangeSection has a default consent value and a list of "exceptions". This represents an "exception" blob
type rangeConsent interface {
	Contains(id uint16) bool
}

// This is a RangeSection exception for a single vendor.
type singleVendorConsent uint16

func (e singleVendorConsent) Contains(id uint16) bool {
	return uint16(e) == id
}

// This is a RangeSection exception for a range of IDs.
// The start and end bounds here are inclusive.
type rangeVendorConsent struct {
	startID uint16
	endID   uint16
}

func (e rangeVendorConsent) Contains(id uint16) bool {
	return e.startID <= id && e.endID >= id
}
