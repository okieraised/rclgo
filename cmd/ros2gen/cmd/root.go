package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/okieraised/rclgo/internal/distro"
	"github.com/okieraised/rclgo/pkg/gogen"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/tools/go/packages"
)

var rootCmd = &cobra.Command{
	Use:   "ros2gen",
	Short: "ROS2 client library in Golang - ROS2 Message generator",
	Long:  `Generate Go-compatible types for the ROS2 messages`,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
}

func validate(cmd *cobra.Command, _ []string) error {
	rootPaths := getRootPaths(cmd)
	if len(rootPaths) == 0 {
		if os.Getenv(distro.AmentPrefixPath) == "" {
			return fmt.Errorf("you haven't sourced your ROS2 environment! Source your ROS2 or pass --root-path")
		}
		return fmt.Errorf("root-path is required")
	}
	rosVersion := filepath.Base(os.Getenv(distro.AmentPrefixPath))

	if _, ok := distro.SupportedDistroMapper[rosVersion]; !ok {
		return fmt.Errorf("unsupported distro: %s", rosVersion)
	}

	destPath := getString(cmd, "dest-path")
	if destPath == "" {
		return fmt.Errorf("dest-path is required")
	}

	_, err := os.Stat(destPath)
	if errors.Is(err, os.ErrNotExist) {
		//#nosec G301 -- The generated directory doesn't contain secrets.
		err = os.MkdirAll(destPath, 0755)
	}
	if err != nil {
		return fmt.Errorf("dest-path error: %v", err)
	}

	return nil
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Go bindings for ROS2 interface definitions under <root-path>",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := getConfig(cmd)
		if err != nil {
			return err
		}
		gen := gogen.New(config)
		if err := gen.GenerateGolangMessageTypes(); err != nil {
			return fmt.Errorf("failed to generate interface bindings: %w", err)
		}
		if err := gen.GenerateROS2AllMessagesImporter(); err != nil {
			return fmt.Errorf("failed to generate all importer: %w", err)
		}
		if err := gen.GenerateCGOFlags(); err != nil {
			return fmt.Errorf("failed to generate CGO flags: %w", err)
		}
		return nil
	},
	Args: validate,
}

var generateRclgoCmd = &cobra.Command{
	Use:   "generate-rclgo",
	Short: "Generate Go code that forms a part of rclgo",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := getConfig(cmd)
		if err != nil {
			return err
		}
		gen := gogen.New(config)
		if err := gen.GeneratePrimitives(); err != nil {
			return fmt.Errorf("failed to generate primitive types: %w", err)
		}
		if err := gen.GenerateRclgoFlags(); err != nil {
			return fmt.Errorf("failed to generate rclgo flags: %w", err)
		}
		if err := gen.GenerateROS2ErrorTypes(); err != nil {
			return fmt.Errorf("failed to generate error types: %w", err)
		}
		return nil
	},
	Args: validate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	configureFlags(generateCmd, "./ros2_msgs")

	rootCmd.AddCommand(generateRclgoCmd)
	configureFlags(generateRclgoCmd, gogen.RclgoRepoRootPath())
}

func configureFlags(cmd *cobra.Command, destPathDefault string) {
	cmd.PersistentFlags().StringArrayP("root-path", "r", []string{os.Getenv("AMENT_PREFIX_PATH")}, "Root lookup path for ROS2 .msg files. If ROS2 environment is sourced, is auto-detected.")
	cmd.PersistentFlags().StringP("dest-path", "d", destPathDefault, "Output directory for the Golang ROS2 messages.")
	cmd.PersistentFlags().String("rclgo-import-path", gogen.DefaultConfig.RclgoImportPath, "Import path of rclgo library")
	cmd.PersistentFlags().String("message-module-prefix", gogen.DefaultConfig.MessageModulePrefix, "Import path prefix for generated message binding modules")
	cmd.PersistentFlags().StringArray("include-package", nil, "Include only packages matching a regex. Can be passed multiple times. If multiple include options are passed, the union of the matches is generated.")
	cmd.PersistentFlags().StringArray("include-package-deps", nil, "Include only packages which are dependencies of listed packages. Can be passed multiple times. If multiple include options are passed, the union of the matches is generated.")
	cmd.PersistentFlags().StringArray("include-go-package-deps", nil, "Include only packages which are dependencies of listed Go packages. Can be passed multiple times. If multiple include options are passed, the union of the matches is generated.")
	cmd.PersistentFlags().String("cgo-flags-path", "cgo-flags.env", `Path to file where CGO flags are written. If empty, no flags are written. If "-", flags are written to stdout.`)
	bindPFlags(cmd)
}

func getPrefix(cmd *cobra.Command) string {
	var parts []string
	for c := cmd; c != c.Root(); c = c.Parent() {
		parts = append(parts, c.Name())
	}
	for i := 0; i < len(parts)/2; i++ {
		parts[i], parts[len(parts)-i-1] = parts[len(parts)-i-1], parts[i]
	}
	prefix := strings.Join(parts, ".")
	if prefix != "" {
		prefix += "."
	}
	return prefix
}

func getString(cmd *cobra.Command, key string) string {
	return viper.GetString(getPrefix(cmd) + key)
}

func getBool(cmd *cobra.Command, key string) bool {
	return viper.GetBool(getPrefix(cmd) + key)
}

func bindPFlags(cmd *cobra.Command) {
	prefix := getPrefix(cmd)
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		_ = viper.BindPFlag(prefix+f.Name, f)
	})
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		_ = viper.BindPFlag(prefix+f.Name, f)
	})
}

func getConfig(cmd *cobra.Command) (*gogen.Config, error) {
	destPath := getString(cmd, "dest-path")
	modulePrefix := getString(cmd, "message-module-prefix")

	if modulePrefix == gogen.DefaultConfig.MessageModulePrefix {
		pkgs, err := packages.Load(&packages.Config{})
		if err == nil && len(pkgs) > 0 {
			modulePrefix = path.Join(pkgs[0].PkgPath, destPath)
		}
	}
	rules, err := getPackageRules(cmd)
	if err != nil {
		return nil, err
	}
	return &gogen.Config{
		RclgoImportPath:     getString(cmd, "rclgo-import-path"),
		MessageModulePrefix: modulePrefix,
		RootPaths:           getRootPaths(cmd),
		DestPath:            destPath,
		CGOFlagsPath:        getString(cmd, "cgo-flags-path"),

		RegexIncludes:  rules,
		ROSPkgIncludes: viper.GetStringSlice(getPrefix(cmd) + "include-package-deps"),
		GoPkgIncludes:  viper.GetStringSlice(getPrefix(cmd) + "include-go-package-deps"),
	}, nil
}

func getRootPaths(cmd *cobra.Command) []string {
	pathLists := viper.GetStringSlice(getPrefix(cmd) + "root-path")
	found := make(map[string]bool)
	var paths []string
	for _, pl := range pathLists {
		for _, p := range filepath.SplitList(pl) {
			if !found[p] {
				found[p] = true
				paths = append(paths, p)
			}
		}
	}
	return paths
}

func getPackageRules(cmd *cobra.Command) (_ gogen.RuleSet, err error) {
	includes := viper.GetStringSlice(getPrefix(cmd) + "include-package")
	rules := make(gogen.RuleSet, len(includes))
	for i, pattern := range includes {
		rules[i], err = gogen.NewRule(pattern)
		if err != nil {
			return nil, err
		}
	}
	return rules, nil
}
