package types

import (
	"strings"

	"gopkg.in/yaml.v2"
)

// String implements fmt.Stringer interface
func (d Denom) String() string {
	out, _ := yaml.Marshal(d)
	return string(out)
}

// Equal implements equal interface
func (d Denom) Equal(d1 *Denom) bool {
	return d.Name == d1.Name
}

// DenomList represents an array of Denom elements
type DenomList []Denom

// String implements fmt.Stringer interface for
func (dl DenomList) String() (out string) {
	for _, denom := range dl {
		out += denom.String() + "\n"
	}
	return strings.TrimSpace(out)
}

// Contains iterates the denomList and return true if the demon is placed on the list
func (dl DenomList) Contains(denom string) bool {
	for _, d := range dl {
		if d.Name == denom {
			return true
		}
	}
	return false
}
