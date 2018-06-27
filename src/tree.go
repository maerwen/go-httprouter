package httprouter

import "strings"

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

}

// findCaseInsensitivePath方法，对大小写不敏感的path区分对待
// findCaseInsensitivePathRec递归
