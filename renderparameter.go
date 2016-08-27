// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package renderview

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

const (
	HINT_NONE     = 1 << iota
	HINT_HIDE     = 1 << iota
	HINT_SIDEBAR  = 1 << iota
	HINT_FOOTER   = 1 << iota
	HINT_FULLTEXT = 1 << iota
)

type RenderParameter interface {
	GetName() string
	GetType() string
	GetHint() int
	SetHint(int)
	GetValueInt() int
	GetValueUInt32() uint32
	GetValueFloat64() float64
	GetValueComplex128() complex128
	GetValueString() string
	SetValueInt(value int) int
	SetValueUInt32(value uint32) uint32
	SetValueFloat64(value float64) float64
	SetValueComplex128(value complex128) complex128
	SetValueString(value string) string
}

type EmptyParameter struct {
	Name string
	Type string
	Hint int
}

func (e *EmptyParameter) GetName() string {
	return e.Name
}

func (e *EmptyParameter) GetType() string {
	return e.Type
}

func (e *EmptyParameter) GetHint() int {
	return e.Hint
}

func (e *EmptyParameter) GetValueInt() int {
	return 0
}

func (e *EmptyParameter) GetValueUInt32() uint32 {
	return 0
}
func (e *EmptyParameter) GetValueFloat64() float64 {
	return 0
}
func (e *EmptyParameter) GetValueString() string {
	return ""
}

func (e *EmptyParameter) GetValueComplex128() complex128 {
	return 0
}

func (e *EmptyParameter) SetHint(value int) {
	e.Hint = value
}
func (e *EmptyParameter) SetValueInt(value int) int {
	return 0
}
func (e *EmptyParameter) SetValueUInt32(value uint32) uint32 {
	return 0
}
func (e *EmptyParameter) SetValueFloat64(value float64) float64 {
	return 0
}
func (e *EmptyParameter) SetValueString(value string) string {
	return ""
}
func (e *EmptyParameter) SetValueComplex128(value complex128) complex128 {
	return 0
}

type UInt32RenderParameter struct {
	EmptyParameter

	Value uint32
}

func (e *UInt32RenderParameter) GetValueUInt32() uint32 {
	return e.Value
}

func (e *UInt32RenderParameter) SetValueUInt32(v uint32) uint32 {
	e.Value = v
	return e.Value
}

func (e *UInt32RenderParameter) GetValueString() string {
	return fmt.Sprintf("%v", e.Value)
}

func (e *UInt32RenderParameter) SetValueString(v string) string {
	r, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		e.SetValueUInt32(uint32(r))
	} else {
		e.SetValueUInt32(0)
	}
	return e.GetValueString()
}

type IntRenderParameter struct {
	EmptyParameter

	Value int
}

func (e *IntRenderParameter) GetValueInt() int {
	return e.Value
}

func (e *IntRenderParameter) SetValueInt(v int) int {
	e.Value = v
	return e.Value
}

type Float64RenderParameter struct {
	EmptyParameter

	Value float64
}

func (e *Float64RenderParameter) GetValueFloat64() float64 {
	return e.Value
}

func (e *Float64RenderParameter) SetValueFloat64(v float64) float64 {
	e.Value = v
	return e.Value
}

type Complex128RenderParameter struct {
	EmptyParameter

	Value complex128
}

func (e *Complex128RenderParameter) GetValueComplex128() complex128 {
	return e.Value
}

func (e *Complex128RenderParameter) SetValueComplex128(v complex128) complex128 {
	e.Value = v
	return e.Value
}

type StringRenderParameter struct {
	EmptyParameter

	Value string
}

func (e *StringRenderParameter) GetValueString() string {
	return e.Value
}

func (e *StringRenderParameter) SetValueString(v string) string {
	e.Value = v
	return e.Value
}

func NewUInt32RP(name string, value uint32) *UInt32RenderParameter {
	return &UInt32RenderParameter{
		EmptyParameter: EmptyParameter{
			Name: name,
			Type: "uint32",
		},
		Value: value,
	}
}

func NewIntRP(name string, value int) *IntRenderParameter {
	return &IntRenderParameter{
		EmptyParameter: EmptyParameter{
			Name: name,
			Type: "int",
		},
		Value: value,
	}
}

func NewFloat64RP(name string, value float64) *Float64RenderParameter {
	return &Float64RenderParameter{
		EmptyParameter: EmptyParameter{
			Name: name,
			Type: "float64",
		},
		Value: value,
	}
}

func NewComplex128RP(name string, value complex128) *Complex128RenderParameter {
	return &Complex128RenderParameter{
		EmptyParameter: EmptyParameter{
			Name: name,
			Type: "complex128",
		},
		Value: value,
	}
}

func NewStringRP(name string, value string) *StringRenderParameter {
	return &StringRenderParameter{
		EmptyParameter: EmptyParameter{
			Name: name,
			Type: "string",
		},
		Value: value,
	}
}

// Utility functions

// GetParameterValueAsString replaces the need to implement GetValueString on
// each parameter. Custom parameters should override GetValueString to override
// this behavior.
func GetParameterValueAsString(p RenderParameter) string {
	switch p.GetType() {
	case "int":
		return fmt.Sprintf("%v", p.GetValueInt())
	case "uint32":
		return fmt.Sprintf("%v", p.GetValueUInt32())
	case "float64":
		return fmt.Sprintf("%v", p.GetValueFloat64())
	case "complex128":
		return fmt.Sprintf("%v", p.GetValueComplex128())
	case "string":
		return p.GetValueString()
	default:
		return p.GetValueString()
	}
}

func ParseComplex(v string) (complex128, error) {
	v = strings.Replace(v, ",", "+", -1)
	l := strings.Split(v, "+")
	r, err := strconv.ParseFloat(l[0], 64)
	if err != nil {
		return 0, err
	}
	if len(l) > 1 {
		l[1] = strings.Replace(l[1], "i", "", -1)
		i, err := strconv.ParseFloat(l[1], 64)
		if err != nil {
			return 0, err
		}
		return complex(r, i), nil
	} else {
		return complex(r, 0), nil
	}

}

// SetParameterValueAsString replaces the need to implement SetValueString on
// each parameter. Custom parameters should override SetValueString to override
// this behavior.
func SetParameterValueFromString(p RenderParameter, v string) {
	switch p.GetType() {
	case "int":
		i, err := strconv.Atoi(v)
		if err == nil {
			p.SetValueInt(i)
		}
	case "uint32":
		i, err := strconv.ParseInt(v, 10, 32)
		if err == nil {
			p.SetValueUInt32(uint32(i))
		}
	case "float64":
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			p.SetValueFloat64(f)
		}
	case "complex128":
		c, err := ParseComplex(v)
		if err == nil {
			p.SetValueComplex128(c)
		}
	case "string":
		p.SetValueString(v)
	default:
		p.SetValueString(v)

	}
}

func SetHints(hint int, params ...RenderParameter) []RenderParameter {
	for _, p := range params {
		p.SetHint(hint)
	}

	return params
}

func GetParameterStatusString(params ...RenderParameter) string {
	buffer := bytes.NewBuffer(make([]byte, 0, len(params)*15))
	for i, p := range params {
		is := strconv.Itoa(i)
		buffer.WriteString(is)
		buffer.WriteString(":")
		buffer.WriteString(GetParameterValueAsString(p))
		buffer.WriteString(",")
	}
	return buffer.String()
}
