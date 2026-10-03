package main

import (
	"bytes"
	"strings"
)

func isGIF(data []byte) bool {
	// GIF header:
	// 47 49 46 38 ("GIF8")
	if len(data) < 3 {
		return false
	}

	return data[0] == 0x47 &&
		data[1] == 0x49 &&
		data[2] == 0x46
}

func isAnimatedWEBP(data []byte) bool {
	/*
	 * WEBP:
	 *
	 * 00-03: RIFF
	 * 04-07: file size
	 * 08-11: WEBP
	 * 12-15: chunk type
	 * 16-19: chunk size
	 * 20:    VP8X flags
	 *
	 * VP8X animation flag = bit 1.
	 */

	if len(data) < 21 {
		return false
	}

	if !bytes.Equal(data[0:4], []byte("RIFF")) {
		return false
	}

	if !bytes.Equal(data[8:12], []byte("WEBP")) {
		return false
	}

	// Animation metadata lives in the VP8X extended header.
	if !bytes.Equal(data[12:16], []byte("VP8X")) {
		return false
	}

	return (data[20] & 0x02) != 0
}

func isAnimatedPNG(data []byte) bool {
	// In APNG, acTL must appear before the first IDAT chunk.
	idatPos := bytes.Index(data, []byte("IDAT"))

	searchArea := data
	if idatPos >= 0 {
		searchArea = data[:idatPos]
	}

	return bytes.Index(searchArea, []byte("acTL")) >= 0
}

func isSVG(data []byte) bool {
	/*
	 * SVG is XML, so don't blindly require "<svg" to be byte 0.
	 *
	 * Handles:
	 *   - UTF-8 BOM
	 *   - whitespace
	 *   - XML declaration
	 *   - comments
	 *   - DOCTYPE
	 */

	data = bytes.TrimSpace(data)

	// UTF-8 BOM
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	data = bytes.TrimSpace(data)

	for {
		switch {
		case bytes.HasPrefix(data, []byte("<?xml")):
			end := bytes.Index(data, []byte("?>"))
			if end < 0 {
				return false
			}

			data = bytes.TrimSpace(data[end+2:])

		case bytes.HasPrefix(data, []byte("<!--")):
			end := bytes.Index(data, []byte("-->"))
			if end < 0 {
				return false
			}

			data = bytes.TrimSpace(data[end+3:])

		case bytes.HasPrefix(data, []byte("<!DOCTYPE")):
			end := bytes.IndexByte(data, '>')
			if end < 0 {
				return false
			}

			data = bytes.TrimSpace(data[end+1:])

		default:
			goto done
		}
	}

done:
	if len(data) < 5 || data[0] != '<' {
		return false
	}

	// Grab the root element name.
	data = data[1:]

	end := len(data)
	for i, c := range data {
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' ||
			c == '>' || c == '/' {
			end = i
			break
		}
	}

	name := strings.ToLower(string(data[:end]))

	return name == "svg"
}