package mymethod

import (
	"fmt"
)
type manager struct {

}

//首先, 方法要求的入参是实现了Runnable接口的类型
func (*manager)startRunnable(r Runnable) {
	fmt.Printf("r is:%t\n", r)
	r.Start()
}

//然后, Runnable接口定义如下
type Runnable interface {
	Start() error
}

//在然后. RunnableFunc是一个方法类型，实现了Runnable接口
//       同时，也表示Start()方法是RunnableFunc类型的一个方法集
type RunnableFunc func() error
// Start implements Runnable
func (r RunnableFunc) Start() error { //表示我的receiver是个Runnable类型
	return r()
}

//最后，自定义一个方法并被强转成RunnableFunc类型，
//     于是,
func TestMyMethodForConversion(){
	tmp := RunnableFunc(func() error {
		fmt.Println("I am in Runnable Func")
		
		return nil
	})  //tmp是一个Runnable类型的值
	m := &manager{}
	m.startRunnable(tmp)
}