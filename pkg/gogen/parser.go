package gogen

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/okieraised/rclgo/internal/utilities"
)

var goKeywords = map[string]struct{}{
	"break": {}, "default": {}, "func": {}, "interface": {}, "select": {},
	"case": {}, "defer": {}, "go": {}, "map": {}, "struct": {},
	"chan": {}, "else": {}, "goto": {}, "package": {}, "switch": {},
	"const": {}, "fallthrough": {}, "if": {}, "range": {}, "type": {},
	"continue": {}, "for": {}, "import": {}, "return": {}, "var": {},
}

func cName(rosName string) string {
	if _, ok := goKeywords[rosName]; ok {
		return "_" + rosName
	}
	return rosName
}

type parser struct {
	config *Config
	// Collect pre-field comments here to be included in the comments. Flushed
	// on empty lines.
	ros2messagesCommentsBuffer strings.Builder
}

func ParseMessage(config *Config, content string) (*ROS2Message, error) {
	p := &parser{config: config}
	msg := ROS2MessageNew("", "")
	if err := p.ParseROS2Message(msg, content); err != nil {
		return nil, err
	}
	return msg, nil
}

// ParseROS2Message parses a message definition.
func (p *parser) ParseROS2Message(res *ROS2Message, content string) error {
	return p.parseSections(content, res)
}

func (p *parser) ParseService(service *ROS2Service, source string) error {
	return p.parseSections(source, service.Request, service.Response)
}

func (p *parser) ParseAction(action *ROS2Action, source string) error {
	err := p.parseSections(source, action.Goal, action.Result, action.Feedback)
	if err != nil {
		return err
	}
	p.addImport(action.SendGoal.Request, "unique_identifier_msgs")
	action.SendGoal.Request.Fields = []*ROS2Field{
		p.goalIDField(),
		p.actionLocalField("goal", "Goal", action, action.Goal),
	}
	p.addImport(action.SendGoal.Response, "builtin_interfaces")
	action.SendGoal.Response.Fields = []*ROS2Field{
		p.primitiveField("accepted", "Accepted", "bool", "bool"),
		{
			RosName: "stamp",
			CName:   "stamp",
			GoName:  "Stamp",

			PkgName:   "builtin_interfaces",
			GoPkgName: "builtin_interfaces_msg",

			RosType: "Time",
			CType:   "Time",
			GoType:  "Time",
		},
	}
	p.addImport(action.GetResult.Request, "unique_identifier_msgs")
	action.GetResult.Request.Fields = []*ROS2Field{p.goalIDField()}
	action.GetResult.Response.Fields = []*ROS2Field{
		p.primitiveField("status", "Status", "int8_t", "int8"),
		p.actionLocalField("result", "Result", action, action.Result),
	}
	p.addImport(action.FeedbackMessage, "unique_identifier_msgs")
	action.FeedbackMessage.Fields = []*ROS2Field{
		p.goalIDField(),
		p.actionLocalField("feedback", "Feedback", action, action.Feedback),
	}
	return nil
}

func (p *parser) goalIDField() *ROS2Field {
	return &ROS2Field{
		RosName: "goal_id",
		CName:   "goal_id",
		GoName:  "GoalID",

		PkgName:   "unique_identifier_msgs",
		GoPkgName: "unique_identifier_msgs_msg",

		RosType: "UUID",
		CType:   "UUID",
		GoType:  "UUID",
	}
}

func (p *parser) primitiveField(cname, goname, ctype, gotype string) *ROS2Field {
	return &ROS2Field{
		RosName: cname,
		CName:   cname,
		GoName:  goname,

		RosType: gotype,
		CType:   ctype,
		GoType:  gotype,
	}
}

func (p *parser) actionLocalField(cname, goname string, action *ROS2Action, msg *ROS2Message) *ROS2Field {
	return &ROS2Field{
		RosName: cname,
		CName:   cname,
		GoName:  goname,

		PkgName:    action.Package,
		GoPkgName:  action.GoPackage(),
		IsPkgLocal: true,

		RosType: msg.Name,
		CType:   msg.Name,
		GoType:  msg.Name,
	}
}

