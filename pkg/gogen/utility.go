package gogen

import (
	"slices"
	"strings"

	"github.com/okieraised/rclgo/internal/utilities"
	"golang.org/x/tools/go/packages"
)

func actionHasSuffix(msg *ROS2Message, suffixes ...string) bool {
	if msg.Type == "action" {
		for _, suffix := range suffixes {
			if strings.HasSuffix(msg.Name, suffix) {
				return true
			}
		}
	}
	return false
}

func matchMsg(msg *ROS2Message, pkg, name string) bool {
	return msg.GoPackage() == pkg && msg.Name == name
}

func loadGoPkgDeps(pkgPaths ...string) (utilities.StringSet, error) {
	deps := utilities.StringSet{}
	if len(pkgPaths) > 0 {
		queries := slices.Clone(pkgPaths)
		for i := range queries {
			queries[i] = "pattern=" + queries[i]
		}
		pkgs, err := packages.Load(&packages.Config{
			Mode:  packages.NeedImports | packages.NeedDeps | packages.NeedName,
			Tests: true,
		}, queries...)
		if err != nil {
			return nil, err
		}
		packages.Visit(pkgs, func(pkg *packages.Package) bool {
			deps.Add(pkg.PkgPath)
			return true
		}, nil)
	}
	return deps, nil
}
