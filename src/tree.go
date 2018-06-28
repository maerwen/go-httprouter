package httprouter

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// 常量
const (
	static   nodeType = iota //default
	root                     //根节点
	param                    //参数匹配
	catchAll                 //全匹配
)

//nodeType 类型
type nodeType uint8

// node类型
// 解惑???????????????
type node struct {
	path      string //该节点所在路径字符串
	wildChild bool   //代表什么？
	nType     nodeType
	maxParams uint8  //0~255
	indices   string //在字典树中索引字符串
	children  []*node
	handle    Handle
	priority  uint32 //优先权,包括本身在内地层级数目
}

// incrementChildPrio方法，增加给出索引对应的子节点的优先权，
// 同时进行排序，并返回排序后的该子节点的新索引
func (n *node) incrementChildPrio(pos int) int {
	// 增加给出索引对应的子节点的优先权
	n.children[pos].priority++
	priority := n.children[pos].priority
	// 排序,根据与前面子节点优先权大小比较，确定该节点是否需要前移
	newPos := pos
	for newPos > 0 && priority > n.children[newPos-1].priority {
		n.children[newPos-1], n.children[newPos] = n.children[newPos], n.children[newPos-1]
		newPos--
	}
	// 修改node的indices属性
	if pos != newPos {
		n.indices = n.indices[:newPos] + //前缀不变，可能为空
			n.indices[pos:pos+1] + //所改变的索引位置
			n.indices[newPos:pos] + //两次位置中间所夹带的部分
			n.indices[pos+1:] //最后的部分
	}
	return newPos
}

// addRoute方法，把给定的handle与path关联起来
// 并发情况下不安全！
func (n *node) addRoute(path string, handle Handle) {
	// 优先权增加（路径越长，节点下路由越多越靠前、越优先）
	fullPath := path
	// 目录层级数目
	numParams := countParams(path)
	n.priority++
	if len(n.path) > 0 || len(n.children) > 0 {
		// 一颗非空的词典树
	walk:
		for {
			// 更新node的maxParams属性
			if numParams > n.maxParams {
				n.maxParams = numParams
			}
			//找到path和node.path公共前缀的长度
			i := 0
			max := min(len(path), len(n.path))
			for i < max && path[i] == n.path[i] {
				i++ //此处是否应确保字符连续性？？？？？？
			}
			//分割
			if i < len(n.path) {
				// 解惑??????????????
				child := node{
					path:      n.path[i:],
					wildChild: n.wildChild, //?
					nType:     static,
					indices:   n.indices,  //?
					children:  n.children, //?
					handle:    n.handle,   //?
					priority:  n.priority - 1,
				}
				// 给child的maxParams属性赋值
				for i := range child.children {
					if child.children[i].maxParams > child.maxParams {
						child.maxParams = child.children[i].maxParams
					}
				}
				// 解惑??????????
				n.children = []*node{&child}          //其他节点呢？直接丢弃吗？
				n.indices = string([]byte{n.path[i]}) //字符索引为单个字母
				n.path = path[:i]                     //那之前部分的从此节点往下的词典树呢？
				n.handle = nil                        //节点处不需设置handle
				n.wildChild = false
			}
			// 使新节点成为这个节点的子节点
			if i < len(path) {
				//公共前缀可以再缩减
				path := path[i:]
				//
				if n.wildChild {
					n = n.children[0]
					n.priority++
					// 更新子节点的maxParams
					if numParams > n.maxParams {
						n.maxParams = numParams
					}
					numParams--
					// 检查通配符匹配是否准确
					if len(path) >= len(n.path) && n.path == path[:len(n.path)] &&
						// 检查更长的通配符
						len(n.path) >= len(path) || path[len(n.path)] == '/' {
						continue walk
					} else {
						// 通配符冲突
						var pathSeg string
						if n.nType == catchAll {
							// 另一种匹配方式
							pathSeg = path
						} else {
							pathSeg = strings.SplitN(path, "/", 2)[0]
						}
						prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path
						panic("'" + pathSeg +
							"' in new path '" + fullPath +
							"' conflicts with existing wildcard '" + n.path +
							"' in existing prefix '" + prefix +
							"'")
					}
				}
				c := path[0]
				// 每个param后面的斜杠
				if n.nType == param && c == '/' && len(n.children) == 1 {
					n = n.children[0]
					n.priority++
					continue walk
				}
				// 检查是否存在一个子节点带有下一条路径
				for i := 0; i < len(n.indices); i++ {
					if c == n.indices[i] {
						i = n.incrementChildPrio(i)
						n = n.children[i]
						continue walk
					}
				}
				// 否则插入节点
				if c != ':' && c != '*' {
					n.indices += string([]byte{c})
					child := &node{
						maxParams: numParams,
					}
					n.children = append(n.children, child)
					n.incrementChildPrio(len(n.children) - 1) //?
					n = child
				}

				n.insertChild(numParams, path, fullPath, handle)
				return
			} else if i == len(path) {
				// 把节点加入路径
				if n.handle != nil {
					panic("a handle is already registered for path '" + fullPath + "'")
				}
				n.handle = handle
			}
			return

		}
	} else {
		// 空词典树
		n.insertChild(numParams, path, fullPath, handle)
		n.nType = root
	}
}

