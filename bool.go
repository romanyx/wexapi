package wexapi

import (
	"fmt"
	"strings"
)

type convertibleBool bool

func (bl *convertibleBool) UnmarshalJSON(data []byte) error {
	asString := string(data)
	asString = strings.Trim(asString, `"`)
	if asString == "1" || asString == "true" {
		*bl = true
	} else if asString == "0" || asString == "false" {
		*bl = false
	} else {
		return fmt.Errorf("boolean unmarshal error: invalid input %s", asString)
	}
	return nil
}

func (bl convertibleBool) String() string {
	return fmt.Sprintf("%t", bool(bl))
}
