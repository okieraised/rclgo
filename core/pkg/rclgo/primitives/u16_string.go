package primitives

// #include "rosidl_runtime_c/u16string.h"
import "C"
import (
	"unicode/utf16"
	"unsafe"
)

func U16StringAsCStruct(dst unsafe.Pointer, m string) {
	// rosidl_runtime_c__U16String__assignn() does something like this,
	// but to call it we still need to make a C string and free it.
	mem := (*C.rosidl_runtime_c__U16String)(dst)
	runescape := utf16.Encode([]rune(m))

	mem.data = (*C.ushort)(C.malloc(C.sizeof_ushort * C.size_t(len(runescape)+1)))
	mem.size = C.size_t(len(runescape))
	mem.capacity = C.size_t(len(runescape) + 1)
	memData := unsafe.Slice((*uint16)(mem.data), mem.capacity)
	copy(memData, runescape)
	memData[len(memData)-1] = 0
}

func U16StringAsGoStruct(msg *string, ros2MessageBuffer unsafe.Pointer) {
	mem := (*C.rosidl_runtime_c__U16String)(ros2MessageBuffer)

	*msg = string(utf16.Decode(unsafe.Slice((*uint16)(mem.data), mem.size)))
}

type CU16String = C.rosidl_runtime_c__U16String
type Cu16stringSequence = C.rosidl_runtime_c__U16String__Sequence

func U16stringSequenceToGo(goSlice *[]string, cSlice Cu16stringSequence) {
	if cSlice.size == 0 {
		return
	}
	*goSlice = make([]string, int64(cSlice.size))
	src := unsafe.Slice(cSlice.data, cSlice.size)
	for i := 0; i < int(cSlice.size); i++ {
		U16StringAsGoStruct(&(*goSlice)[i], unsafe.Pointer(&src[i]))
	}
}

func U16stringSequenceToC(cSlice *Cu16stringSequence, goSlice []string) {
	if len(goSlice) == 0 {
		cSlice.data = nil
		cSlice.capacity = 0
		cSlice.size = 0
		return
	}
	cSlice.data = (*C.rosidl_runtime_c__U16String)(C.malloc((C.size_t)(C.sizeof_struct_rosidl_runtime_c__U16String * uintptr(len(goSlice)))))
	cSlice.capacity = C.size_t(len(goSlice))
	cSlice.size = cSlice.capacity
	dst := unsafe.Slice(cSlice.data, cSlice.size)
	for i := range goSlice {
		U16StringAsCStruct(unsafe.Pointer(&dst[i]), goSlice[i])
	}
}

func U16stringArrayToGo(goSlice []string, cSlice []CU16String) {
	for i := 0; i < len(cSlice); i++ {
		U16StringAsGoStruct(&goSlice[i], unsafe.Pointer(&cSlice[i]))
	}
}

func U16stringArrayToC(cSlice []CU16String, goSlice []string) {
	for i := 0; i < len(goSlice); i++ {
		U16StringAsCStruct(unsafe.Pointer(&cSlice[i]), goSlice[i])
	}
}
