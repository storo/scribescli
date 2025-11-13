package tui

import "strings"

// safeRepeat safely repeats a string, ensuring count is never negative
func safeRepeat(s string, count int) string {
	if count < 0 {
		return ""
	}
	return strings.Repeat(s, count)
}

// safePadding calculates padding ensuring it's never negative
func safePadding(totalWidth, contentWidth int) int {
	padding := totalWidth - contentWidth
	if padding < 0 {
		return 0
	}
	return padding
}

// ensureMinWidth ensures width is at least minWidth
func ensureMinWidth(width, minWidth int) int {
	if width < minWidth {
		return minWidth
	}
	return width
}

// ensureMinHeight ensures height is at least minHeight
func ensureMinHeight(height, minHeight int) int {
	if height < minHeight {
		return minHeight
	}
	return height
}
