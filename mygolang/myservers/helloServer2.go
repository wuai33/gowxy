package servers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

/*Handler接口要求实现ServeHTTP方法, 如下
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
*/

//0.自定义自己的handler方法，实现Handler接口
type myHandler struct{}
func (myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s, err := ioutil.ReadAll(r.Body); err != nil {
		log.Fatalf("Read err= %v;", err)
	} else {
		fmt.Printf("I GOT the request:%s",s)
	}
	w.Write([]byte("I GOT YOU, too!"))
	w.WriteHeader(200)
}


func TestHelloServer2(){
	fmt.Println(">>>>>HTTP SERVER2  START <<<<<<<<")

	//1.初始化一个自定义的路由分发树，添加路由以及对应的handler方法
	mux := http.NewServeMux()
	mux.Handle("/notify2", myHandler{})
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// The "/" pattern matches everything, so we need to check
		// that we're at the root here.
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		fmt.Fprintf(w, "Welcome to the home page2!")
	})

	//2. 初始化http服务实例：指定监听的地址和端口号
	server := &http.Server{Addr: ":30303", Handler: mux}

	//3.异步启动服务器的listen and server
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Httpserver2: ListenAndServe() error: %s", err)
	}
}

/*
说明：
1.func NewServeMux() *ServeMux
返回值为一个ServeMux类型的指针，该类型实现了如下的接口, 即实现了Handler接口
func (mux *ServeMux) ServeHTTP(w ResponseWriter, r *Request) {...}

2. server的启动，sdk已经帮我们合并了如下的操作
server := &http.Server{Addr: ":30303", Handler: mux}
server.ListenAndServe()
等价于=>
http.ListenAndServe(":8080", mux)


*/