package usercontroller

import "strings"

// AvatarResolver 将库内逻辑 URI 转为可访问 URL/路径（由宿主注入 mediastore）。
type AvatarResolver func(uri string) string

// AvatarNormalizer 写库前将公开 URL 还原为逻辑 URI。
type AvatarNormalizer func(raw string) string

var (
	resolveAvatar   AvatarResolver
	normalizeAvatar AvatarNormalizer
)

// SetAvatarResolver 设置头像出站 Resolve；未设置则原样返回库内值。
func SetAvatarResolver(fn AvatarResolver) {
	resolveAvatar = fn
}

// SetAvatarNormalizer 设置头像入站 Normalize；未设置则原样写入。
func SetAvatarNormalizer(fn AvatarNormalizer) {
	normalizeAvatar = fn
}

func avatarOut(uri string) string {
	if uri == "" {
		return ""
	}
	if resolveAvatar != nil {
		return resolveAvatar(uri)
	}
	return uri
}

func avatarIn(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if normalizeAvatar != nil {
		return normalizeAvatar(raw)
	}
	return raw
}
