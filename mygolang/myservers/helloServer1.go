package servers

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var serverPort = 30303

func TestHelloServer1() {
	fmt.Println(">>>>>HTTP SERVER START<<<<<<<<")
	//0. 初始化一个上下文
	ctx, cancel := context.WithCancel(context.Background())

	//1.注册handler方法到缺省路由树:DefaultServeMux上，针对指定的path注册对应的处理函数
	http.HandleFunc("/clusters ", func(w http.ResponseWriter, r *http.Request) {
		if s, err := ioutil.ReadAll(r.Body); err != nil {
			log.Fatalf("Read err= %v;", err)
		} else {
			fmt.Printf("I GOT the request:%s",s)
		}
		w.Write([]byte("I GOT YOU!"))
		w.WriteHeader(200)
		cancel()    //准备退出
	})

	//2. 初始化http服务实例：指定监听的地址和端口号
	server := &http.Server{Addr: ":30303"}

	//3.异步启动服务器的listen and server
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	//4.(可选)阻塞等待服务器退出
	<-ctx.Done()
	if err := server.Shutdown(ctx); err != nil && err != context.Canceled {
		log.Println(err)
	}
	fmt.Println("I GOT the message, BYE BYE!")

}
