package wallet

import (
	"bytes"
	"math/big"
)

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// Base58Encode converts bytes into a Base58 string.
func Base58Encode(input []byte) string {
	x := new(big.Int).SetBytes(input)
	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	var result []byte
	for x.Cmp(zero) > 0 {
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}

	for _, b := range input {
		if b != 0x00 {
			break
		}
		result = append(result, b58Alphabet[0])
	}

	reverseBytes(result)
	return string(result)
}

// Base58Decode converts a Base58 string into bytes.
func Base58Decode(input string) []byte {
	result := big.NewInt(0)
	base := big.NewInt(int64(len(b58Alphabet)))

	for _, b := range []byte(input) {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		if charIndex < 0 {
			return nil
		}

		result.Mul(result, base)
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()
	leadingZeroes := 0
	for leadingZeroes < len(input) && input[leadingZeroes] == b58Alphabet[0] {
		leadingZeroes++
	}

	return append(bytes.Repeat([]byte{0x00}, leadingZeroes), decoded...)
}

func reverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