// insertChild方法，插入子节点
func (n *node) insertChild(numParams uint8, path, fullPath string, handle Handle) {
	var offset int //路径中已经处理的字节数
	// 发现第一个通配符前面的前缀
	for i, max := 0, len(path); numParams > 0; i++ {
		// 首先判断是不是通配符
		c := path[i]
		if c != ':' && c != '*' {
			continue
		}
		// 找到结束时的通配符（path结束或者/）
		end := i + 1
		for end < max && path[end] != '/' {
			switch path[end] {
			// 通配符名字必须不包含':' 与 '*'
			case ':', '*':
				panic("only one wildcard per path segment is allowed, has: '" +
					path[i:] + "' in path '" + fullPath + "'")
			default:
				end++
			}
		}
		// 检查当我们在此处插入这个通配符时这个node是否会产生无法到达的子节点
		if len(n.children) > 0 {
			panic("wildcard route '" + path[i:end] +
				"' conflicts with existing children in path '" + fullPath + "'")
		}
		// 检查通配符是否有一个名字,而不仅仅是':' 与 '*'两个单独的字符
		if end-i < 2 {
			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
		}
		if c == ':' {
			// param 匹配
			// 在通配符刚开始的时候分割路径
			if i > 0 {
				n.path = path[offset:i]
				offset = i
			}
			child := &node{
				nType:     param,
				maxParams: numParams,
			}
			// 解惑？？？？？？？？？
			n.children = []*node{child}
			n.wildChild = true
			n = child
			n.priority++
			numParams--
			// 如果路径不以通配符结尾,那么此处将会存在一个以'/'开头的非通配符的子路径
			if end < max {
				n.path = path[offset:end]
				offset = end
				child := &node{
					maxParams: numParams,
					priority:  1,
				}
				n.children = []*node{child}
				n = child
			}
		} else { //全匹配
			// 不是在路径结束的位置
			if end != max || numParams > 1 {
				panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
			}
			// 根路径
			if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
				panic("catch-all conflicts with existing handle for the path segment root in path '" + fullPath + "'")
			}
			//目前为'/'修正宽度为1
			i--
			if path[i] != '/' {
				panic("no / before catch-all in path '" + fullPath + "'")
			}
			n.path = path[offset:i]
			//第一个节点:空路径全匹配
			child := &node{
				wildChild: true,
				nType:     catchAll,
				maxParams: 1,
			}
			n.children = []*node{child}
			n.indices = string(path[i])
			n = child
			n.priority++
			// 第二个节点:节点存储变量
			child = &node{
				path:      path[i:],
				nType:     catchAll,
				maxParams: 1,
				handle:    handle,
				priority:  1,
			}
			n.children = []*node{child}
			return
		}
	}
	//将剩余路径部分和句柄handle插入到链条中
	n.path = path[offset:]
	n.handle = handle
}