func (p *parser) addImportSpecial(msg *ROS2Message, cPkg, goPkg, goImport string) {
	if goImport == "" {
		goImport = goPkg
	}
	if msg.GoImports[goImport] == "" {
		msg.GoImports[goImport] = goPkg
		msg.CImports.Add(cPkg)
	}
}

func (p *parser) addImport(msg *ROS2Message, pkg string) {
	goImport := p.config.MessageModulePrefix + "/" + pkg + "/msg"
	p.addImportSpecial(msg, pkg, pkg+"_msg", goImport)
}

func (p *parser) parseSections(source string, sections ...*ROS2Message) error {
	current := 0
	for i, line := range strings.Split(source, "\n") {
		line = strings.TrimSpace(line)
		if line == "---" {
			if current >= len(sections) {
				return errors.New("too many sections")
			}
			current++
		} else if err := p.parseLine(sections[current], line); err != nil {
			return fmt.Errorf("error on line %d: %w", i+1, err)
		}
	}
	return nil
}

func (p *parser) parseLine(msg *ROS2Message, line string) error {
	obj, err := p.parseMessageLine(line, msg)
	if err != nil {
		return err
	}

	switch obj := obj.(type) {
	case *ROS2Constant:
		msg.Constants = append(msg.Constants, obj)
	case *ROS2Field:
		msg.Fields = append(msg.Fields, obj)
		switch obj.PkgName {
		case "":
		case ".":
		case "time":
			msg.GoImports["time"] = ""
		case "primitives":
			msg.GoImports[p.config.RclgoImportPath+"/pkg/rclgo/"+obj.PkgName] = obj.GoPkgName
		default:
			msg.GoImports[p.config.MessageModulePrefix+"/"+obj.PkgName+"/msg"] = obj.GoPkgName
			msg.CImports.Add(obj.PkgName)
		}
	case nil:
	default:
		return fmt.Errorf("couldn't parse the input row '%s'", line)
	}
	return nil
}

var msgCommentRE = regexp.MustCompile(`^#\s*(.*)$`)

func (p *parser) parseMessageLine(testRow string, ros2msg *ROS2Message) (interface{}, error) {
	// Comment line: "# ...". Accumulate comment text and stop.
	if m := msgCommentRE.FindStringSubmatch(testRow); m != nil {
		if m[1] != "" {
			p.ros2messagesCommentsBuffer.WriteString(m[1])
		}
		return nil, nil
	}

	// Empty line: flush the accumulated comment buffer.
	if testRow == "" {
		p.ros2messagesCommentsBuffer.Reset()
		return nil, nil
	}

	typeChar, capture := isRowConstantOrField(testRow)
	switch typeChar {
	case 'c':
		if con, err := p.ParseROS2MessageConstant(capture); err == nil {
			return con, nil
		}
	case 'f':
		if f, err := p.ParseROS2MessageField(capture, ros2msg); err == nil {
			return f, nil
		}
	}

	return nil, fmt.Errorf("couldn't parse the input row as either ROS2 Field or Constant? input %q", testRow)
}

// Constant row, e.g.:
//
//	pkg/Type[<=N] NAME = VALUE  # comment
var reConstRow = regexp.MustCompile(
	`^(?:(?P<package>\w+)/)?(?P<type>\w+)(?P<array>\[(?P<bounded><=)?(?P<size>\d*)\])?\s+(?P<field>\w+)\s*=\s*(?P<default>[^#]+)?(?:\s*#\s*(?P<comment>.*))?$`,
)

// Field row, e.g. (incl. bounded strings):
//
//	pkg/Type<=N[<=M] name  default  # comment
var reFieldRow = regexp.MustCompile(
	`^(?:(?P<package>\w+)/)?(?P<type>\w+)(?P<boundedString><=\d*)?(?P<array>\[(?P<bounded><=)?(?P<size>\d*)\])?\s+(?P<field>\w+)\s*(?P<default>[^#]+)?(?:\s+#\s*(?P<comment>.*))?$`,
)

// helper: build a map[name]value from a named-group match
func findNamed(re *regexp.Regexp, s string) (map[string]string, bool) {
	m := re.FindStringSubmatch(s)
	if m == nil {
		return nil, false
	}
	names := re.SubexpNames()
	out := make(map[string]string, len(m)-1)
	for i := 1; i < len(m); i++ {
		if names[i] != "" {
			out[names[i]] = m[i]
		}
	}
	return out, true
}

