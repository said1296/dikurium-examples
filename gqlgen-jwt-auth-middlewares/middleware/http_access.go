package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

type HttpAccess struct {
	Writer    http.ResponseWriter
	Request   *http.Request
	IP        string
	UserAgent string
}

const (
	authCookieName   = "Authorization"
	authExpiryName   = "Authorization-expiration"
	authCookiePrefix = "Bearer "
)

var (
	ctxKeyHttpAccess = &ContextKey{"httpAccess"}
)

func (c *HttpAccess) SetAuthorizationCookie(cookieValue string, expireIn time.Duration) {
	expiry := time.Now().Add(expireIn)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     authCookieName,
		Value:    authCookiePrefix + cookieValue,
		HttpOnly: true,
		Path:     "/",
		Expires:  expiry,
		Domain:   ".jevels.com",
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     authExpiryName,
		Value:    strconv.FormatInt(expiry.Unix(), 10),
		HttpOnly: false,
		Path:     "/",
		Expires:  expiry.Add((24 * time.Hour) * 3600),
		Domain:   ".jevels.com",
	})
}

// AuthMiddleware decodes the share session cookie and packs the session into context, also provides IP and User Agent
func HttpAccessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get IP and User Agent to read later
		ip := r.Header.Get("X-Forwarded-For")
		userAgent := r.UserAgent()

		// cookieAccess is a pointer so any changes in future is changing cookieAccess in context
		cookieAccess := &HttpAccess{
			Writer:    w,
			Request:   r,
			IP:        ip,
			UserAgent: userAgent,
		}
		ctx := context.WithValue(r.Context(), ctxKeyHttpAccess, cookieAccess)

		// and call the next with our new context
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})

}

func GetHttpAccess(ctx context.Context) *HttpAccess {
	cookieAccess, _ := ctx.Value(ctxKeyHttpAccess).(*HttpAccess)
	return cookieAccess
}
