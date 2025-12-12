package tpm

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/google/go-attestation/attest"
)

// FormatEventData returns a human-readable representation of the event data
func FormatEventData(event attest.Event) string {
	switch event.Type {
	case attest.EventType(0x0000000D), // EV_IPL
		attest.EventType(0x00000008), // EV_S_CRTM_VERSION
		attest.EventType(0x00000005), // EV_ACTION
		attest.EventType(0x80000007): // EV_EFI_ACTION
		return cleanString(string(event.Data))

	case attest.EventType(0x00000004): // EV_SEPARATOR
		if len(event.Data) == 4 {
			val := binary.LittleEndian.Uint32(event.Data)
			return fmt.Sprintf("Separator: 0x%08x", val)
		}
		return "Separator (invalid length)"

	case attest.EventType(0x80000001), // EV_EFI_VARIABLE_DRIVER_CONFIG
		attest.EventType(0x80000002), // EV_EFI_VARIABLE_BOOT
		attest.EventType(0x800000e0): // EV_EFI_VARIABLE_AUTHORITY
		if name, err := parseUEFIVariableName(event.Data); err == nil {
			return fmt.Sprintf("Variable: %s", name)
		}
		return "UEFI Variable (parse error)"

	case attest.EventType(0x80000003), // EV_EFI_BOOT_SERVICES_APPLICATION
		attest.EventType(0x80000004), // EV_EFI_BOOT_SERVICES_DRIVER
		attest.EventType(0x80000005): // EV_EFI_RUNTIME_SERVICES_DRIVER
		if path, err := parseEFIImageLoadPath(event.Data); err == nil {
			return fmt.Sprintf("Image Load: %s", path)
		}
		return "Image Load (parse error)"

	default:
		if isPrintable(event.Data) {
			return cleanString(string(event.Data))
		}
		return hex.EncodeToString(event.Data)
	}
}

func isPrintable(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	printable := 0
	for _, b := range data {
		if b >= 32 && b <= 126 {
			printable++
		}
	}
	return float64(printable)/float64(len(data)) > 0.9
}

func cleanString(s string) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, s))
}

type efiGUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

func parseUEFIVariableName(data []byte) (string, error) {
	r := bytes.NewReader(data)
	var header struct {
		VariableName       efiGUID
		UnicodeNameLength  uint64
		VariableDataLength uint64
	}
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return "", err
	}
	if header.UnicodeNameLength > 2048 {
		return "", fmt.Errorf("name too long")
	}
	nameBuf := make([]uint16, header.UnicodeNameLength)
	if err := binary.Read(r, binary.LittleEndian, &nameBuf); err != nil {
		return "", err
	}
	return string(utf16.Decode(nameBuf)), nil
}

func parseEFIImageLoadPath(data []byte) (string, error) {
	r := bytes.NewReader(data)
	var header struct {
		LoadAddr      uint64
		Length        uint64
		LinkAddr      uint64
		DevicePathLen uint64
	}
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return "", err
	}
	return fmt.Sprintf("Addr: 0x%x, Len: %d", header.LoadAddr, header.Length), nil
}
