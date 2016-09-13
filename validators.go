package cbflag

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
)

type HostNameError struct {
	msg string
}

func (e HostNameError) Error() string {
	return fmt.Sprintf("%s\n\nPlease specify a hostname using one of the following"+
		" patterns:\n\n* <addr>:<port>\n* http://<addr>:<port>\n* couchbase://<addr>", e.msg)
}

func HostValidator(value Value) error {
	// Valid hostname should be in the form:
	// <addr>:<port>
	// http://<addr>:<port>
	// couchbase://<addr>

	parsed, err := url.Parse(value.String())
	if err == nil && parsed.Scheme == "" {
		parsed, err = url.Parse("http://" + value.String())
	}

	if err != nil {
		return err
	}

	if parsed.Path != "" {
		return HostNameError{fmt.Sprintf("Host has path `%s` specified, but paths are not allowed",
			parsed.Path)}
	}

	if parsed.RawQuery != "" {
		return HostNameError{fmt.Sprintf("Host has query `%s` specified, but queries are not allowed",
			parsed.RawQuery)}
	}

	if parsed.User != nil {
		return HostNameError{fmt.Sprintf("Host has credentials `%s` specified, but credentials are not allowed",
			parsed.User.String())}
	}

	if parsed.Scheme == "" || parsed.Scheme == "http" || parsed.Scheme == "https" {
		_, port, _ := net.SplitHostPort(parsed.Host)
		if port == "" {
			if parsed.Scheme == "" || parsed.Scheme == "http" {
				port = "8091"
			} else {
				port = "18091"
			}
			parsed.Host = fmt.Sprintf("%s:%s", parsed.Host, port)
		}

		p, err := strconv.ParseUint(port, 10, 64)
		if err != nil {
			return HostNameError{fmt.Sprintf("Port specified `%s` is not a number", port)}
		}

		if p > 65535 {
			return HostNameError{fmt.Sprintf("Port specified `%s` is too big", port)}
		}

		if parsed.Scheme == "" {
			parsed.Scheme = "http"
		}
	} else if parsed.Scheme == "couchbase" || parsed.Scheme == "couchbases" {
		_, port, _ := net.SplitHostPort(parsed.Host)
		if port == "" && parsed.Scheme == "couchbase" {
			parsed.Host = fmt.Sprintf("%s:%d", parsed.Host, 8091)
		} else if port == "" && parsed.Scheme == "couchbases" {
			parsed.Host = fmt.Sprintf("%s:%d", parsed.Host, 18091)
		} else {
			p, err := strconv.ParseUint(port, 10, 64)
			if err != nil {
				return HostNameError{fmt.Sprintf("Port specified `%s` is not a number", port)}
			}

			if p > 65535 {
				return HostNameError{fmt.Sprintf("Port specified `%s` is too big", port)}
			}
		}

		if parsed.Scheme == "couchbase" {
			parsed.Scheme = "http"
		} else if parsed.Scheme == "couchbases" {
			parsed.Scheme = "https"
		}
	} else {
		return HostNameError{fmt.Sprintf("Invalid hostname, %s is not an accepted scheme", parsed.Scheme)}
	}

	value.Set(parsed.String())
	return nil
}
