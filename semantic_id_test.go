package ids

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
	"unicode"
)

func checkDecodingID(idgen IDCodec, id string, idValue []byte, t *testing.T) {
	newIDValue, err := idgen.Decode(id)
	if err != nil {
		t.Fatalf("id decoding failed for id=%s, err is not nil: %s", id, err)
	}

	if bytes.Compare(newIDValue, idValue) != 0 {
		t.Fatalf(
			"id decoding is wrong for id=%s: expected value=%s, actual value=%s",
			id,
			idValue,
			newIDValue)
	}
}

func TestSemanticID(t *testing.T) {

	t.Run("unprefixed fixed semantic IDs", func(t *testing.T) {
		pairs := map[string][]byte{
			"00":       {0},
			"10":       {1},
			"z7":       {255},
			"1802g040": {1, 1, 1, 1, 1},
		}

		idgen := NewCodecForNames()

		for k, v := range pairs {
			id, err := idgen.Encode(v)
			if err != nil {
				t.Fatalf("id encoding failed for id=%s, err is not nil: %s", k, err)
			}

			if id != k {
				t.Fatalf("id encoding is wrong for expected id: %s, actual: %s", k, id)
			}

			newVal, err := idgen.Decode(id)
			if err != nil {
				t.Fatalf("id decoding failed for id=%s, err is not nil: %s", k, err)
			}

			if bytes.Compare(newVal, v) != 0 {
				t.Fatalf("id decoding is wrong for id=%s: expected value=%s, actual value=%s", k, v, newVal)
			}
		}
	})

	t.Run("prefixed random semantic IDs", func(t *testing.T) {
		prefixes := [][]string{
			{},
			{"a"},
			{"a", "bb", "ccc"},
		}

		for _, p := range prefixes {
			idgen := NewCodecForNames(p...)
			for i := 1; i <= MaxBytesIDSize; i++ {
				idValue := make([]byte, i)
				rand.Read(idValue)

				id, err := idgen.Encode(idValue)
				if err != nil {
					t.Fatalf("id encoding failed err=%s", err)
				}

				checkDecodingID(idgen, id, idValue, t)

				// create syntetic ID containing uppercase and lowercase characters
				buf := bytes.Buffer{}
				for _, ch := range string(id) {
					if rand.Intn(2) == 0 {
						buf.WriteByte(byte(unicode.ToLower(rune(ch))))
					} else {
						buf.WriteByte(byte(unicode.ToUpper(rune(ch))))
					}
				}

				checkDecodingID(idgen, string(buf.String()), idValue, t)
			}
		}
	})

	t.Run("encode same IDs", func(t *testing.T) {
		name1 := "myid1"
		name2 := "users"
		idValue := []byte{1, 2, 3}

		idgen1 := NewCodecForNames(name1, name2)
		idgen2 := NewCodecForNames([]string{name1, name2}...)

		id1, err := idgen1.Encode(idValue)
		if err != nil {
			t.Fatalf("idgen1: can't encode value=%s", idValue)
		}

		id2, err := idgen2.Encode(idValue)
		if err != nil {
			t.Fatalf("idgen2: can't encode value=%s", idValue)
		}

		if id1 != id2 {
			t.Fatalf("id mismatch, id1=%s, id2=%s", id1, id2)
		}
	})

	t.Run("encode malformed semantic IDs", func(t *testing.T) {
		idgen := NewCodecForNames("test")
		_, err := idgen.Encode([]byte{})
		if err != ErrIDEmpty {
			t.Error("empty byte array shall not be encoded")
		}

		_, err = idgen.Encode(make([]byte, MaxBytesIDSize+1))
		if err != ErrIDTooBig {
			t.Error("large array should not be encoded")
		}
	})

	t.Run("decode malformed semantic IDs, get prefix", func(t *testing.T) {
		prefixes := map[string][]string{
			"":     {},
			"abc-": {"abc"},
			"a-b-": {"a", "b"},
		}
		for k, v := range prefixes {
			idgen := NewCodecForNames(v...)
			_, err := idgen.Decode(k)
			if err != ErrMalformedID {
				t.Error("decode shall prohibit empty IDs")
			}

			assertSamePrefix(t, k, idgen.GetPrefix())

			newID, err := idgen.Encode([]byte{1})
			if err != nil {
				t.Fatalf("unable to encode ID: %v", err)
			}

			assertSamePrefix(t, idgen.GetPrefix(), GetPrefix(newID))
		}
	})

	t.Run("decode invalid chars in semantic IDs", func(t *testing.T) {
		prefixes := map[string][]string{
			"":     {},
			"abc-": {"abc"},
			"a-b-": {"a", "b"},
		}
		for k, v := range prefixes {
			idgen := NewCodecForNames(v...)

			// sanity check: idgen should allow decoding legitimate IDs
			_, err := idgen.Decode(k + "00")
			if err != nil {
				t.Fatal("decode shall allow legitimate characters")
			}

			// actual verification for illegal characters
			_, err = idgen.Decode(k + "0!")
			if err != ErrInvalidChar {
				t.Error("decode shall prohibit IDs with illegal chars")
			}
		}
	})

	t.Run("equivalence of ID codecs that use same prefix names in distinct registers", func(t *testing.T) {
		idgen1 := NewCodecForNames("abc", "def")
		idgen2 := NewCodecForNames("ABC", "DeF")
		assertSamePrefix(t, idgen1.GetPrefix(), idgen2.GetPrefix())

		idValue := []byte{1, 2, 3}

		id1, err := idgen1.Encode(idValue)
		if err != nil {
			t.Fatal("unable to encode legitimate value")
		}

		id2, err := idgen2.Encode(idValue)
		if err != nil {
			t.Fatal("unable to encode legitimate value")
		}

		checkDecodingID(idgen1, id2, idValue, t)
		checkDecodingID(idgen2, id1, idValue, t)
	})

	t.Run("get prefix", func(t *testing.T) {
		assertSamePrefix(t, "", GetPrefix(""))
		assertSamePrefix(t, "", GetPrefix("123"))
		assertSamePrefix(t, "a-", GetPrefix("a-1"))
		assertSamePrefix(t, "a-bb-cc123-", GetPrefix("a-Bb-cC123-1"))
	})
}

func assertSamePrefix(t *testing.T, expected string, actual string) {
	if strings.Compare(expected, actual) != 0 {
		t.Fatalf("unexpected prefix '%s', wanted '%s'", actual, expected)
	}
}
