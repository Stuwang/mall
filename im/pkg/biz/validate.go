package biz

import (
	"regexp"
	"strings"
	"unicode/utf8"

	apicommon "mall/im/api/common"
	"mall/file/mediastore"
	"mall/im/pkg/bizerror"
	"mall/kit/logger"
)

var idempotencyKeyRe = regexp.MustCompile(`^[A-Za-z0-9_-]{8,64}$`)

func validatePeer(selfID, peerID int64) error {
	if peerID <= 0 {
		logger.Errorf("im peer_user_id invalid peer=%d", peerID)
		return bizerror.ErrorBadMessage
	}
	if peerID == selfID {
		logger.Errorf("im peer is self user=%d", selfID)
		return bizerror.ErrorSelfPeer
	}
	return nil
}

func validateText(text string, maxRunes int) error {
	n := utf8.RuneCountInString(text)
	if n < 1 || n > maxRunes {
		logger.Errorf("im text length invalid n=%d max=%d", n, maxRunes)
		return bizerror.ErrorBadMessage
	}
	return nil
}

func validateContentType(ct apicommon.ContentType) error {
	switch ct {
	case apicommon.ContentType_CONTENT_TYPE_TEXT,
		apicommon.ContentType_CONTENT_TYPE_IMAGE:
		return nil
	case apicommon.ContentType_CONTENT_TYPE_GIFT:
		logger.Errorf("im content_type gift rejected type=%d", ct)
		return bizerror.ErrorBadMessage
	default:
		logger.Errorf("im content_type unsupported type=%d", ct)
		return bizerror.ErrorBadMessage
	}
}

// normalizeImageURI 校验并归一化为 local_file URI；失败返回业务错误。
func normalizeImageURI(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		logger.Errorf("im image_uri empty")
		return "", bizerror.ErrorBadMessage
	}
	uri := mediastore.Get().Normalize(raw)
	key, ok := parseLocalFileKey(uri)
	if !ok {
		logger.Errorf("im image_uri invalid raw=%s uri=%s", raw, uri)
		return "", bizerror.ErrorBadMessage
	}
	f, _, err := mediastore.Get().Open(key)
	if err != nil {
		logger.Errorf("im image_uri open key=%s err=%v", key, err)
		return "", bizerror.ErrorBadMessage
	}
	_ = f.Close()
	return uri, nil
}

func parseLocalFileKey(uri string) (string, bool) {
	const prefix = "local_file:///"
	if strings.HasPrefix(uri, prefix) {
		return strings.TrimPrefix(uri, prefix), true
	}
	alt := "local_file://"
	if strings.HasPrefix(uri, alt) {
		return strings.TrimPrefix(uri, alt), true
	}
	return "", false
}

func validateIdempotencyKey(key string) error {
	if !idempotencyKeyRe.MatchString(key) {
		logger.Errorf("im idempotency key invalid len=%d", len(key))
		return bizerror.ErrorBadMessage
	}
	return nil
}

func validateVisibility(v apicommon.Visibility) error {
	switch v {
	case apicommon.Visibility_VISIBILITY_PEER_ONLY,
		apicommon.Visibility_VISIBILITY_SENDER_ONLY,
		apicommon.Visibility_VISIBILITY_BOTH:
		return nil
	default:
		logger.Errorf("im visibility invalid v=%d", v)
		return bizerror.ErrorBadMessage
	}
}

func visibleFor(userID, senderID, peerID int64, v apicommon.Visibility) bool {
	switch v {
	case apicommon.Visibility_VISIBILITY_BOTH:
		return userID == senderID || userID == peerID
	case apicommon.Visibility_VISIBILITY_PEER_ONLY:
		return userID == peerID
	case apicommon.Visibility_VISIBILITY_SENDER_ONLY:
		return userID == senderID
	default:
		return false
	}
}
