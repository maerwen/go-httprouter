package httprouter

// 常量，变量，函数
// 常量
const (
	static   nodeType = iota //default
	root                     //根节点
	param                    //参数匹配
	catchAll                 //全匹配
)

// 一个空标识符变量调用New（）

// New() 返回一个*Router

// min 返回两个数之间较小值
func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// countParams，计算params的个数
func countParams(path string) uint8 {
	var n uint
	for i := 0; i < len(path); i++ {
		// 与两种路径匹配关键字符比较
		if path[i] != ':' && path[i] != '*' {
			continue
		}
		n++
	}
	// url上限255byte
	if n >= 255 {
		return 255
	}
	return uint8(n)
}

// shiftNRuneBytes将给定数组中元素按照给定的数字向左移位
// shift bytes in array by n bytes left
func shiftNRuneBytes(rb [4]byte, n int) [4]byte {
	switch n {
	case 0:
		return rb
	case 1:
		return [4]byte{rb[1], rb[2], rb[3], 0}
	case 2:
		return [4]byte{rb[2], rb[3]}
	case 3:
		return [4]byte{rb[3]}
	default:
		return [4]byte{}
	}
}