// getValue方法
// 返回注册了指定路径的handle
// 通配符的值被存储到了一个map中
// 如果该路径没有对应的handle,但却有一个在其基础上尾部含有'/'的路径,建议重定向
func (n *node) getValue(path string) (handle Handle, p Params, tsr bool) {
walk:
	for {
		if len(path) > len(n.path) {
			if path[:len(n.path)] == n.path {
				path = path[len(n.path):]
				// 如果当前节点没有一个通配符,我们只能继续检索词典树和下一个子节点
				if !n.wildChild {
					c := path[0]
					for i := 0; i < len(n.indices); i++ {
						if c == n.indices[i] {
							n = n.children[i]
							continue walk
						}
					}
					// 没有找到,
					// 如果一个同网址的链条存在,我们可以建议重定向到相同的不包含'/'的网址
					tsr = (path == "/" && n.handle != nil)
					return
				}
				// 处理通配符子节点
				n = n.children[0]
				switch n.nType {
				case param:
					// 发现param结束('/'或结束)
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}
					// 保存param的value
					if p == nil {
						// 延迟分配
						p = make(Params, 0, n.maxParams)
					}
					i := len(p)
					// 在预先分配的容量内展开切片
					p = p[:i+1]
					p[i].Key = n.path[1:]
					p[i].Value = path[:end]
					// 进一步深入
					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue walk
						}
						//
						tsr = (len(path) == end+1)
						return
					}
					if handle = n.handle; handle != nil {
						return
					} else if len(n.children) == 1 {
						// 没有处理器,检查是否存在一个处理该路径上带'/'的处理器
						n = n.children[0]
						tsr = (n.path == "/" && n.handle != nil)
					}
					return
				case catchAll:
					// 保存param的value
					if p == nil {
						// 延迟分配
						p = make(Params, 0, n.maxParams)
					}
					i := len(p)
					// 在预先分配的容量内展开切片
					p = p[:i+1]
					p[i].Key = n.path[2:]
					p[i].Value = path
					handle = n.handle
					return
				default:
					panic("invalid node type")
				}
			}
		} else if path == n.path {
			// 或许我们又来到了已经包含handle处理器的节点
			// 检查我们所找的节点是否已经有处理器
			if handle = n.handle; handle != nil {
				return
			}
			if path == "/" && n.wildChild && n.nType != root {
				tsr = true
				return
			}
			// 没有处理器,检查是否存在一个处理该路径上带'/'的处理器
			for i := 0; i < len(n.indices); i++ {
				if n.indices[i] == '/' {
					n = n.children[i]
					tsr = (len(n.path) == 1 && n.handle != nil) ||
						(n.nType == catchAll && n.children[0].handle != nil)
					return
				}
			}
			return
		}
		// 没有找到,
		// 如果一个同网址的链条存在,我们可以建议重定向到相同的不包含'/'的网址
		tsr = (path == "/") ||
			(len(n.path) == len(path)+1 && n.path[len(path)] == '/' &&
				path == n.path[:len(n.path)-1] && n.handle != nil)
		return
	}
}

// findCaseInsensitivePath方法
// 利用给定的路径进行一次大小写敏感的检索并尝试去找到一个处理器
// 它可以随意的修正尾部的'/'
// 返回一个大小写校正的路径和一个表明检索是否成功的布尔值
func (n *node) findCaseInsensitivePath(path string, fixTrailingSlash bool) (ciPath []byte, found bool) {
	return n.findCaseInsensitivePathRec(
		path,
		strings.ToLower(path),
		// 为新路径预留足够内存
		make([]byte, 0, len(path)+1),
		// 空字符缓冲区
		[4]byte{},
		fixTrailingSlash,
	)
}

