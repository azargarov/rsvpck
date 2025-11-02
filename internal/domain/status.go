package domain

type Status int

const (
	StatusUnknown Status = iota
	StatusSkipped
	StatusFail
	StatusPass
	StatusWarning
	StatusTimeout
	StatusConnectionRefused
	StatusInvalid
	StatusInvalidCommand
	StatusDNSFailure
	StatusHTTPError
	StatusProxyAuth
)

func (s Status) IsValid() bool {
	switch s {
	case StatusUnknown, StatusSkipped, StatusFail, StatusPass, StatusWarning, StatusTimeout,
		StatusConnectionRefused, StatusInvalid, StatusDNSFailure, StatusHTTPError, StatusProxyAuth:
		return true
	}
	return false
}

func (s Status) String() string {
	if !s.IsValid() {
		return "Not valid"
	}
	switch s {
	case StatusUnknown:
		return "Unknown"
	case StatusSkipped:
		return "Skipped"
	case StatusFail:
		return "Failed"
	case StatusPass:
		return "Passed"
	case StatusWarning:
		return "Warning"
	case StatusTimeout:
		return "Timeout"
	case StatusConnectionRefused:
		return "Connection Refused"
	case StatusInvalid:
		return "Invalid"
	case StatusInvalidCommand:
		return "Command is invalid"
	case StatusDNSFailure:
		return "DNS Failure"
	case StatusHTTPError:
		return "HTTP Error"
	case StatusProxyAuth:
		return "Proxy Auth error"
	}
	return "Undefined"
}

func (s Status) IsSuccess() bool {
	return s == StatusPass || s == StatusWarning
}

func (s Status) IsTerminal() bool {
	return s != StatusUnknown
}