func isRowConstantOrField(textRow string) (byte, map[string]string) {
	if z, ok := findNamed(reConstRow, textRow); ok {
		return 'c', z
	}
	if z, ok := findNamed(reFieldRow, textRow); ok {
		return 'f', z
	}
	return 'e', nil
}

func (p *parser) ParseROS2MessageConstant(capture map[string]string) (*ROS2Constant, error) {
	d := &ROS2Constant{
		RosType: capture["type"],
		RosName: capture["field"],
		Value:   strings.TrimSpace(capture["default"]),
		Comment: utilities.CommentSerializer(capture["comment"], &p.ros2messagesCommentsBuffer),
	}

	t, ok := primitiveTypeMappings[d.RosType]
	if !ok {
		d.GoType = fmt.Sprintf("<MISSING translation from ROS2 Constant type '%s'>", d.RosType)
		return d, fmt.Errorf("Unknown ROS2 Constant type '%s'\n", d.RosType)
	}
	d.GoType = t.GoType
	d.PkgName = t.PackageName
	return d, nil
}

func (p *parser) ParseROS2MessageField(capture map[string]string, ros2msg *ROS2Message) (*ROS2Field, error) {
	size, err := strconv.ParseInt(capture["size"], 10, 32)
	if err != nil && capture["size"] != "" {
		return nil, err
	}
	if capture["boundedString"] != "" &&
		!(capture["package"] == "" && capture["type"] == "string") {
		return nil, errors.New("the only base type that supports an upper boundary is string")
	}
	if capture["bounded"] != "" {
		capture["array"] = strings.Replace(capture["array"], capture["bounded"]+capture["size"], "", 1)
		capture["bounded"] += capture["size"]
		size = 0
	}
	f := &ROS2Field{
		Comment:      utilities.CommentSerializer(capture["comment"], &p.ros2messagesCommentsBuffer),
		GoName:       utilities.SnakeToCamel(capture["field"]),
		RosName:      capture["field"],
		CName:        cName(capture["field"]),
		RosType:      capture["type"],
		TypeArray:    capture["array"],
		ArrayBounded: capture["bounded"],
		ArraySize:    int(size),
		DefaultValue: capture["default"],
		PkgName:      capture["package"],
	}

	f.PkgName, f.CType, f.GoType = translateROS2Type(f, ros2msg)
	f.GoPkgName = f.PkgName
	switch f.PkgName {
	case "", "time", "primitives":
	case ".":
		if ros2msg.Type == "msg" {
			f.IsPkgLocal = true
		} else {
			f.PkgName = ros2msg.Package
			f.GoPkgName = ros2msg.Package + "_msg"
		}
	default:
		f.GoPkgName = f.PkgName + "_msg"
	}
	// Prepopulate extra Go imports
	p.cSerializationCode(f, ros2msg)
	p.goSerializationCode(f, ros2msg)

	return f, nil
}

func translateROS2Type(f *ROS2Field, m *ROS2Message) (pkgName string, cType string, goType string) {
	t, ok := primitiveTypeMappings[f.RosType]
	if ok {
		f.RosType = t.RosType
		return t.PackageName, t.CType, t.GoType
	}

	if f.PkgName == "" && m.Type != "msg" {
		return m.Package, f.RosType, f.RosType
	}

	// explicit package
	if f.PkgName != "" {
		// type of same package
		if f.PkgName == m.Package {
			return ".", f.RosType, f.RosType
		}

		return f.PkgName, f.RosType, f.RosType
	}

	// implicit package, type of std_msgs
	if m.Package != "std_msgs" {
		switch f.RosType {
		case "Bool", "ColorRGBA",
			"Duration", "Empty", "Float32MultiArray", "Float32",
			"Float64MultiArray", "Float64", "Header", "Int8MultiArray",
			"Int8", "Int16MultiArray", "Int16", "Int32MultiArray", "Int32",
			"Int64MultiArray", "Int64", "MultiArrayDimension", "MultiarrayLayout",
			"String", "Time", "UInt8MultiArray", "UInt8", "UInt16MultiArray", "UInt16",
			"UInt32MultiArray", "UInt32", "UInt64MultiArray", "UInt64":
			return "std_msgs", f.RosType, f.RosType
		}
	}

	// These are not actually primitive types, but same-package complex types.
	return ".", f.RosType, f.RosType
}

