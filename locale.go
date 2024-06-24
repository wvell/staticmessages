package staticmessages

import "context"

var (
	localeKey = ctxKey("locale")
)

// WrapLocale sets the locale in the ctx.
func WrapLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeKey, locale)
}

// GetLocale returns the locale from the ctx.
func GetLocale(ctx context.Context) string {
	l, ok := ctx.Value(localeKey).(string)
	if ok {
		return l
	}

	return ""
}

type ctxKey string
