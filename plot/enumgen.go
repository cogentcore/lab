// Code generated by "core generate -add-types"; DO NOT EDIT.

package plot

import (
	"cogentcore.org/core/enums"
)

var _AxisScalesValues = []AxisScales{0, 1, 2, 3}

// AxisScalesN is the highest valid value for type AxisScales, plus one.
const AxisScalesN AxisScales = 4

var _AxisScalesValueMap = map[string]AxisScales{`Linear`: 0, `Log`: 1, `InverseLinear`: 2, `InverseLog`: 3}

var _AxisScalesDescMap = map[AxisScales]string{0: `Linear is a linear axis scale.`, 1: `Log is a Logarithmic axis scale.`, 2: `InverseLinear is an inverted linear axis scale.`, 3: `InverseLog is an inverted log axis scale.`}

var _AxisScalesMap = map[AxisScales]string{0: `Linear`, 1: `Log`, 2: `InverseLinear`, 3: `InverseLog`}

// String returns the string representation of this AxisScales value.
func (i AxisScales) String() string { return enums.String(i, _AxisScalesMap) }

// SetString sets the AxisScales value from its string representation,
// and returns an error if the string is invalid.
func (i *AxisScales) SetString(s string) error {
	return enums.SetString(i, s, _AxisScalesValueMap, "AxisScales")
}

// Int64 returns the AxisScales value as an int64.
func (i AxisScales) Int64() int64 { return int64(i) }

// SetInt64 sets the AxisScales value from an int64.
func (i *AxisScales) SetInt64(in int64) { *i = AxisScales(in) }

// Desc returns the description of the AxisScales value.
func (i AxisScales) Desc() string { return enums.Desc(i, _AxisScalesDescMap) }

// AxisScalesValues returns all possible values for the type AxisScales.
func AxisScalesValues() []AxisScales { return _AxisScalesValues }

// Values returns all possible values for the type AxisScales.
func (i AxisScales) Values() []enums.Enum { return enums.Values(_AxisScalesValues) }

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i AxisScales) MarshalText() ([]byte, error) { return []byte(i.String()), nil }

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *AxisScales) UnmarshalText(text []byte) error {
	return enums.UnmarshalText(i, text, "AxisScales")
}

var _RolesValues = []Roles{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

// RolesN is the highest valid value for type Roles, plus one.
const RolesN Roles = 13

var _RolesValueMap = map[string]Roles{`NoRole`: 0, `X`: 1, `Y`: 2, `Z`: 3, `U`: 4, `V`: 5, `W`: 6, `Low`: 7, `High`: 8, `Size`: 9, `Color`: 10, `Label`: 11, `Group`: 12}

var _RolesDescMap = map[Roles]string{0: `NoRole is the default no-role specified case.`, 1: `X axis`, 2: `Y axis`, 3: `Z axis`, 4: `U is the X component of a vector or first quartile in Box plot, etc.`, 5: `V is the Y component of a vector or third quartile in a Box plot, etc.`, 6: `W is the Z component of a vector`, 7: `Low is a lower error bar or region.`, 8: `High is an upper error bar or region.`, 9: `Size controls the size of points etc.`, 10: `Color controls the color of points or other elements.`, 11: `Label renders a label, typically from string data, but can also be used for values.`, 12: `Group is a special role for table-based plots. The unique values of this data are used to split the other plot data into groups, with each group added to the legend. A different default color will be used for each such group.`}

var _RolesMap = map[Roles]string{0: `NoRole`, 1: `X`, 2: `Y`, 3: `Z`, 4: `U`, 5: `V`, 6: `W`, 7: `Low`, 8: `High`, 9: `Size`, 10: `Color`, 11: `Label`, 12: `Group`}

// String returns the string representation of this Roles value.
func (i Roles) String() string { return enums.String(i, _RolesMap) }

// SetString sets the Roles value from its string representation,
// and returns an error if the string is invalid.
func (i *Roles) SetString(s string) error { return enums.SetString(i, s, _RolesValueMap, "Roles") }

// Int64 returns the Roles value as an int64.
func (i Roles) Int64() int64 { return int64(i) }

// SetInt64 sets the Roles value from an int64.
func (i *Roles) SetInt64(in int64) { *i = Roles(in) }

// Desc returns the description of the Roles value.
func (i Roles) Desc() string { return enums.Desc(i, _RolesDescMap) }

// RolesValues returns all possible values for the type Roles.
func RolesValues() []Roles { return _RolesValues }

// Values returns all possible values for the type Roles.
func (i Roles) Values() []enums.Enum { return enums.Values(_RolesValues) }

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i Roles) MarshalText() ([]byte, error) { return []byte(i.String()), nil }

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *Roles) UnmarshalText(text []byte) error { return enums.UnmarshalText(i, text, "Roles") }

var _StepKindValues = []StepKind{0, 1, 2, 3}

// StepKindN is the highest valid value for type StepKind, plus one.
const StepKindN StepKind = 4

var _StepKindValueMap = map[string]StepKind{`NoStep`: 0, `PreStep`: 1, `MidStep`: 2, `PostStep`: 3}

var _StepKindDescMap = map[StepKind]string{0: `NoStep connects two points by simple line.`, 1: `PreStep connects two points by following lines: vertical, horizontal.`, 2: `MidStep connects two points by following lines: horizontal, vertical, horizontal. Vertical line is placed in the middle of the interval.`, 3: `PostStep connects two points by following lines: horizontal, vertical.`}