func (p *parser) cSerializationCode(f *ROS2Field, m *ROS2Message) string {
	if f.TypeArray != "" && f.ArraySize > 0 && f.PkgName != "" && f.IsPkgLocal {
		// Complex value Array local package reference
		return utilities.UpperCaseFirst(f.RosType) + `__Array_to_C(mem.` + f.CName + `[:], m.` + f.GoName + `[:])`

	} else if f.TypeArray != "" && f.ArraySize > 0 && f.PkgName != "" && !f.IsPkgLocal {
		// Complex value Array remote package reference
		return `cSlice_` + f.RosName + ` := mem.` + f.CName + `[:]
	` + f.GoPkgReference() + utilities.UpperCaseFirst(f.RosType) + `__Array_to_C(*(*[]` + f.GoPkgReference() + `C` + utilities.UpperCaseFirst(f.RosType) + `)(unsafe.Pointer(&cSlice_` + f.RosName + `)), m.` + f.GoName + `[:])`
	} else if f.TypeArray != "" && f.ArraySize == 0 && f.PkgName != "" && f.IsPkgLocal {
		// Complex value Slice local package reference
		return utilities.UpperCaseFirst(f.RosType) + `__Sequence_to_C(&mem.` + f.CName + `, m.` + f.GoName + `)`

	} else if f.TypeArray != "" && f.ArraySize == 0 && f.PkgName != "" && !f.IsPkgLocal {
		// Complex value Slice remote package reference
		return f.GoPkgReference() + utilities.UpperCaseFirst(f.RosType) + `__Sequence_to_C((*` + f.GoPkgReference() + `C` + utilities.UpperCaseFirst(f.RosType) + `__Sequence)(unsafe.Pointer(&mem.` + f.CName + `)), m.` + f.GoName + `)`

	} else if f.TypeArray == "" && f.PkgName != "" {
		// Complex value single
		return f.GoPkgReference() + f.GoType + "TypeSupport.AsCStruct(unsafe.Pointer(&mem." + f.CName + "), &m." + f.GoName + ")"

	} else if f.TypeArray != "" && f.ArraySize > 0 && f.PkgName == "" {
		// Primitive value Array
		m.GoImports[p.config.RclgoImportPath+"/pkg/rclgo/primitives"] = "primitives"
		return `cSlice_` + f.RosName + ` := mem.` + f.CName + `[:]
	` + `primitives.` + utilities.UpperCaseFirst(f.RosType) + `__Array_to_C(*(*[]primitives.C` + utilities.UpperCaseFirst(f.RosType) + `)(unsafe.Pointer(&cSlice_` + f.RosName + `)), m.` + f.GoName + `[:])`

	} else if f.TypeArray != "" && f.ArraySize == 0 && f.PkgName == "" {
		// Primitive value Slice
		m.GoImports[p.config.RclgoImportPath+"/pkg/rclgo/primitives"] = "primitives"
		return `primitives.` + utilities.UpperCaseFirst(f.RosType) + `__Sequence_to_C((*primitives.C` + utilities.UpperCaseFirst(f.RosType) + `__Sequence)(unsafe.Pointer(&mem.` + f.CName + `)), m.` + f.GoName + `)`

	} else if f.TypeArray == "" && f.PkgName == "" {
		// Primitive value single

		// string and U16String are special cases because they have custom
		// serialization implementations but still use a non-generated type in
		// generated message fields.
		if f.RosType == "string" {
			m.GoImports[p.config.RclgoImportPath+"/pkg/rclgo/primitives"] = "primitives"
			return "primitives.StringAsCStruct(unsafe.Pointer(&mem." + f.CName + "), m." + f.GoName + ")"
		} else if f.RosType == "U16String" {
			m.GoImports[p.config.RclgoImportPath+"/pkg/rclgo/primitives"] = "primitives"
			return "primitives.U16StringAsCStruct(unsafe.Pointer(&mem." + f.CName + "), m." + f.GoName + ")"
		}
		return `mem.` + f.CName + ` = C.` + f.CType + `(m.` + f.GoName + `)`
	}
	return "//<MISSING cSerializationCode!!>"
}

