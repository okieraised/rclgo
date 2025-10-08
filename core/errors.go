package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/okieraised/rclgo/core/internal/distro"
	"github.com/okieraised/rclgo/core/internal/utilities"
)

var (
	errorTypesCFileRE *regexp.Regexp
	reErrDefine       = regexp.MustCompile(
		`^#define\s+(?P<name>(?:RCL|RMW)_RET_\w+)\s+(?:(?P<int>\d+)|(?P<reference>\w+))\s*(?://\s*(?P<comment>.+))?\s*$`,
	)
	reCommentLine = regexp.MustCompile(`^\s*//+\s*(.+)$`)
)

func init() { prepareErrorTypesCFileRegexp() }

// Build a single union-regex from cErrorTypeFiles.
// Each entry is treated as a regex, wrapped in a non-capturing group.
func prepareErrorTypesCFileRegexp() {
	if len(cErrorTypeFiles) == 0 {
		errorTypesCFileRE = nil
		return
	}
	union := "(?:" + strings.Join(cErrorTypeFiles, ")|(?:") + ")"
	re, err := regexp.Compile(union)
	if err != nil {
		panic(fmt.Errorf("invalid cErrorTypeFiles pattern: %w", err))
	}
	errorTypesCFileRE = re
}

func (g *Generator) GenerateROS2ErrorTypes() error {
	destFilePath := filepath.Join(g.config.DestPath, "error_types.gen.go")
	var errorTypes []*ROS2ErrorType

	for _, root := range g.config.RootPaths {
		includeLookupDir := root
		for tries := 0; tries < 10; tries++ {
			_, _ = fmt.Fprintf(os.Stderr, "Looking for rcl C include files to parse error definitions from '%s'\n", includeLookupDir)

			_ = filepath.Walk(includeLookupDir, func(path string, info os.FileInfo, err error) error {
				if err == nil && matchesErrorTypesCFile(path) {
					_, _ = fmt.Fprintf(os.Stderr, "Analyzing: %s\n", path)
					var convErr error
					errorTypes, convErr = generateGolangErrorTypesFromROS2ErrorDefinitionsPath(errorTypes, path)
					if convErr != nil {
						_, _ = fmt.Fprintf(os.Stderr, "Error converting ROS2 Errors from '%s' to '%s', error: %v\n", path, destFilePath, convErr)
					}
				}
				return nil
			})

			if len(errorTypes) == 0 {
				includeLookupDir = filepath.Join(includeLookupDir, "..")
				if includeLookupDir == "/" {
					break
				}
			} else {
				break
			}
		}
	}

	if len(errorTypes) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to find any rcl C error header files?\n")
		return nil
	}

	_, _ = fmt.Fprintf(os.Stderr, "Generating ROS2 Error definitions: %s\n", destFilePath)
	return g.generateGoFile(
		destFilePath,
		ros2ErrorCodes,
		templateData{
			"errorTypes":  errorTypes,
			"includes":    cErrorTypeFiles,
			"dedupFilter": ros2errorTypesDeduplicationFilter,
			"ROSDistro":   filepath.Base(os.Getenv(distro.AmentPrefixPath)),
		},
	)
}

func matchesErrorTypesCFile(path string) bool {
	if errorTypesCFileRE == nil {
		return false
	}
	return errorTypesCFileRE.MatchString(path)
}

func generateGolangErrorTypesFromROS2ErrorDefinitionsPath(errorTypes []*ROS2ErrorType, path string) ([]*ROS2ErrorType, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(content), "\n") {
		if et := parseROS2ErrorType(line); et != nil {
			errorTypes = append(errorTypes, et)
		}
	}
	return errorTypes, nil
}

var ros2errorTypesCommentsBuffer = strings.Builder{}                  // Collect pre-field comments here to be included in the comments. Flushed on empty lines.
var ros2errorTypesDeduplicationMap = make(map[string]string, 1024)    // Some RMW and RCL error codes overlap, so we need to deduplicate them from the dynamic type casting switch-case
var ros2errorTypesDeduplicationFilter = make(map[string]string, 1024) // Entries ending up here actually filter template entries

func parseROS2ErrorType(row string) *ROS2ErrorType {
	// Match a #define of RCL/RMW return codes.
	if m := reErrDefine.FindStringSubmatch(row); m != nil {
		idxName := reErrDefine.SubexpIndex("name")
		idxInt := reErrDefine.SubexpIndex("int")
		idxRef := reErrDefine.SubexpIndex("reference")
		idxCmt := reErrDefine.SubexpIndex("comment")

		et := &ROS2ErrorType{
			Name:      m[idxName],
			RclRetT:   m[idxInt], // empty if a symbolic reference was used
			Reference: m[idxRef],
			Comment:   utilities.CommentSerializer(m[idxCmt], &ros2errorTypesCommentsBuffer),
		}
		ros2errorTypesCommentsBuffer.Reset()
		updateROS2errorTypesDeduplicationMap(et.RclRetT, et.Name)
		return et
	}

	// Standalone comment line: accumulate into buffer.
	if cm := reCommentLine.FindStringSubmatch(row); cm != nil {
		if cm[1] != "" {
			ros2errorTypesCommentsBuffer.WriteString(cm[1])
		}
		return nil
	}

	// Empty line: flush buffer.
	if strings.TrimSpace(row) == "" {
		ros2errorTypesCommentsBuffer.Reset()
		return nil
	}

	return nil
}

func updateROS2errorTypesDeduplicationMap(rclRetT, name string) {
	if _, taken := ros2errorTypesDeduplicationMap[rclRetT]; taken {
		ros2errorTypesDeduplicationFilter[name] = rclRetT // duplicate value -> filter out later
		return
	}
	ros2errorTypesDeduplicationMap[rclRetT] = name
}
