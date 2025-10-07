package primitives

// #include "rosidl_runtime_c/string.h"
import "C"
import (
	"strings"
	"unsafe"
)

func StringAsCStruct(dst unsafe.Pointer, m string) {
	mem := (*C.rosidl_runtime_c__String)(dst)
	mem.data = (*C.char)(C.malloc(C.sizeof_char * C.size_t(len(m)+1)))
	mem.size = C.size_t(len(m))
	mem.capacity = C.size_t(len(m) + 1)
	memData := unsafe.Slice((*byte)(unsafe.Pointer(mem.data)), mem.capacity)
	copy(memData, m)
	memData[len(memData)-1] = 0
}

func StringAsGoStruct(m *string, ros2MessageBuffer unsafe.Pointer) {
	mem := (*C.rosidl_runtime_c__String)(ros2MessageBuffer)
	sb := strings.Builder{}
	sb.Grow(int(mem.size))
	sb.Write(unsafe.Slice((*byte)(unsafe.Pointer(mem.data)), mem.size))
	*m = sb.String()
}

type CString = C.rosidl_runtime_c__String
type CStringSequence = C.rosidl_runtime_c__String__Sequence

func StringSequenceToGo(goSlice *[]string, cSlice CStringSequence) {
	if cSlice.size == 0 {
		return
	}
	*goSlice = make([]string, int64(cSlice.size))
	src := unsafe.Slice(cSlice.data, cSlice.size)
	for i := 0; i < int(cSlice.size); i++ {
		StringAsGoStruct(&(*goSlice)[i], unsafe.Pointer(&src[i]))
	}
}

func StringSequenceToC(cSlice *CStringSequence, goSlice []string) {
	if len(goSlice) == 0 {
		cSlice.data = nil
		cSlice.capacity = 0
		cSlice.size = 0
		return
	}
	cSlice.data = (*C.rosidl_runtime_c__String)(C.malloc((C.size_t)(C.sizeof_struct_rosidl_runtime_c__String * uintptr(len(goSlice)))))
	cSlice.capacity = C.size_t(len(goSlice))
	cSlice.size = cSlice.capacity
	dst := unsafe.Slice(cSlice.data, cSlice.size)
	for i := range goSlice {
		StringAsCStruct(unsafe.Pointer(&dst[i]), goSlice[i])
	}
}

func StringArrayToGo(goSlice []string, cSlice []CString) {
	for i := 0; i < len(cSlice); i++ {
		StringAsGoStruct(&goSlice[i], unsafe.Pointer(&cSlice[i]))
	}
}

func StringArrayToC(cSlice []CString, goSlice []string) {
	for i := 0; i < len(goSlice); i++ {
		StringAsCStruct(unsafe.Pointer(&cSlice[i]), goSlice[i])
	}
}
