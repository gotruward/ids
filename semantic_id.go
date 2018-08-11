package ids

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/gotruward/ids/gen"
)

// MaxBytesIDSize defines upper bound limit for number of bytes in a given semantic ID
const MaxBytesIDSize = 256

var (
	// ErrIDTooBig happens when given ID is too big to be encoded as semantic ID
	// see also MaxBytesIDSize that defines this limit
	ErrIDTooBig = errors.New("value is too big to be encoded into semantic ID")

	// ErrIDEmpty happens when empty ID is asked to be decoded
	ErrIDEmpty = errors.New("can't encode empty value to semantic ID")

	// ErrInvalidChar happens when invalid character is found in a given semantic ID
	ErrInvalidChar = errors.New("given semantic ID contains invalid character")

	// ErrMalformedID happens when given semantic ID is malformed
	ErrMalformedID = errors.New("invalid semantic ID")
)

// GetPrefix returns prefix which is a part of a given semantic ID
func GetPrefix(maybeSemanticID string) string {
	lastPrefixIndex := strings.LastIndexByte(maybeSemanticID, prefixSeparator)
	if lastPrefixIndex <= 0 {
		return ""
	}
	return strings.ToLower(string([]byte(maybeSemanticID)[:lastPrefixIndex+1]))
}

// IDCodec defines an interface that abstracts out operations on the semantic IDs
type IDCodec interface {
	CanDecode(id string) bool

	Encode(value []byte) (string, error)

	Decode(id string) ([]byte, error)

	GetPrefix() string
}

// NewCodecForNames creates IDCodec for a given sequence of names
func NewCodecForNames(names ...string) IDCodec {
	lowercasedNames := make([]string, len(names))
	for index, name := range names {
		lowercasedNames[index] = strings.ToLower(name)
	}

	return &prefixedIDCodec{Names: lowercasedNames}
}

//
// Implementation
//

const prefixSeparator byte = '-'

type prefixedIDCodec struct {
	IDCodec
	Names []string
}

func newBufferWithPrefix(names []string, capacity int) *bytes.Buffer {
	buf := &bytes.Buffer{}
	buf.Grow(capacity)
	for i := 0; i < len(names); i++ {
		buf.WriteString(names[i])
		buf.WriteByte(prefixSeparator)
	}
	return buf
}

func (c *prefixedIDCodec) GetPrefix() string {
	return newBufferWithPrefix(c.Names, len(c.Names)*8).String()
}

func (c *prefixedIDCodec) CanDecode(id string) bool {
	_, err := computeAndValidatePrefix(c, id)
	if err != nil {
		return false
	}

	return true
}

func (c *prefixedIDCodec) Encode(value []byte) (string, error) {
	valueLen := len(value)
	if value == nil || valueLen == 0 {
		return "", ErrIDEmpty
	}

	if valueLen > MaxBytesIDSize {
		return "", ErrIDTooBig
	}

	capacity := getPrefixLength(c.Names) + int(getEncodedSize(uint(valueLen)))
	buf := newBufferWithPrefix(c.Names, capacity)
	appendBytes(value, buf)
	if buf.Len() != capacity {
		return "", fmt.Errorf("internal: unexpected buffer size") // shouldn't happen
	}
	return buf.String(), nil
}

func (c *prefixedIDCodec) Decode(id string) ([]byte, error) {
	prefixLength, err := computeAndValidatePrefix(c, id)
	if err != nil {
		return nil, err
	}

	return decodeBytes(id, prefixLength, len(id))
}

//
// prefixed SemanticID codec private methods
//

func getPrefixLength(names []string) int {
	result := 0
	for i := 0; i < len(names); i++ {
		result = result + len(names[i]) + 1
	}
	return result
}

func computeAndValidatePrefix(c *prefixedIDCodec, id string) (int, error) {
	idLen := len(id)

	charIndex := 0
	for nameIndex := 0; nameIndex < len(c.Names); nameIndex++ {
		name := c.Names[nameIndex]
		for nameCharIndex := 0; nameCharIndex < len(name); nameCharIndex++ {
			// check, that current prefix part matches corresponding SemanticID region
			nameChar := name[nameCharIndex]
			if charIndex < idLen {
				ch := byte(unicode.ToLower(rune(id[charIndex])))
				charIndex++

				if ch == nameChar {
					continue
				}
			}

			return 0, ErrMalformedID
		}

		if charIndex < idLen {
			// check, that character after name matches prefix
			ch := id[charIndex]
			charIndex++

			if ch == prefixSeparator {
				continue
			}
		}

		return 0, ErrMalformedID
	}

	// charIndex now should be at the beginning of SemanticID
	if (idLen - charIndex) > 0 {
		// validate SemanticID body
		for i := uint(charIndex); i < uint(idLen); i++ {
			_, err := getBaseCharCode(id, i)
			if err != nil {
				return 0, err
			}
		}

		// OK: return prefix length
		return charIndex, nil
	}

	return 0, ErrMalformedID
}

