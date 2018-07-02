package httprouter

import "net/http"

// 变量定义
// 空标识符调用new方法
// 确认Router调用了http.Handler接口
var _ = New()

// 结构体定义
// Param
// 是一个独立的URL参数,由一个key和一个value组成
type Param struct {
	Key   string
	Value string
}

// Params
// 一个Param的slice，其内部有序，所可以用索引来取出元素
type Params []Param

// ByName（）
// 返回key值与给出的参数名称相匹配的Param的value
// 如果没有Param匹配,那么就返回空字符串
func (ps Params) ByName(name string) string {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value
		}
	}
	return ""
}

// Handle
// 一种可以被注册到一个路由上去处理HTTP请求的方法.
// 很接近于http.HandlerFunc,但是添加包含通配符的值的Param为第三个参数
type Handle func(http.ResponseWriter, *http.Request, Params)

// Router
// 类http.Handler结构体，把请求经过配置的路由转接到不同方法上去
type Router struct {
	trees map[string]*node

	// 如果当前的路由不能匹配到一个路由器但存在一个与该路径后添加'/'的路径匹配的处理器,则自动重定向
	// 例如:
	// 		如果/foo/是请求路径但是仅存在一个匹配/foo的路由,
	// 		如果是GET请求方式,http status被设置成301,
	// 		其他请求方式则http status被设置成307
	// 		最后客户端将会被重定向去/foo
	RedirectTrailingSlash bool

	// 如果对于当前路径,没有处理器与它相匹配,如果设为true的话,路由会尝试着去修正该路径
	// 首先多余的路径元素如../或//会被移除
	// 之后路由会对已经精简过的路径进行一次不区分大小写的检索
	// 如果对应该路径找到了一个处理器,路由将会重定向到校正过的路径去
	// 在该过程中,
	// 如果是GET请求方式,http status被设置成301,
	// 其他请求方式则http status被设置成307.
	// 例如:
	// 		/FOO和/..//Foo会被重定向到/foo.
	// RedirectTrailingSlash与该选项无关
	RedirectFixedPath bool

	// 如果设置为true,在当前的请求路径无法被处理时,
	// 路由会去检查是否存在另一个方法对应着这个路径
	// 如果为false,
	// 请求将会收到一段HTTP status为405的响应消息"Method Not Allowed"
	// 如果没有别的方法被允许,该请求会被转接到NotFound处理器上去
	HandleMethodNotAllowed bool

	// 如果为true,路由将会自动响应处理可选范围内的请求
	// 经常性可选的处理器的优先级要高于自动响应
	HandleOPTIONS bool

	// 当没有发现匹配的路径时执行的http.Handler
	// 如果未设置,那么将会调用http.NotFound.
	NotFound http.Handler

	// 当一个请求不能被按照路径处理并且HandleMethodNotAllowed设置为true,
	// 那么配置好的http.Handler将会被调用.
	// 如果这个属性没有设置,一个包含了http.StatusMethodNotAllowed信息的http.Error将会被调用.
	// 在处理器被调用之前,先会设置一个附带允许的请求方法的"Allow"的请求头
	MethodNotAllowed http.Handler

	// 处理宕机并且从http handler恢复数据的功能
	// 它被用来生成一个错误反馈页面并且返回一个为500的http error(Internal Server Error).
	// 这个处理器可以避免你的服务器因为不能恢复的宕机而崩溃
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})
}

// Handle
// Handle为给定的路径和方法注册了一个新的请求处理器
// 对于GET,POST,PUT,PATCH以及DELETE的请求,都有各自的快捷方法可供调用
// 这个方法可以在高负荷下正常使用,并且允许不频繁地,非标准化的私有的方法调用(例如在代理下的内部通信)
func (r *Router) Handle(method, path string, handle Handle) {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}
	if r.trees == nil {
		r.trees = make(map[string]*node)
	}
	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}
	root.addRoute(path, handle)
}

//GET
//快捷调用router.Handle("GET", path, handle)
func (r *Router) GET(path string, handle Handle) {
	r.Handle("GET", path, handle)
}

//HEAD
//快捷调用router.Handle("HEAD", path, handle)
func (r *Router) HEAD(path string, handle Handle) {
	r.Handle("HEAD", path, handle)
}

//OPTIONS
//快捷调用router.Handle("OPTIONS", path, handle)
func (r *Router) OPTIONS(path string, handle Handle) {
	r.Handle("OPTIONS", path, handle)
}

//POST
//快捷调用router.Handle("POST", path, handle)
func (r *Router) POST(path string, handle Handle) {
	r.Handle("POST", path, handle)
}

//PUT
//快捷调用router.Handle("PUT", path, handle)
func (r *Router) PUT(path string, handle Handle) {
	r.Handle("PUT", path, handle)
}

