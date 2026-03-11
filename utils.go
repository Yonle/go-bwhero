package main

import (
	"fmt"
	"strings"
	"sync"
)

var stringsBuilderPool = sync.Pool{
	New: func() interface{} {
		b := new(strings.Builder)
		b.Grow(128)
		return b
	},
}

func ffmpegFilterBuilder(grayscale, thumbWidth int, isVideo bool) string {
	b := stringsBuilderPool.Get().(*strings.Builder)
	b.Reset()
	defer stringsBuilderPool.Put(b)

	b.WriteString("format=yuva420p")

	if grayscale == 1 {
		b.WriteString(",hue=s=0")
	}

	if thumbWidth != 0 {
		fmt.Fprintf(b, ",scale='min(%d,iw)':-1", thumbWidth)
	}

	if isVideo {
		b.WriteString(",thumbnail")
	}

	if b.Len() == 0 {
		return "null"
	}

	return b.String()
}
