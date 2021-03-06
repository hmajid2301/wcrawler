package core

import (
	"fmt"
	"net/url"
	"time"
)

// ----------
// Interfaces
// ----------

// Connector describes the connector interface
type Connector interface {
	GetLinks(rawURL string) (statusCode int, links []URLEntity, latency time.Duration, err error)
}

type StatsManager interface {
	UpdateStats(updates ...func(*StatsCLIOutput))
	RunOutputFlusher()
}

// ---------------
// Utils functions
// ---------------

// IsAbsoluteURL validates URL
func IsAbsoluteURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// ExtractParentURL takes any URL and returns a URL string with scheme,authority,path ready
// to be used as a parent URL.
func ExtractParentURL(rawURL string) (baseURL string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("URL not in a valid format: %s", err)
	}

	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("URL provided is not absolute")
	}

	baseURL = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
	return baseURL, nil
}

// ExtractURL behaves the same way as parent URL, except that it also includes query params.
// If URL provided is relative, it will join the URLs.
// It will return an error if URL is of an unwanted type, like 'mailto'.
func ExtractURL(baseURL string, rawURL string) (URLEntity, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return URLEntity{}, fmt.Errorf("URL not in a valid format: %s", err)
	}

	// Check if rawURL is one of those types we don't want to handle, i.e., mailto, telephone, etc.
	if u.Opaque != "" || (u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https") {
		return URLEntity{}, fmt.Errorf("URL not in a supported format")
	}

	baseU, err := url.Parse(baseURL)
	if err != nil {
		return URLEntity{}, fmt.Errorf("base URL not in a valid format: %s", err)
	}

	if baseU.Opaque != "" || (baseU.Scheme != "" && baseU.Scheme != "http" && baseU.Scheme != "https") {
		return URLEntity{}, fmt.Errorf("URL not in a supported format")
	}

	mergedU := baseU.ResolveReference(u)
	rawURL = fmt.Sprintf("%s://%s%s", mergedU.Scheme, mergedU.Host, mergedU.Path)

	if mergedU.RawQuery != "" {
		rawURL += "?" + mergedU.RawQuery
	}

	return URLEntity{Host: mergedU.Host, Raw: rawURL}, nil
}