func (p *parser) goSerializationCode(f *ROS2Field, m *ROS2Message) string {

	if f.TypeArray != "" && f.ArraySize > 0 && f.PkgName != "" && f.IsPkgLocal {
		// Complex value Array local package reference
		return utilities.UpperCaseFirst(f.RosType) + `__Array_to_Go(m.` + f.GoName + `[:], mem.` + f.CName + `[:])`

	} else if f.TypeArray != "" && f.ArraySize > 0 && f.PkgName != "" {
		// Complex value Array remote package reference
		return `cSlice_` + f.RosName + ` := mem.` + f.CName + `[:]
	` + f.GoPkgReference() + utilities.UpperCaseFirst(f.RosType) + `__Array_to_Go(m.` + f.GoName + `[:], *(*[]` + f.GoPkgReference() + `C` + utilities.UpperCaseFirst(f.RosType) + `)(unsafe.Pointer(&cSlice_` + f.RosName + `)))`

	} else if f.TypeArray != "" && f.ArraySize == 0 && f.PkgName != "" && f.IsPkgLocal {
		// Complex value Slice local package reference
		return utilities.UpperCaseFirst(f.RosType) + `__Sequence_to_Go(&m.` + f.GoName + `, mem.` + f.CName + `)`

	} else if f.TypeArray != "" && f.ArraySize == 0 && f.PkgName != "" && !f.IsPkgLocal {
		// Complex value Slice remote package reference
		return f.GoPkgReference() + utilities.UpperCaseFirst(f.RosType) + `__Sequence_to_Go(&m.` + f.GoName + `, *(*` + f.GoPkgReference() + `C` + utilities.UpperCaseFirst(f.RosType) + `__Sequence)(unsafe.Pointer(&mem.` + f.CName + `)))`

	} else if f.TypeArray == "" && f.PkgName != "" {
		// Complex value single
		return f.GoPkgReference() + f.GoType + "TypeSupport.AsGoStruct(&m." + f.GoName + ", unsafe.Pointer(&mem." + f.CName + "))"

	} else if f.TypeArray != "" && f.ArraySize > 0 && f.PkgName == "" {
		// Primitive value Array
		m.GoImports[p.config.RclgoImportPath+"/pkg/rclgo/primitives"] = "primitives"
		return `cSlice_` + f.RosName + ` := mem.` + f.CName + `[:]
	` + `primitives.` + utilities.UpperCaseFirst(f.RosType) + `__Array_to_Go(m.` + f.GoName + `[:], *(*[]primitives.C` + utilities.UpperCaseFirst(f.RosType) + `)(unsafe.Pointer(&cSlice_` + f.RosName + `)))`

	} else if f.TypeArray != "" && f.ArraySize == 0 && f.PkgName == "" {
		// Primitive value Slice
		m.GoImports[p.config.RclgoImportPath+"/pkg/rclgo/primitives"] = "primitives"
		return `primitives.` + utilities.UpperCaseFirst(f.RosType) + `__Sequence_to_Go(&m.` + f.GoName + `, *(*primitives.C` + utilities.UpperCaseFirst(f.RosType) + `__Sequence)(unsafe.Pointer(&mem.` + f.CName + `)))`

	} else if f.TypeArray == "" && f.PkgName == "" {
		// Primitive value single

		// string and U16String are special cases because they have custom
		// serialization implementations but still use a non-generated type in
		// generated message fields.
		if f.RosType == "string" {
			m.GoImports[p.config.RclgoImportPath+"/pkg/rclgo/primitives"] = "primitives"
			return "primitives.StringAsGoStruct(&m." + f.GoName + ", unsafe.Pointer(&mem." + f.CName + "))"
		} else if f.RosType == "U16String" {
			m.GoImports[p.config.RclgoImportPath+"/pkg/rclgo/primitives"] = "primitives"
			return "primitives.U16StringAsGoStruct(&m." + f.GoName + ", unsafe.Pointer(&mem." + f.CName + "))"
		}
		return `m.` + f.GoName + ` = ` + f.GoType + `(mem.` + f.CName + `)`

	}
	return "//<MISSING goSerializationCode!!>"
}