//PATCH
//快捷调用router.Handle("PATCH", path, handle)
func (r *Router) PATCH(path string, handle Handle) {
	r.Handle("PATCH", path, handle)
}

//DELETE
//快捷调用router.Handle("DELETE", path, handle)
func (r *Router) DELETE(path string, handle Handle) {
	r.Handle("DELETE", path, handle)
}

// HandlerFunc
// 一个允许把http.HandleFunc当做request handle来调用的适配器
func (r *Router) HandlerFunc(method, path string, handler http.HandlerFunc) {
	// r.Handler(method, path, handler) //r.Handler?
}

// ServeFiles
// 从给定的文件系统根目录中读取文件
// 路径必须以"/*filepath"结尾,文件都从本地路径/defined/root/dir/*filepath处获取
// 例如:
// 		如果根路径是"/etc"并且*filepath是"passwd",将会找到本地文件"/etc/passwd".
// 本质上调用了一个http.FileServer,因此调用了http.NotFound而不是Router的NotFound处理器.
// 为了使用操作系统的文件系统实现,使用http.Dir:
// 		router.ServeFiles("/src/*filepath", http.Dir("/var/www"))
func (r *Router) ServeFiles(path string, root http.FileSystem) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}
	fileServer := http.FileServer(root)
	r.GET(path, func(w http.ResponseWriter, req *http.Request, ps Params) {
		req.URL.Path = ps.ByName("filepath")
		fileServer.ServeHTTP(w, req)
	})
}

// recv
// 遇到宕机时进行恢复
func (r *Router) recv(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(w, req, rcv)
	}
}

// Lookup
// 允许手动检索一个方法和路径的结合体
// 这是一个围绕路由去建立框架的有用案例.
// 如果该路径被找到了,返回这个处理器函数和路径的参数值.
// 否则第三个返回值表明是否重定向到相同的头部包含'/'的路径
func (r *Router) Lookup(method, path string) (Handle, Params, bool) {
	if root := r.trees[method]; root != nil {
		return root.getValue(path)
	}
	return nil, nil, false
}

// allowed
func (r *Router) allowed(path, reqMethod string) (allow string) {
	if path == "*" { //服务器范围
		for method := range r.trees {
			if method == "OPTIONS" {
				continue
			}
			//把请求的方法添加到允许的方法列表中去
			if len(allow) == 0 {
				allow = method
			} else {
				allow += "," + method
			}
		}
	} else { //特别的路径
		for method := range r.trees {
			// 跳过请求的方法,我们已经尝试过这一个了
			if method == reqMethod || method == "OPTIONS" {
				continue
			}
			handle, _, _ := r.trees[method].getValue(path)
			if handle != nil {
				//把请求的方法添加到允许的方法列表中去
				if len(allow) == 0 {
					allow = method
				} else {
					allow += "," + method
				}
			}
		}
	}
	if len(allow) > 0 {
		allow += ",OPTIONS"
	}
	return
}

// ServeHTTP
// 使Router实现http.Handle接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.PanicHandler != nil {
		defer r.recv(w, req)
	}
	path := req.URL.Path
	if root := r.trees[req.Method]; root != nil {
		if handle, ps, tsr := root.getValue(path); handle != nil {
			handle(w, req, ps)
			return
		} else if req.Method != "CONNECT" && path != "/" {
			code := 301 //GET请求,永久重定向
			if req.Method != "GET" {
				//相同方法,临时重定向
				//在Go1.3版本,不支持308状态码
				code = 307
			}
			if tsr && r.RedirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					req.URL.Path = path[:len(path)-1]
				} else {
					req.URL.Path = path + "/"
				}
				http.Redirect(w, req, req.URL.String(), code)
				return
			}
			// 尝试去修正请求路径
			if r.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path), r.RedirectTrailingSlash,
				)
				if found {
					req.URL.Path = string(fixedPath)
					http.Redirect(w, req, req.URL.String(), code)
					return
				}
			}
		}
	}
	if req.Method == "OPTIONS" && r.HandleOPTIONS {
		// 处理OPTIONS请求
		if allow := r.allowed(path, req.Method); len(allow) > 0 {
			w.Header().Set("Allow", allow)
			return
		}
	} else {
		// 处理405响应状态码
		if r.HandleMethodNotAllowed {
			if allow := r.allowed(path, req.Method); len(allow) > 0 {
				w.Header().Set("Allow", allow)
				if r.MethodNotAllowed != nil {
					r.MethodNotAllowed.ServeHTTP(w, req)
				} else {
					http.Error(w,
						http.StatusText(http.StatusMethodNotAllowed),
						http.StatusMethodNotAllowed,
					)
				}
				return
			}
		}
	}
	// 处理404响应状态码
	if r.NotFound != nil {
		r.NotFound.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}

// New
// 返回一个全新的已经初始化的Router
// 路径自动校正(包含'/'处理)默认调用.
func New() *Router {
	return &Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}
}
