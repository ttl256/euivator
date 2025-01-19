package hwaddr

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"
	"strings"
)

const (
	ByteHex  = 2
	EUI48Len = 6
	EUI64Len = 8
)

const (
	EUI48HexLen = 2 * EUI48Len
	EUI64HexLen = 2 * EUI64Len
)

type ParseError struct {
	Input string
	Msg   string
	Err   error
}

func (e ParseError) Error() string {
	var baseMsg = "invalid hardware address"
	fullMsg := fmt.Sprintf("%s %q", baseMsg, e.Input)
	if e.Msg != "" {
		fullMsg = fmt.Sprintf("%s: %s", fullMsg, e.Msg)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", fullMsg, e.Err)
	}
	return fullMsg
}

func (e ParseError) Unwrap() error {
	return e.Err
}

var (
	ErrInputTooShort           = errors.New("input is too short")
	ErrInputTooLong            = errors.New("input is too long")
	ErrInputUnbalanced         = errors.New("input is unbalanced")
	ErrInputUnexpectedNumBytes = errors.New("input contains unexpected number of bytes")
)

/*
ParseAddr parses a string of EUI48/EUI64 into a slice of bytes.

Supported formats:

XXXXXXXXXXXX
XXXXXXXXXXXXXXXX
XX:XX:XX:XX:XX:XX
XX:XX:XX:XX:XX:XX:XX:XX
XX-XX-XX-XX-XX-XX
XX-XX-XX-XX-XX-XX-XX-XX
XXXX.XXXX.XXXX
XXXX.XXXX.XXXX.XXXX
*/
//nolint: mnd,gocognit // fine
func ParseAddr(s string) ([]byte, error) {
	var err error

	if len(s) < EUI48HexLen {
		return nil, ParseError{
			Input: s, Msg: fmt.Sprintf("input length must be >= %d, got %d", EUI48HexLen, len(s)), Err: ErrInputTooShort,
		}
	}
	if len(s) > 23 {
		return nil, ParseError{
			Input: s, Msg: fmt.Sprintf("input length must be <= %d, got %d", 23, len(s)), Err: ErrInputTooLong,
		}
	}

	switch {
	case s[2] == ':' || s[2] == '-':
		if (len(s)+1)%3 != 0 {
			return nil, ParseError{Input: s, Msg: "", Err: ErrInputUnbalanced}
		}

		n := (len(s) + 1) / 3
		if n != EUI48Len && n != EUI64Len {
			return nil, ParseError{Input: s, Msg: "", Err: ErrInputUnexpectedNumBytes}
		}

		r := make([]byte, 0, n)

		for i := 0; i < len(s); i += 3 {
			r, err = hex.AppendDecode(r, []byte(s[i:i+2]))
			if err != nil {
				return nil, ParseError{Input: s, Msg: "", Err: err}
			}
		}

		return r, nil
	case s[4] == '.':
		if (len(s)+1)%5 != 0 {
			return nil, ParseError{Input: s, Msg: "", Err: ErrInputUnbalanced}
		}

		n := 2 * (len(s) + 1) / 5
		if n != EUI48Len && n != EUI64Len {
			return nil, ParseError{Input: s, Msg: "", Err: ErrInputUnexpectedNumBytes}
		}

		r := make([]byte, 0, n)

		for i := 0; i < len(s); i += 5 {
			r, err = hex.AppendDecode(r, []byte(s[i:i+4]))
			if err != nil {
				return nil, ParseError{Input: s, Msg: "", Err: err}
			}
		}

		return r, nil
	default:
		if len(s)%2 != 0 {
			return nil, ParseError{Input: s, Msg: "", Err: ErrInputUnbalanced}
		}

		n := len(s) / 2
		if n != EUI48Len && n != EUI64Len {
			return nil, ParseError{Input: s, Msg: "", Err: ErrInputUnexpectedNumBytes}
		}

		r := make([]byte, 0, n)

		r, err = hex.AppendDecode(r, []byte(s))
		if err != nil {
			return nil, ParseError{Input: s, Msg: "", Err: err}
		}

		return r, nil
	}
}

// EUI48FromBytes convert a slice of bytes into [EUI48].
func EUI48FromBytes(s []byte) (EUI48, error) {
	if len(s) != EUI48Len {
		return EUI48{}, fmt.Errorf("invalid slice length %d, expected %d", len(s), EUI48Len)
	}

	var r [EUI48Len]byte
	copy(r[:], s)

	return EUI48(r), nil
}

type EUI48 [6]byte

// Equivalent to ToString(a[:], []byte{':'}, 1).
func (a EUI48) String() string {
	return AsColon(a[:])
}

func (a EUI48) EUI64Modified() EUI64 {
	var eui64 = [8]byte{}

	eui64[0] = a[0]
	eui64[1] = a[1]
	eui64[2] = a[2]

	eui64[3] = 0xFF
	eui64[4] = 0xFE

	eui64[5] = a[3]
	eui64[6] = a[4]
	eui64[7] = a[5]

	eui64[0] ^= 0x02

	return EUI64(eui64)
}

// EUI64FromBytes convert a slice of bytes into [EUI64].
func EUI64FromBytes(s []byte) (EUI64, error) {
	if len(s) != EUI64Len {
		return EUI64{}, fmt.Errorf("invalid slice length %d, expected %d", len(s), EUI48Len)
	}

	var r [EUI64Len]byte
	copy(r[:], s)

	return EUI64(r), nil
}

type EUI64 [8]byte

// Equivalent to ToString(a[:], []byte{':'}, 1).
func (a EUI64) String() string {
	return AsColon(a[:])
}

// AppendToPrefix writes [EUI64] into 8 least significant bytes of a prefix.
func AppendToPrefix(prefix netip.Prefix, eui64 EUI64) netip.Addr {
	prefixBytes := prefix.Addr().As16()
	copy(prefixBytes[8:], eui64[:])
	return netip.AddrFrom16(prefixBytes)
}

func AsColon(addr []byte) string {
	return ToString(addr, []byte{':'}, 1)
}

func AsDash(addr []byte) string {
	return ToString(addr, []byte{'-'}, 1)
}

func AsDot(addr []byte) string {
	return ToString(addr, []byte{'.'}, ByteHex)
}

func AsPlain(addr []byte) string {
	return ToString(addr, []byte{}, 0)
}

/*
ToString converts a slice of bytes into a hardware address string
representation. Each byte in the input slice `addr` is encoded as two
hexadecimal characters. The resulting string will contain separators (specified
by `sep`) after every `count` bytes of the input. Panics on count < 0.
*/
func ToString(addr []byte, sep []byte, count int) string {
	if count < 0 {
		panic(fmt.Sprintf("count must be a non-negative integer, got %d", count))
	}

	var maxLen = max(len(addr)*(ByteHex+len(sep))-len(sep), 0)

	if count == 0 {
		count = maxLen
	}

	var s strings.Builder
	s.Grow(maxLen)

	var buf = []byte{0, 0}

	for i := range addr {
		if i > 0 && i%count == 0 {
			s.Write(sep)
		}
		hex.Encode(buf, addr[i:i+1])
		s.Write(buf)
	}

	return s.String()
}