func defaultCode(f *ROS2Field) string {
	if f.PkgName != "" && f.TypeArray != "" && f.DefaultValue == "" {
		// Complex value array and slice common default
		if f.ArraySize == 0 {
			return "t." + f.GoName + " = nil"
		}
		return "for i := range t." + f.GoName + " {\n" +
			"\t\tt." + f.GoName + "[i].SetDefaults()\n" +
			"\t}"

	} else if f.PkgName != "" && f.TypeArray != "" {
		defaultValues := utilities.SplitMsgDefaultArrayValues(f.RosType, f.DefaultValue)
		sb := strings.Builder{}
		if f.ArraySize <= 0 && len(defaultValues) > 0 {
			_, _ = fmt.Fprint(&sb, `t.`, f.GoName, ` = make(`, f.TypeArray, f.GoPkgReference(), f.GoType, `, `, len(defaultValues), ")\n\t")
		}

		_, _ = fmt.Fprint(&sb,
			"for i := range t.", f.GoName, " {\n",
			"\t\tt.", f.GoName, "[i].SetDefaults()\n",
			"\t}",
		)
		return sb.String()

	} else if f.PkgName != "" && f.TypeArray == "" {
		return `t.` + f.GoName + ".SetDefaults()"

	} else if f.DefaultValue != "" && f.TypeArray != "" {
		defaultValues := utilities.SplitMsgDefaultArrayValues(f.RosType, f.DefaultValue)
		for i := range defaultValues {
			defaultValues[i] = utilities.DefaultValueSanitizer(f.RosType, defaultValues[i])
		}
		return `t.` + f.GoName + ` = ` + f.TypeArray + f.GoPkgReference() + f.GoType + `{` + strings.Join(defaultValues, ",") + `}`

	} else if f.DefaultValue == "" && f.TypeArray != "" {
		if f.ArraySize == 0 {
			return "t." + f.GoName + " = nil"
		}
		return fmt.Sprint("t.", f.GoName, " = [", f.ArraySize, "]", f.GoType, "{}")

	} else if f.DefaultValue != "" {
		return `t.` + f.GoName + ` = ` + utilities.DefaultValueSanitizer(f.RosType, f.DefaultValue)

	} else if f.DefaultValue == "" {
		return "t." + f.GoName + " = " + primitiveCommonDefault(f)
	}
	return "//<MISSING defaultCode!!>"
}

func primitiveCommonDefault(f *ROS2Field) string {
	switch f.RosType {
	case "string", "wstring", "U16String":
		return `""`
	case "bool":
		return "false"
	case "float32", "float64", "byte", "char", "int8", "int16",
		"int32", "int64", "uint8", "uint16", "uint32", "uint64":
		return "0"
	default:
		panic("common default value for ROS type " + f.RosType + " is not defined")
	}
}

func cloneCode(f *ROS2Field) string {
	if f.PkgName != "" && f.TypeArray != "" && f.ArraySize == 0 {
		return "if t." + f.GoName + " != nil {\n" +
			"\t\tc." + f.GoName + " = make([]" + f.GoPkgReference() + f.GoType + ", len(t." + f.GoName + "))\n" +
			"\t\t" + f.GoPkgReference() + "Clone" + f.GoType + "Slice(c." + f.GoName + ", t." + f.GoName + ")\n" +
			"\t}"
	} else if f.PkgName != "" && f.TypeArray != "" && f.ArraySize > 0 {
		return f.GoPkgReference() + "Clone" + f.GoType + "Slice(c." + f.GoName + "[:], t." + f.GoName + "[:])"
	} else if f.PkgName != "" && f.TypeArray == "" {
		return "c." + f.GoName + " = *t." + f.GoName + ".Clone()"
	} else if f.PkgName == "" && f.TypeArray != "" && f.ArraySize == 0 {
		return "if t." + f.GoName + " != nil {\n" +
			"\t\tc." + f.GoName + " = make([]" + f.GoType + ", len(t." + f.GoName + "))\n" +
			"\t\tcopy(c." + f.GoName + ", t." + f.GoName + ")\n" +
			"\t}"
	} else if f.PkgName == "" {
		// primitive value single and array
		return "c." + f.GoName + " = t." + f.GoName
	}
	return "//<MISSING cloneCode!!>"
}
