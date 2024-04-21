package ptr

// Ptr takes in non-pointer and returns a pointer
func Ptr[T any](v T) *T {
	return &v
}

// ConvPtr takes in a pointer and returns a non-pointer
func ConvPtr[T any](v *T) T { return *v }
