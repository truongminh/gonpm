package storage

// credit https://github.com/cloudfoundry/bytefmt
import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

const (
	aBYTE = 1 << (10 * iota)
	aKILOBYTE
	aMEGABYTE
	aGIGABYTE
	aTERABYTE
)

var errInvalidByteQuantityError = errors.New("byte quantity must be a positive integer with a unit of measurement like M, MB, MiB, G, GiB, or GB")

// ToBytes parses a string formatted by ByteSize as bytes. Note binary-prefixed and SI prefixed units both mean a base-2 units
// KB = K = KiB	= 1024
// MB = M = MiB = 1024 * K
// GB = G = GiB = 1024 * M
// TB = T = TiB = 1024 * G
func ToBytes(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	i := strings.IndexFunc(s, unicode.IsLetter)

	if i == -1 {
		return 0, errInvalidByteQuantityError
	}

	bytesString, multiple := s[:i], s[i:]
	bytes, err := strconv.ParseFloat(bytesString, 64)
	if err != nil || bytes < 0 {
		return 0, errInvalidByteQuantityError
	}

	switch multiple {
	case "T", "TB", "TIB":
		return uint64(bytes * aTERABYTE), nil
	case "G", "GB", "GIB":
		return uint64(bytes * aGIGABYTE), nil
	case "M", "MB", "MIB":
		return uint64(bytes * aMEGABYTE), nil
	case "K", "KB", "KIB":
		return uint64(bytes * aKILOBYTE), nil
	case "B":
		return uint64(bytes), nil
	default:
		return 0, errInvalidByteQuantityError
	}
}
