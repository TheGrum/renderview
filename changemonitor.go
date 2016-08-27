// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package renderview

import "sync"

// ChangeMonitor implements comparing the value
// of a subset of parameters to determine whether
// a redraw or recalculation is required
type ChangeMonitor struct {
	lastValue string

	sync.Mutex
	Params []RenderParameter
}

func NewChangeMonitor() *ChangeMonitor {
	return &ChangeMonitor{
		lastValue: "",
		Params:    make([]RenderParameter, 0, 10),
	}
}

func (c *ChangeMonitor) AddParameters(params ...RenderParameter) {
	c.Lock()
	defer c.Unlock()

	c.Params = append(c.Params, params...)
}

func (c *ChangeMonitor) HasChanged() bool {
	c.Lock()
	defer c.Unlock()

	newValue := GetParameterStatusString(c.Params...)
	//fmt.Printf("Comparing (%v) to (%v)\n", c.lastValue, newValue)
	if c.lastValue != newValue {
		c.lastValue = newValue
		return true
	}
	return false
}
