package util

// 计算切片总长度
/*
 传入双层切片, 返回其所有元素长度之和
 e.g. LenLoop({`ele3`,`ele2`,`ele1`}) => 12
*/
func LenLoop[T str](arr []T) (o int) {
	for i, r := 0, len(arr); i < r; i++ {
		o += len(arr[i])
	}
	return
}
