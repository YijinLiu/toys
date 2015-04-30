package main

// #cgo LDFLAGS: -Wl,-Bstatic -lhello -lstdc++ -Wl,-Bdynamic
// #include "hello.h"
import "C"

func main() {
	C.hello_world()
}