// findCaseInsensitivePathRec
// 递归大小写敏感的检索功能
func (n *node) findCaseInsensitivePathRec(path, loPath string, ciPath []byte, rb [4]byte, fixTrailingSlash bool) ([]byte, bool) {
	loNPath := strings.ToLower(n.path)
walk: //遍历词典树的外层循环
	for len(loPath) >= len(loNPath) && (len(loNPath) == 0 || loPath[1:len(loNPath)] == loNPath[1:]) {
		// 给结果添加公共路径
		ciPath = append(ciPath, n.path...)
		if path = path[len(n.path):]; len(path) > 0 {
			loOld := loPath
			loPath = loPath[len(loNPath):]
			// 如果当前节点没有一个通配符子节点,我们可以检索下一个子节点并且继续遍历词典树
			if !n.wildChild {
				//跳过已经处理过的字节数组
				rb = shiftNRuneBytes(rb, len(loNPath))
				if rb[0] != 0 {
					// 旧的字节还未结束
					for i := 0; i < len(n.indices); i++ {
						if n.indices[i] == rb[0] {
							// 继续对子节点进行该操作
							n = n.children[i]
							loNPath = strings.ToLower(n.path)
							continue walk
						}
					}
				} else {
					// 处理新的字节数据
					var rv rune
					// 查找开始数据
					// 数据长度为4
					var off int
					for max := min(len(loNPath), 3); off < max; off++ {
						if i := len(loNPath) - off; utf8.RuneStart(loOld[i]) {
							//从缓存的小写路径中读取数据
							rv, _ = utf8.DecodeRuneInString(loOld[i:])
							break
						}
					}
					//计算当前数据的小写字节
					utf8.EncodeRune(rb[:], rv)
					// 跳过已处理的数据
					rb = shiftNRuneBytes(rb, off)
					for i := 0; i < len(n.indices); i++ {
						// 小写匹配
						if n.indices[i] == rb[0] {
							// 必须使用递归去处理,直到大写和小写都作为索引而存在
							if out, found := n.children[i].findCaseInsensitivePathRec(
								path, loPath, ciPath, rb, fixTrailingSlash,
							); found {
								return out, true
							}
							break
						}
					}
					//如果不同,对于大写字符也是如此
					if up := unicode.ToUpper(rv); up != rv {
						utf8.EncodeRune(rb[:], up)
						rb = shiftNRuneBytes(rb, off)
						for i := 0; i < len(n.indices); i++ {
							// 大写匹配
							if n.indices[i] == rb[0] {
								// 继续对子节点进行处理
								n = n.children[i]
								loNPath = strings.ToLower(n.path)
								continue walk
							}
						}
					}
				}
				// 没有找到,
				// 如果一个同网址的链条存在,我们可以建议重定向到相同的不包含'/'的网址
				return ciPath, (fixTrailingSlash && path == "/" && n.handle != nil)
			}
			n = n.children[0]
			switch n.nType {
			case param:
				// 发现param结束('/'或者结束)
				k := 0
				for k < len(path) && path[k] != '/' {
					k++
				}
				// 把param的值添加到大小写不敏感的path上去
				ciPath = append(ciPath, path[:k]...)
				// 进一步深入
				if k < len(path) {
					if len(n.children) > 0 {
						// 在子节点上继续
						n = n.children[0]
						loNPath = strings.ToLower(n.path)
						loPath = loPath[k:]
						path = path[k:]
						continue
					}
					//
					if fixTrailingSlash && len(path) == k+1 {
						return ciPath, false
					}
					return ciPath, false
				}
				if n.handle != nil {
					return ciPath, true
				} else if fixTrailingSlash && len(n.children) == 1 {
					// 没有处理器,检查是否存在一个处理该路径上带'/'的处理器
					n = n.children[0]
					if n.path == "/" && n.handle != nil {
						return append(ciPath, '/'), true
					}
				}
				return ciPath, false
			case catchAll:
				return append(ciPath, path...), true
			default:
				panic("invalid node byte")
			}
		} else {
			// 或许我们又来到了已经包含handle处理器的节点
			// 检查我们所找的节点是否已经有处理器
			if n.handle != nil {
				return ciPath, true
			}
			// 如果没发现handle,就尝试去通过添加一个'/'去修正路径
			if fixTrailingSlash {
				for i := 0; i < len(n.indices); i++ {
					if n.indices[i] == '/' {
						n = n.children[i]
						if (len(n.path) == 1 && n.handle != nil) ||
							(n.nType == catchAll && n.children[0].handle != nil) {
							return append(ciPath, '/'), true
						}
						return ciPath, false
					}
				}
			}
			return ciPath, false
		}
	}
	// 如果没找到,就尝试通过添加/移除一个'/'去修正路径
	if fixTrailingSlash {
		if path == "/" {
			return ciPath, true
		}
		if len(loPath)+1 == len(loNPath) && loNPath[len(loPath)] == '/' &&
			loPath[1:] == loNPath[1:len(loPath)] && n.handle != nil {
			return append(ciPath, n.path...), true
		}
	}
	return ciPath, false
}

// 公用函数
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
