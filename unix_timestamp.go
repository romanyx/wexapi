package wexapi

import (
	"fmt"
	"strconv"
	"time"
)

type unixTimestamp time.Time

func (ut *unixTimestamp) UnmarshalJSON(data []byte) error {
	i, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*ut = unixTimestamp(time.Unix(i, 0))
	return nil
}

func (ut unixTimestamp) String() string {
	return fmt.Sprintf("%s", time.Time(ut))
}
