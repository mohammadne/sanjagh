package failure

import (
	"fmt"
	"strings"
)

type Failure []string

func (r *Failure) RegisterReason(f string, p ...any) {
	*r = append(*r, fmt.Sprintf(f, p...))
}

func (r Failure) IsAllowed() bool {
	return len(r) == 0
}

func (r Failure) Reason() string {
	var reason string
	for _, response := range r {
		reason = reason + response + ","
	}
	reason = strings.TrimSuffix(reason, ",")
	return reason
}
