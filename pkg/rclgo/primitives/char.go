package primitives

// #include "rosidl_runtime_c/string.h"
// #include "rosidl_runtime_c/primitives_sequence.h"
import "C"
import (
	"unsafe"
)

type CChar = C.schar
type CcharSequence = C.rosidl_runtime_c__char__Sequence

func CharSequenceToGo(goSlice *[]byte, cSlice CcharSequence) {
	if cSlice.size == 0 {
		return
	}
	*goSlice = make([]byte, cSlice.size)
	src := unsafe.Slice((*byte)(unsafe.Pointer(cSlice.data)), cSlice.size)
	copy(*goSlice, src)
}
func CharSequenceToC(cSlice *CcharSequence, goSlice []byte) {
	if len(goSlice) == 0 {
		cSlice.data = nil
		cSlice.capacity = 0
		cSlice.size = 0
		return
	}
	cSlice.data = (*C.schar)(C.malloc(C.sizeof_schar * C.size_t(len(goSlice))))
	cSlice.capacity = C.size_t(len(goSlice))
	cSlice.size = cSlice.capacity
	dst := unsafe.Slice((*byte)(unsafe.Pointer(cSlice.data)), cSlice.size)
	copy(dst, goSlice)
}
func CharArrayToGo(goSlice []byte, cSlice []CChar) {
	for i := 0; i < len(cSlice); i++ {
		goSlice[i] = byte(cSlice[i])
	}
}
func CharArrayToC(cSlice []CChar, goSlice []byte) {
	for i := 0; i < len(goSlice); i++ {
		cSlice[i] = C.schar(goSlice[i])
	}
}
