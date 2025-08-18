package utils

// Contains 判断切片是否包含指定元素
func Contains[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}