//
// PadlessBase32
//

const byteSize uint = 8
const baseBits uint = 5
const base uint = 1 << baseBits
const baseMask = uint8(base - 1)

func getBaseChar(index uint8) uint8 {
	return gen.Chars[index]
}

func getBaseCharCode(value string, charPos uint) (uint8, error) {
	ch := int(value[int(charPos)])

	if ch < len(gen.CharToIndex) {
		index := gen.CharToIndex[ch]
		if index >= 0 {
			return uint8(index), nil
		}
	}

	return uint8(0), ErrInvalidChar
}

func getEncodedSize(size uint) uint {
	return (size*byteSize + baseBits - 1) / baseBits
}

func appendBytes(body []byte, buf *bytes.Buffer) {
	bodyLen := uint(len(body))
	bodyBits := byteSize * bodyLen
	fullBase32ElemCount := bodyBits / baseBits
	partialBase32ElemBits := bodyBits % baseBits

	for startPosByte, offsetBitPos, i := uint(0), uint(0), uint(0); i < fullBase32ElemCount; i++ {
		endBitPos := offsetBitPos + baseBits
		elem := body[startPosByte]

		// NB: element shall be within the byte range
		base32ElemIndex := (elem >> uint8(offsetBitPos)) & baseMask
		if endBitPos > byteSize {
			tailBitCount := endBitPos - byteSize
			base32ElemIndex |= (body[startPosByte+1] & ((1 << tailBitCount) - 1)) << (baseBits - tailBitCount)
			startPosByte++
			offsetBitPos = tailBitCount
		} else {
			offsetBitPos = endBitPos
		}

		buf.WriteByte(getBaseChar(base32ElemIndex))
	}

	if partialBase32ElemBits > 0 {
		lastElem := body[bodyLen-1]
		base32ElemIndex := lastElem >> (byteSize - partialBase32ElemBits)
		buf.WriteByte(getBaseChar(base32ElemIndex))
	}
}

func decodeBytes(value string, startPos int, endPos int) ([]byte, error) {
	if startPos < 0 {
		// shouldn't happen
		return nil, fmt.Errorf("internal: negative startPos=%d", startPos)
	}

	if endPos <= startPos {
		// shouldn't happen
		return nil, fmt.Errorf("internal: endPos=%d is not greater than startPos=%d", endPos, startPos)
	}

	length := endPos - startPos
	byteCount := (uint(length) * baseBits) / byteSize
	result := make([]byte, byteCount)

	tailBits, tailBitsCount, resultBitsOffset, resultPos := uint8(0), uint(0), uint(0), uint(0)
	for charPos := uint(startPos); (resultPos < byteCount) && (charPos < uint(endPos)); {
		if tailBitsCount > 0 {
			result[resultPos] = result[resultPos] | tailBits
			resultBitsOffset = tailBitsCount
			tailBitsCount = 0
			charPos++
			continue
		}

		base32Digit, err := getBaseCharCode(value, charPos)
		if err != nil {
			return nil, err
		}

		nextBitOffset := resultBitsOffset + baseBits
		nextResultPos := resultPos

		if nextBitOffset > byteSize {
			// this entire digit doesn't fits into current byte, apply only part of it
			headBitCount := byteSize - resultBitsOffset
			tailBits = base32Digit >> headBitCount
			tailBitsCount = nextBitOffset - byteSize
			nextResultPos++
			nextBitOffset = 0
		} else if nextBitOffset == byteSize {
			// proceed to the next byte and next base32 element (bit bounds matched)
			nextResultPos++
			charPos++
			nextBitOffset = 0
		} else {
			// this current character fits the entire byte and there still enough place in result byte to put something else
			charPos++
		}

		result[resultPos] = result[resultPos] | (base32Digit << resultBitsOffset)
		resultPos = nextResultPos
		resultBitsOffset = nextBitOffset
	}

	return result, nil
}
