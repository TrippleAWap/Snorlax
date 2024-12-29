package PrismicDatabase

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type AvatarEntry struct {
	Id   string
	Name string
}

func decompressUTF16ToUTF8(data []byte) string {
	// Check for BOM and remove it if present
	bom := []byte{0xFE, 0xFF}
	if len(data) >= len(bom) && data[0] == bom[0] && data[1] == bom[1] {
		data = data[2:]
	}

	result := ""
	for i := 0; i < len(data); i += 2 {
		utf16Char := int(data[i])<<8 | int(data[i+1])

		// Extract UTF-8 characters from the integer
		utf8Chars := []rune{}
		for utf16Char > 0 {
			utf8Char := rune(utf16Char & 0x3F)
			utf16Char >>= 6

			if utf16Char == 0 {
				break
			}

			utf8Char |= 192 // Add leading bits for multi-byte UTF-8 characters
			utf8Chars = append(utf8Chars, utf8Char)
		}

		// Reverse the order of extracted UTF-8 characters
		result += string(reversedRunes(utf8Chars))
	}

	return result
}
func reversedRunes(runes []rune) []rune {
	result := make([]rune, len(runes))
	copy(result, runes)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

func reverseString(s string) string {
	runes := []rune(s)
	size := len(runes)
	for i, j := 0, size-1; i < size>>1; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func GetDatabase() ([]AvatarEntry, error) {
	var database []AvatarEntry
	req, err := http.NewRequest("GET", "https://gist.githubusercontent.com/Mwr247/ef9a06ee1d3209a558b05561f7332d8e/raw/vrcavtrdb.txt", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, err
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	entries := strings.Split(string(bytes), "\n")
	entries = entries[1:]
	for _, entry := range entries {
		// reverse the line to get the avatar name and URL
		reversed := strings.Split(reverseString(entry), "	")
		avatarName := reversed[0]
		avatarDescription := strings.Join(reversed[1:len(reversed)-1], "	")
		fmt.Println(avatarName, avatarDescription)
		avatarIdCompressed := reversed[len(reversed)-1]
		avatarId := decompressUTF16ToUTF8([]byte(avatarIdCompressed))
		fmt.Println(string(avatarId))
		database = append(database, AvatarEntry{
			Name: avatarName,
			Id:   string(avatarId),
		})
	}
	return database, nil
}
