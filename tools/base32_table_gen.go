// Tool, that generates base32 tables:
// 		one, that maps base32 index to the corresponding character
// 		and the other, that maps character (byte) to the corresponding base32 index
// Since this is only a tool, it should be ignored in the build process

// +build ignore

package main

import (
	"fmt"
	"bytes"
)

var chars = [...]byte{
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'k',
	'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'x',
	'y', 'z',
}


func main() {
	// sanity check
	if len(chars) != 32 {
		panic("Length of base32 chars is too big")
	}

	// include standard code generator warning and package
	fmt.Println("// Code generated by base32_table_gen. DO NOT EDIT.")
	fmt.Println("package gen")
	fmt.Println()

	maxChar := byte(0);
	reverseCharIndices := make(map[byte]int)

	// generate base32 index to chars, in addition calculate reverse index and do sanity check
	fmt.Printf("var Chars = [%d]byte{", len(chars))
	for i := 0; i < len(chars); i++ {
		ch := chars[i]
		lowerChars := bytes.ToLower([]byte{ch})
		upperChars := bytes.ToUpper([]byte{ch})
		if len(lowerChars) != 1 || len(upperChars) != 1 {
			panic("lowercase & uppercase chars len should be 1")
		}

		lowerChar := lowerChars[0]
		upperChar := upperChars[0]

		fmt.Printf("'%c',", lowerChar)

		// ensure there is no duplicate char
		_, contains := reverseCharIndices[ch]
		if contains {
			panic(fmt.Sprintf("duplicate char %c", ch))
		}

		reverseCharIndices[lowerChar], reverseCharIndices[upperChar] = i, i

		if lowerChar > maxChar {
			maxChar = lowerChar
		}
		if upperChar > maxChar {
			maxChar = lowerChar
		}
	}
	fmt.Println("}")

	// generate chars to index array
	fmt.Printf("var CharToIndex = [%d]int{", maxChar + 1)
	for i := byte(0); i <= maxChar; i++ {
		ch, contains := reverseCharIndices[i]
		charIndex := -1
		if contains {
			charIndex = int(ch)
		}

		fmt.Printf("%d,", charIndex)
	}
	fmt.Println("}")
}
