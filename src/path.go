package httprouter

/*
	CleanPath
		bufApp
*/
// bufApp
// 如果需要的话,在内部帮你延迟创建一个缓冲区
func bufApp(buf *[]byte, s string, w int, c byte) {
	if *buf == nil {
		if s[w] == c {
			return
		}
		*buf = make([]byte, len(s))
		copy(*buf, s[:w])
	}
	(*buf)[w] = c
}

// CleanPath
// path.Clean的URL版本,返回一个权威的剔除了"."和".."元素的URL路径
// 以下规则被迭代/连环运用,直到所有的过程都被执行:
// 		1.把多个'/'替换成一个'/'
// 		2.在当前目录下,消除每个.路径名称元素
// 		3.在父目录下,连同..之前元素与每个..路径名称元素一并消除
// 		4.消除开始了一个根路径的..元素(即把一个路径起始处的/..替换为/)
// 如果执行完以上流程,得到的结果为空字符串,那么返回"/"
func CleanPath(p string) string {
	//把空字符串转换为"/"
	if p == "" {
		return "/"
	}
	n := len(p)
	var buf []byte
	//不变量:
	// 		从path中读取;r是下一个待处理字节的索引
	// 		写到buf中去;w是下一个待写入字节的索引
	// path必须以'/'开头
	r := 1
	w := 1
	if p[0] != '/' {
		r = 0
		buf = make([]byte, n+1)
		buf[0] = '/'
	}
	trailing := n > 1 && p[n-1] == '/'
	// 没有像原生的path包一样设置一个懒加载缓冲区会显得更笨拙,但这个循环的内部关联更完整
	// 因此相比于path包,这个循环没有繁重复杂的函数调用
	for r < n {
		switch {
		case p[r] == '/':
			//对于空路径元素,在其后添加'/'
			r++
		case p[r] == '.' && r+1 == n:
			trailing = true
			r++
		case p[r] == '.' && p[r+1] == '/':
			//.字符
			r++
		case p[r] == '.' && p[r+1] == '.' && (r+2 == n || p[r+2] == '/'):
			// ..元素,移除
			r += 2
			if w > 1 {
				// 回溯
				w--
				if buf == nil {
					for w > 1 && p[w] != '/' {
						w--
					}
				} else {
					for w > 1 && buf[w] != '/' {
						w--
					}
				}
			}
		default:
			// 真正的路径元素
			// 如果必要的话添加'/'
			if w > 1 {
				bufApp(&buf, p, w, '/')
				w++
			}
			// 复制元素
			for r < n && p[r] != '/' {
				bufApp(&buf, p, w, p[r])
				w++
				r++
			}
		}
	}
	// 再添加'/'
	if trailing && w > 1 {
		bufApp(&buf, p, w, '/')
		w++
	}
	if buf == nil {
		return p[:w]
	}
	return string(buf[:w])
}
