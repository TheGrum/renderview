// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

// +build example

package main

import (
	"bytes"
	"strings"
)

func Calculate(lsystem string, depth int) string {
	// Parse the rules
	rules := strings.Split(lsystem, "\n")
	if len(rules) < 1 {
		return ""
	}

	// First line is starting state
	start := rules[0]
	rules = rules[1:]

	if len(rules) < 1 {
		// no rules, so state is unchanged
		return start
	}

	elements := make([]string, len(rules), len(rules))
	elementrules := make([]string, len(rules), len(rules))
	// → \u2192
	for i, k := range rules {
		k = strings.Replace(k, "\u2192", "=", -1)
		parts := strings.Split(k, "=")
		elements[i] = parts[0]
		if len(parts) > 1 {
			elementrules[i] = parts[1]
		} else {
			elementrules[i] = ""
		}
	}

	result := start
	for i := 0; i < depth; i++ {
		result = CalculateSinglePass(result, elements, elementrules)
	}
	return result
}

func CalculateSinglePass(state string, elements []string, elementrules []string) string {
	buffer := bytes.NewBuffer(make([]byte, 0, len(state)*100))

	for _, k := range state {
		s := string(k)
		for i := 0; i < len(elements); i++ {
			if s == elements[i] {
				buffer.WriteString(elementrules[i])
				s = ""
				break
			}
		}
		if s != "" {
			buffer.WriteString(s)
		}
	}

	return buffer.String()
}
