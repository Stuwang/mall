package biz

import (
	"fmt"
	"unicode/utf8"
)

func PairIDs(a, b int64) (low, high int64) {
	if a < b {
		return a, b
	}
	return b, a
}

func ConvID(a, b int64) string {
	low, high := PairIDs(a, b)
	return fmt.Sprintf("%d_%d", low, high)
}

func PreviewText(text string, maxRunes int) string {
	if maxRunes <= 0 {
		maxRunes = 50
	}
	if utf8.RuneCountInString(text) <= maxRunes {
		return text
	}
	runes := []rune(text)
	return string(runes[:maxRunes])
}

func isNewerMsg(ct, msgID, otherCt, otherMsgID int64) bool {
	if otherMsgID == 0 {
		return true
	}
	if ct > otherCt {
		return true
	}
	if ct < otherCt {
		return false
	}
	return msgID > otherMsgID
}
