package tpm

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"

	"github.com/google/go-attestation/attest"
)

func TestFormatEventData(t *testing.T) {
	tests := []struct {
		name     string
		event    attest.Event
		expected string
	}{
		{
			name: "ASCII Event",
			event: attest.Event{
				Type: attest.EventType(0x0000000D), // EV_IPL
				Data: []byte("Grub Bootloader"),
			},
			expected: "Grub Bootloader",
		},
		{
			name: "Separator",
			event: attest.Event{
				Type: attest.EventType(0x00000004), // EV_SEPARATOR
				Data: []byte{0, 0, 0, 0},
			},
			expected: "Separator: 0x00000000",
		},
		{
			name: "UEFI Variable",
			event: attest.Event{
				Type: attest.EventType(0x80000001), // EV_EFI_VARIABLE_DRIVER_CONFIG
				Data: func() []byte {
					buf := make([]byte, 0)
					// GUID
					buf = append(buf, make([]byte, 16)...)
					// UnicodeNameLength (4 chars)
					lenBuf := make([]byte, 8)
					binary.LittleEndian.PutUint64(lenBuf, 4)
					buf = append(buf, lenBuf...)
					// VariableDataLength (0)
					buf = append(buf, make([]byte, 8)...)
					// UnicodeName "Test"
					uName := utf16.Encode([]rune("Test"))
					for _, u := range uName {
						uBuf := make([]byte, 2)
						binary.LittleEndian.PutUint16(uBuf, u)
						buf = append(buf, uBuf...)
					}
					return buf
				}(),
			},
			expected: "Variable: Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatEventData(tt.event)
			if got != tt.expected {
				t.Errorf("FormatEventData() = %v, want %v", got, tt.expected)
			}
		})
	}
}