var _StepKindMap = map[StepKind]string{0: `NoStep`, 1: `PreStep`, 2: `MidStep`, 3: `PostStep`}

// String returns the string representation of this StepKind value.
func (i StepKind) String() string { return enums.String(i, _StepKindMap) }

// SetString sets the StepKind value from its string representation,
// and returns an error if the string is invalid.
func (i *StepKind) SetString(s string) error {
	return enums.SetString(i, s, _StepKindValueMap, "StepKind")
}

// Int64 returns the StepKind value as an int64.
func (i StepKind) Int64() int64 { return int64(i) }

// SetInt64 sets the StepKind value from an int64.
func (i *StepKind) SetInt64(in int64) { *i = StepKind(in) }

// Desc returns the description of the StepKind value.
func (i StepKind) Desc() string { return enums.Desc(i, _StepKindDescMap) }

// StepKindValues returns all possible values for the type StepKind.
func StepKindValues() []StepKind { return _StepKindValues }

// Values returns all possible values for the type StepKind.
func (i StepKind) Values() []enums.Enum { return enums.Values(_StepKindValues) }

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i StepKind) MarshalText() ([]byte, error) { return []byte(i.String()), nil }

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *StepKind) UnmarshalText(text []byte) error { return enums.UnmarshalText(i, text, "StepKind") }

var _ShapesValues = []Shapes{0, 1, 2, 3, 4, 5, 6, 7}

// ShapesN is the highest valid value for type Shapes, plus one.
const ShapesN Shapes = 8

var _ShapesValueMap = map[string]Shapes{`Ring`: 0, `Circle`: 1, `Square`: 2, `Box`: 3, `Triangle`: 4, `Pyramid`: 5, `Plus`: 6, `Cross`: 7}

var _ShapesDescMap = map[Shapes]string{0: `Ring is the outline of a circle`, 1: `Circle is a solid circle`, 2: `Square is the outline of a square`, 3: `Box is a filled square`, 4: `Triangle is the outline of a triangle`, 5: `Pyramid is a filled triangle`, 6: `Plus is a plus sign`, 7: `Cross is a big X`}

var _ShapesMap = map[Shapes]string{0: `Ring`, 1: `Circle`, 2: `Square`, 3: `Box`, 4: `Triangle`, 5: `Pyramid`, 6: `Plus`, 7: `Cross`}

// String returns the string representation of this Shapes value.
func (i Shapes) String() string { return enums.String(i, _ShapesMap) }

// SetString sets the Shapes value from its string representation,
// and returns an error if the string is invalid.
func (i *Shapes) SetString(s string) error { return enums.SetString(i, s, _ShapesValueMap, "Shapes") }

// Int64 returns the Shapes value as an int64.
func (i Shapes) Int64() int64 { return int64(i) }

// SetInt64 sets the Shapes value from an int64.
func (i *Shapes) SetInt64(in int64) { *i = Shapes(in) }

// Desc returns the description of the Shapes value.
func (i Shapes) Desc() string { return enums.Desc(i, _ShapesDescMap) }

// ShapesValues returns all possible values for the type Shapes.
func ShapesValues() []Shapes { return _ShapesValues }

// Values returns all possible values for the type Shapes.
func (i Shapes) Values() []enums.Enum { return enums.Values(_ShapesValues) }

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i Shapes) MarshalText() ([]byte, error) { return []byte(i.String()), nil }

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *Shapes) UnmarshalText(text []byte) error { return enums.UnmarshalText(i, text, "Shapes") }

var _DefaultOffOnValues = []DefaultOffOn{0, 1, 2}

// DefaultOffOnN is the highest valid value for type DefaultOffOn, plus one.
const DefaultOffOnN DefaultOffOn = 3

var _DefaultOffOnValueMap = map[string]DefaultOffOn{`Default`: 0, `Off`: 1, `On`: 2}

var _DefaultOffOnDescMap = map[DefaultOffOn]string{0: `Default means use the default value.`, 1: `Off means to override the default and turn Off.`, 2: `On means to override the default and turn On.`}

var _DefaultOffOnMap = map[DefaultOffOn]string{0: `Default`, 1: `Off`, 2: `On`}

// String returns the string representation of this DefaultOffOn value.
func (i DefaultOffOn) String() string { return enums.String(i, _DefaultOffOnMap) }

// SetString sets the DefaultOffOn value from its string representation,
// and returns an error if the string is invalid.
func (i *DefaultOffOn) SetString(s string) error {
	return enums.SetString(i, s, _DefaultOffOnValueMap, "DefaultOffOn")
}

// Int64 returns the DefaultOffOn value as an int64.
func (i DefaultOffOn) Int64() int64 { return int64(i) }

// SetInt64 sets the DefaultOffOn value from an int64.
func (i *DefaultOffOn) SetInt64(in int64) { *i = DefaultOffOn(in) }

// Desc returns the description of the DefaultOffOn value.
func (i DefaultOffOn) Desc() string { return enums.Desc(i, _DefaultOffOnDescMap) }

// DefaultOffOnValues returns all possible values for the type DefaultOffOn.
func DefaultOffOnValues() []DefaultOffOn { return _DefaultOffOnValues }

// Values returns all possible values for the type DefaultOffOn.
func (i DefaultOffOn) Values() []enums.Enum { return enums.Values(_DefaultOffOnValues) }

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i DefaultOffOn) MarshalText() ([]byte, error) { return []byte(i.String()), nil }

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *DefaultOffOn) UnmarshalText(text []byte) error {
	return enums.UnmarshalText(i, text, "DefaultOffOn")
}
