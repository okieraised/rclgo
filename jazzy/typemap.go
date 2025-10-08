package jazzy

import (
	"fmt"
)

// messageTypeMap maps the ROS2 Message type name to the type implementation in Go,
// so the correct type can be dynamically chosen. Needs to be defined for all supported ROS2 message types.
var messageTypeMap = make(map[string]MessageTypeSupport)

// RegisterMessage sets the type string to implementation dispatcher, so the
// correct type can be dynamically chosen. The Golang types of ROS2 Message use
//
//	func init() {}
//
// to automatically populate this when imported.
func RegisterMessage(alias string, msgType MessageTypeSupport) {
	messageTypeMap[alias] = msgType
}

// GetMessage translates "std_msgs/ColorRGBA" to std_msgs.ColorRGBA -Go type
// returns true if the type mapping is found
func GetMessage(msgType string) (MessageTypeSupport, bool) {
	ros2msg, ok := messageTypeMap[msgType]
	return ros2msg, ok
}

// MustGetMessage panics if there is no mapping
func MustGetMessage(msgType string) MessageTypeSupport {
	ros2msg, ok := messageTypeMap[msgType]
	if !ok {
		panic(fmt.Sprintf("No registered implementation for ROS2 message type '%s'!\n", msgType))
	}
	return ros2msg
}

// serviceTypeMap is the messageTypeMap equivalent for services.
var serviceTypeMap = make(map[string]ServiceTypeSupport)

// RegisterService is the RegisterMessage equivalent for services.
func RegisterService(alias string, srvType ServiceTypeSupport) {
	serviceTypeMap[alias] = srvType
}

// GetService is the GetMessage equivalent for services.
func GetService(srvType string) (ServiceTypeSupport, bool) {
	srv, ok := serviceTypeMap[srvType]
	return srv, ok
}

// MustGetService is the MustGetMessage equivalent for services.
func MustGetService(srvType string) ServiceTypeSupport {
	srv, ok := serviceTypeMap[srvType]
	if !ok {
		panic(fmt.Sprintf("No registered implementation for ROS2 message type '%s'!\n", srvType))
	}
	return srv
}

// actionTypeMap is the messageTypeMap equivalent for actions.
var actionTypeMap = make(map[string]ActionTypeSupport)

// RegisterAction is the RegisterMessage equivalent for actions.
func RegisterAction(alias string, actionType ActionTypeSupport) {
	actionTypeMap[alias] = actionType
}

// GetAction is the GetMessage equivalent for actions.
func GetAction(actionType string) (ActionTypeSupport, bool) {
	ac, ok := actionTypeMap[actionType]
	return ac, ok
}

// MustGetAction is the MustGetMessage equivalent for actions.
func MustGetAction(actionType string) ActionTypeSupport {
	ac, ok := actionTypeMap[actionType]
	if !ok {
		panic(fmt.Sprintf("No registered implementation for ROS2 message type '%s'!\n", actionType))
	}
	return ac
}
