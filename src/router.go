package httprouter

import "net/http"

// 结构体定义
// Param 一个用以存储url的结构体
type Param struct {
	Key   string
	Value string
}

// Params 一个Params的slice，内部有序，且可以用索引来取出元素
type Params []Param

// ByName（）,返回Params里面第一个匹配到的Param的Value
func (ps Params) ByName(name string) string {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value
		}
	}
	return ""
}

// Handle 在http.HandlerFunc的基础上新增了一个参数，用以处理http请求
type Handle func(http.ResponseWriter, *http.Request, Params)

// Router 类http.Handler结构体，把不同请求转接到配置好了相应代理的方法上去
type Router struct {
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
