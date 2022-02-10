package bytefmt

// from: github.com/cloudfoundry/bytefmt by Apache License

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

const (
	sizeByte = 1 << (10 * iota)
	sizeKilo
	sizeMega
	sizeGiga
	sizeTera
	sizePeta
	sizeExa
)

var errInvalidByteQuantity = errors.New("byte quantity must be a positive integer with a unit of measurement like M, MB, MiB, G, GiB, or GB")

// FormatSize returns a human-readable byte string of the form 10M, 12.5K, and so forth.  The following units are available:
//  E: Exabyte
//  P: Petabyte
//  T: Terabyte
//  G: Gigabyte
//  M: Megabyte
//  K: Kilobyte
//  B: Byte
// The unit that results in the smallest number greater than or equal to 1 is always chosen.
func FormatSize(bytes uint64) string {
	unit := ""
	value := float64(bytes)

	switch {
	case bytes >= sizeExa:
		unit = "E"
		value = value / sizeExa
	case bytes >= sizePeta:
		unit = "P"
		value = value / sizePeta
	case bytes >= sizeTera:
		unit = "T"
		value = value / sizeTera
	case bytes >= sizeGiga:
		unit = "G"
		value = value / sizeGiga
	case bytes >= sizeMega:
		unit = "M"
		value = value / sizeMega
	case bytes >= sizeKilo:
		unit = "K"
		value = value / sizeKilo
	case bytes >= sizeByte:
		unit = "B"
	case bytes == 0:
		return "0"
	}

	result := strconv.FormatFloat(value, 'f', 1, 64)
	result = strings.TrimSuffix(result, ".0")
	return result + unit
}

// ParseSize parses a string formatted by FormatSize as bytes. Note binary-prefixed and SI prefixed units both mean a base-2 units
// KB = K = KiB = 1024
// MB = M = MiB = 1024 * K
// GB = G = GiB = 1024 * M
// TB = T = TiB = 1024 * G
// PB = P = PiB = 1024 * T
// EB = E = EiB = 1024 * P
func ParseSize(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	i := strings.IndexFunc(s, unicode.IsLetter)

	if i == -1 {
		return 0, errInvalidByteQuantity
	}

	bytesString, multiple := s[:i], s[i:]
	bytes, err := strconv.ParseFloat(bytesString, 64)
	if err != nil || bytes <= 0 {
		return 0, errInvalidByteQuantity
	}

	switch multiple {
	case "E", "EB", "EIB":
		return uint64(bytes * sizeExa), nil
	case "P", "PB", "PIB":
		return uint64(bytes * sizePeta), nil
	case "T", "TB", "TIB":
		return uint64(bytes * sizeTera), nil
	case "G", "GB", "GIB":
		return uint64(bytes * sizeGiga), nil
	case "M", "MB", "MIB":
		return uint64(bytes * sizeMega), nil
	case "K", "KB", "KIB":
		return uint64(bytes * sizeKilo), nil
	case "B":
		return uint64(bytes), nil
	default:
		return 0, errInvalidByteQuantity
	}
}
