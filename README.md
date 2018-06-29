# go-httprouter
original repository: https:github.com/julienschmidt/httprouter

httprouter包是一个建立在高性能的HTTP请求路由上的词典树

一个很寻常的例子:
    package main
    import (
        "fmt"
        "github.com/julienschmidt/httprouter"
        "net/http"
        "log"
    )
    func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
        fmt.Fprint(w, "Welcome!\n")
    }
    func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
        fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
    }
    func main() {
        router := httprouter.New()
        router.GET("/", Index)
        router.GET("/hello/:name", Hello)
        log.Fatal(http.ListenAndServe(":8080", router))
    }

这个路由器通过请求方法和路径来匹配接入的请求.
如果为该路径和方法已经注册了一个处理器,路由会把请求导向这个处理器.
对于GET,POST,PUT,PATCH和DELETE请求方式,已经注册了便捷处理器.对于其他方式的请求,可以通过router.Handle来注册路由
对于已经注册的路径,在路由器去匹配传入的请求时,路径内可以包含两类参数:
    符号    类型
    :name   named parameter(命名参数)
    *name   catch-all parameter(全匹配参数)
named parameter是动态路径部分,他们匹配一切路径直到下一个'/'出现或者路径结束:
    Path: /blog/:category/:post

    Requests                            result
    /blog/go/request-routers            match: category="go", post="request-routers"
    /blog/go/request-routers/           no match, but the router would redirect
    /blog/go/                           no match
    /blog/go/request-routers/comments   no match
Catch-all parameters匹配路径结束之前的一切字段,其中包含了目录索引(开始匹配位置处的'/').因为它需要匹配到路径结束,所以它必须处于路径最后一部分.
    Path: /files/*filepath

    Requests                            result
    /files/                             match: filepath="/"
    /files/LICENSE                      match: filepath="/LICENSE"
    /files/templates/article.html       match: filepath="/templates/article.html"
    /files                              no match, but the router would redirect
参数的值被存储为一个Param结构的slice,每个参数对应一个key和一个value.这个slice最终作为第三个参数被传入Handle方法里面.
有以下两种方式去获取这个参数的value:
    1.通过这个参数的名称:
        user := ps.ByName("user") 定义为 :user或*user
    2.通过这个参数的索引,同时这个方式还可以获取key(参数名称):
        thirdKey   := ps[2].Key   第三个参数的名称
        thirdValue := ps[2].Value 第三个参数的值
