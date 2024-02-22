package environment

import "runtime"

func OperatingSystemDetect() *string {
	// Return the operating system runtime
	var osRuntime string = runtime.GOOS
	return &osRuntime
}
