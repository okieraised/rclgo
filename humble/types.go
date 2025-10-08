package humble

import (
	"encoding/hex"
	"time"
	"unsafe"
)

type Message interface {
	CloneMsg() Message
	SetDefaults()
	GetTypeSupport() MessageTypeSupport
}

type MessageTypeSupport interface {
	New() Message
	PrepareMemory() unsafe.Pointer
	ReleaseMemory(p unsafe.Pointer)
	AsCStruct(dst unsafe.Pointer, src Message)
	AsGoStruct(dst Message, src unsafe.Pointer)
	TypeSupport() unsafe.Pointer // *C.rosidl_message_type_support_t
}

type ServiceTypeSupport interface {
	Request() MessageTypeSupport
	Response() MessageTypeSupport
	TypeSupport() unsafe.Pointer // *C.rosidl_service_type_support_t
}

const GoalIDLen = 16

type GoalID [GoalIDLen]byte

func (id *GoalID) String() string {
	return hex.EncodeToString(id[:])
}

type ActionTypeSupport interface {
	Goal() MessageTypeSupport
	SendGoal() ServiceTypeSupport
	NewSendGoalResponse(accepted bool, stamp time.Duration) Message

	Result() MessageTypeSupport
	GetResult() ServiceTypeSupport
	NewGetResultResponse(status int8, result Message) Message

	CancelGoal() ServiceTypeSupport

	Feedback() MessageTypeSupport
	FeedbackMessage() MessageTypeSupport
	NewFeedbackMessage(goalID *GoalID, feedback Message) Message

	GoalStatusArray() MessageTypeSupport

	TypeSupport() unsafe.Pointer // *C.rosidl_action_type_support_t
}
