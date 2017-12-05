package route

import (
	"context"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

type matchedDomainKey int

// OriginSecurity is a SecurityHandler that will ensure a request comes from a known origin. This
// can be an effective way to mitigate CSRF attacks, which are unable to forge headers due to the
// passive nature of the attack vector.
//
// OriginSecurity will store the matching domain in the http.Request's Context. Use MatchedDomain
// to retrieve the value in later logic.
func OriginSecurity(domains []Domain) SecurityHandler {
	logger := log.WithFields(log.Fields{"validDomains": domains})

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			url, err := url.Parse(origin)
			if err != nil {
				panic(err)
			}

			for _, d := range domains {
				if d.Matches(url) {
					ctx := r.Context()
					ctx = context.WithValue(ctx, matchedDomainKey(0), &d)

					h.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			logger.WithFields(log.Fields{"origin": origin}).Info("Origin validation failed")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Origin is not a trusted host."))
		})
	}
}

// MatchedDomain will retrieve from the http.Request's Context the domain that satisfied
// OriginSecurity.
func MatchedDomain(r *http.Request) *Domain {
	d, ok := r.Context().Value(matchedDomainKey(0)).(*Domain)
	if ok {
		return d
	}
	return nil
}
