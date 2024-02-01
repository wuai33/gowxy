package assert

import (
	"fmt"
	"io"
)

/*
参考链接：https://docs.studygolang.com/ref/spec#Type_assertions
是一个使用在接口值的的操作，标准语法为：x.(T)
断言成功, the value of the expression is the value stored in x and its type is T, 即, X的"动态值"  + 接口的可见范围
断言失败, 则发生run-time panic

方式一：判定接口"类型"变量的动态"值"
x: 表示一个接口类型(interface type)变量;
T: 表示一个具体类型(type, not an interface type)
作用：检查X的dynamic type是否是一个T的"type"
结果：如果检查成功, 则结果值为:"动态值" + "动态类型"
例如：

方式二:
x: ;
T: 接口类型(interface type)
作用：检查X的dynamic type是否实现了接口T

例如：

进阶: 使用ok pattern来代表断言是否成功, 如果失败则不会panic, 例如
v, ok := x.(T)

一个变量有两部分：类型 和 值
静态的"类型": 声明的类型，为编译阶段确定
动态的"值"( dynamic type): 在程序执行阶段, 被动态赋予的类型

*/


type myRW struct {
	content string
}
//io.Read接口要求:reads up to len(p) bytes into p. It returns the number of bytes
func(x myRW)Read(p []byte)(n int, err error){
	p = []byte(x.content)  //入参p将指向content对应的
	fmt.Printf("Read self content:%s to p:%s.\n", x.content, p)
	return
}
//io.Write接口要求:writes len(p) bytes from p to the underlying data stream.
func(x myRW)Write(p []byte)(n int, err error){
	x.content = string(p)
	fmt.Printf("Write self content:%s from p:%s.\n", x.content, p)
	return
}
func TestAssert1(){
	var p_out []byte

	//w的静态类型为io.Writer, 被赋予了myRW类型的实际值,
	//进行方式一判定: 判定接口w的实际"值"为myRW类型实例,
	var w io.Reader = myRW{"orgin"}
	f := w.(myRW)
	f.Read(p_out)   //将p
	f.Write([]byte("wxy"))
	fmt.Printf("Now sef content for f is:%s\n", f.content)
	//fmt.Printf("Now sef content for w is:\n", w.content) 经过断言也不能直接访问w
	w.Read(p_out)



	rw := w.(io.ReadWriter)
	rw.Read([]byte("rw只有Read/Write方法, 不能直接访问content"))

	ww := rw.(io.Writer)
	ww.Write([]byte(""))
}

func TestAssert(){
	fmt.Println("Start Assert test")
	TestAssert1()

}