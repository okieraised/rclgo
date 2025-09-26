package gogen

import (
	"path/filepath"

	"github.com/okieraised/rclgo/internal/utilities"
)

type Metadata struct {
	Name, Package, Type string
}

func (m *Metadata) ImportPath() string {
	return filepath.Join(m.Package, m.Type)
}

func (m *Metadata) GoPackage() string {
	return m.Package + "_" + m.Type
}

// ROS2Message is a message definition. https://design.ros2.org/articles/legacy_interface_definition.html
// Use ROS2MessageNew() to initialize the struct
type ROS2Message struct {
	*Metadata
	Fields    []*ROS2Field
	Constants []*ROS2Constant
	GoImports map[string]string
	CImports  utilities.StringSet
}

func ROS2MessageNew(pkg, name string) *ROS2Message {
	return newMessageWithType(pkg, name, "msg")
}

func newMessageWithType(pkg, name, typ string) *ROS2Message {
	return &ROS2Message{
		Metadata: &Metadata{
			Name:    name,
			Package: pkg,
			Type:    typ,
		},
		GoImports: map[string]string{},
		CImports:  utilities.StringSet{},
	}
}

// ROS2Constant is a message definition.
type ROS2Constant struct {
	RosType    string
	GoType     string
	RosName    string
	Value      string
	Comment    string
	PkgName    string
	IsPkgLocal bool
}

func (t *ROS2Constant) GoPkgReference() string {
	if t.PkgName == "" || t.IsPkgLocal {
		return ""
	}
	return t.PkgName + "."
}

// ROS2Field is a message field.
type ROS2Field struct {
	TypeArray    string
	ArrayBounded string
	ArraySize    int
	DefaultValue string
	PkgName      string
	GoPkgName    string
	IsPkgLocal   bool
	RosType      string
	CType        string
	GoType       string
	RosName      string
	CName        string
	GoName       string
	Comment      string
}

func (t *ROS2Field) GoPkgReference() string {
	if t.PkgName == "" || t.IsPkgLocal {
		return ""
	}
	return t.GoPkgName + "."
}

func (t *ROS2Field) IsSingleComplex() bool {
	return t.TypeArray == "" && t.PkgName != ""
}

type ROS2Service struct {
	*Metadata
	Request  *ROS2Message
	Response *ROS2Message
}

func newServiceWithType(pkg, name, typ string) *ROS2Service {
	return &ROS2Service{
		Metadata: &Metadata{
			Name:    name,
			Package: pkg,
			Type:    typ,
		},
		Request:  newMessageWithType(pkg, name+"_Request", typ),
		Response: newMessageWithType(pkg, name+"_Response", typ),
	}
}

func NewROS2Service(pkg, name string) *ROS2Service {
	return newServiceWithType(pkg, name, "srv")
}

type ROS2Action struct {
	*Metadata
	Goal            *ROS2Message
	Result          *ROS2Message
	Feedback        *ROS2Message
	SendGoal        *ROS2Service
	GetResult       *ROS2Service
	FeedbackMessage *ROS2Message
}

func NewROS2Action(pkg, name string) *ROS2Action {
	return &ROS2Action{
		Metadata: &Metadata{
			Name:    name,
			Package: pkg,
			Type:    "action",
		},
		Goal:            newMessageWithType(pkg, name+"_Goal", "action"),
		Result:          newMessageWithType(pkg, name+"_Result", "action"),
		Feedback:        newMessageWithType(pkg, name+"_Feedback", "action"),
		SendGoal:        newServiceWithType(pkg, name+"_SendGoal", "action"),
		GetResult:       newServiceWithType(pkg, name+"_GetResult", "action"),
		FeedbackMessage: newMessageWithType(pkg, name+"_FeedbackMessage", "action"),
	}
}

// ROS2ErrorType must have fields exported otherwise they cannot be used by the
// test/template package
type ROS2ErrorType struct {
	Name      string
	RclRetT   string // The function call return value the error is mapped to
	Reference string // This is a reference to another type, so we just redefine the same type with another name
	Comment   string // Any found comments before or over the #definition
}

type rosIDLRuntimeCTypeMapping struct {
	RosType     string
	GoType      string
	CType       string
	CStructName string
	PackageName string
	SkipAutogen bool
}

var primitiveTypeMappings = map[string]rosIDLRuntimeCTypeMapping{
	"string":   {RosType: "string", GoType: "string", CStructName: "String", CType: "String", SkipAutogen: true},
	"time":     {RosType: "time", GoType: "Time", CStructName: "Time", CType: "time", SkipAutogen: true},
	"duration": {RosType: "duration", GoType: "Duration", CStructName: "Duration", CType: "duration", SkipAutogen: true},
	"float32":  {RosType: "float32", GoType: "float32", CStructName: "float", CType: "float"},
	"float64":  {RosType: "float64", GoType: "float64", CStructName: "double", CType: "double"},
	"bool":     {RosType: "bool", GoType: "bool", CStructName: "boolean", CType: "bool"},
	"byte":     {RosType: "byte", GoType: "byte", CStructName: "octet", CType: "uint8_t"},
	"char":     {RosType: "char", GoType: "byte", CStructName: "char", CType: "uchar", SkipAutogen: true},
	"int8":     {RosType: "int8", GoType: "int8", CStructName: "int8", CType: "int8_t"},
	"int16":    {RosType: "int16", GoType: "int16", CStructName: "int16", CType: "int16_t"},
	"int32":    {RosType: "int32", GoType: "int32", CStructName: "int32", CType: "int32_t"},
	"int64":    {RosType: "int64", GoType: "int64", CStructName: "int64", CType: "int64_t"},
	"uint8":    {RosType: "uint8", GoType: "uint8", CStructName: "uint8", CType: "uint8_t"},
	"uint16":   {RosType: "uint16", GoType: "uint16", CStructName: "uint16", CType: "uint16_t"},
	"uint32":   {RosType: "uint32", GoType: "uint32", CStructName: "uint32", CType: "uint32_t"},
	"uint64":   {RosType: "uint64", GoType: "uint64", CStructName: "uint64", CType: "uint64_t"},
	"wstring":  {RosType: "U16String", GoType: "string", CStructName: "U16String", CType: "U16String", SkipAutogen: true},
}

// blacklistedMessages is matched against the paths gogen inspects
// if it is a ROS2 Message file and needs to be turned into a Go type.
// If the path matches the blacklist, it is ignored and a notification is logged.
var blacklistedMessages = []string{
	"libstatistics_collector/msg/DummyMessage",
	"this-is-a-test-blacklist-entry-do-not-remove-used-for-internal-testing",
}

// cErrorTypeFiles are looked for #definitions and parsed as Golang ros2 error types
var cErrorTypeFiles = []string{
	"rcl/types.h",
	"rmw/ret_types.h",
	"rcl_action/types.h",
}
