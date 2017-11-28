// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package testing

import (
	"bufio"
	"fmt"
	"strings"

	"gopkg.in/check.v1"
)

var MultilineEquals = &multilineEqualsChecker{
	&check.CheckerInfo{
		Name:   "MultilineEquals",
		Params: []string{"obtained", "expected"},
	},
}

type multilineEqualsChecker struct {
	*check.CheckerInfo
}

func (checker *multilineEqualsChecker) Check(params []interface{}, names []string) (result bool, error string) {
	if len(params) != 2 {
		return false, "Expecting 2 values to compare."
	}

	v1, ok := params[0].(string)
	if !ok {
		return false, "Both values need to be strings."
	}
	v2, ok := params[1].(string)
	if !ok {
		return false, "Both values need to be strings."
	}

	sc1 := bufio.NewScanner(strings.NewReader(v1))
	sc2 := bufio.NewScanner(strings.NewReader(v2))
	sc1.Split(bufio.ScanLines)
	sc2.Split(bufio.ScanLines)

	for line := 0; ; line++ {
		s1 := sc1.Scan()
		s2 := sc2.Scan()

		if s1 && !s2 {
			return false, fmt.Sprintf("unequal line count: %q has more lines than %q", names[0], names[1])
		}
		if s2 && !s1 {
			return false, fmt.Sprintf("unequal line count: %q has more lines than %q", names[1], names[0])
		}
		if !s1 && !s2 {
			break
		}
		v1 := sc1.Text()
		v2 := sc2.Text()

		if v1 != v2 {
			return false, fmt.Sprintf("line %d mismatch, %s vs %s:\n%q\n%q", line, names[0], names[1], v1, v2)
		}
	}
	return true, ""
}
