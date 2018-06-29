package httprouter

import "net/http"

// 变量定义
// 空标识符调用new方法

// 结构体定义
// Param
// 是一个独立的URL参数,由一个key和一个value组成
type Param struct {
	Key   string
	Value string
}

// Params
// 一个Params的slice，其内部有序，所可以用索引来取出元素
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

	//
}

// Handle方法，处理不同类型请求
// HandlerFunc方法，一个允许Router调用http.HandleFunc的适配器
// ServeFiles方法，从给定的文件系统中读取文件

//Get方法 相应路径的Get类型请求处理
//HEAD方法 相应路径的HEAD类型请求处理
//OPTIONS方法 相应路径的OPTIONS类型请求处理
//POST方法 相应路径的POST类型请求处理
//PUT方法 相应路径的PUT类型请求处理
//PATCH方法 相应路径的PATCH类型请求处理
//DELETE方法 相应路径的DELETE类型请求处理

// recv方法，遇到宕机时进行恢复

// Lookup方法用以检索指定方法制定路径下的路由代理，发现了就返回，否则返回false
// allowed方法
// ServeHTTP方法，使Router实现http.Handle接口

// New
