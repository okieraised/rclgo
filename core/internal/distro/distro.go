package distro

var AmentPrefixPath = "AMENT_PREFIX_PATH"

var (
	ROSHumble = "humble"
	ROSJazzy  = "jazzy"
)

var SupportedDistroMapper = map[string]struct{}{
	ROSJazzy:  {},
	ROSHumble: {},
}
