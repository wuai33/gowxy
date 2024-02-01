package designMode

import (
	"fmt"
)

/*
参考：
https://github.com/senghoo/golang-design-pattern/blob/master/09_proxy/proxy.go
代理模式用于延迟处理操作或者在进行实际操作前后进行其它处理。


核心原理：
所谓代理，即首先代表自己的客户做一些事情，然后再让客户出面
所以存在两种实现接口的结构体，分别扮演了两个角色：被代理者(本体) 和 代理者，
后者将前者的"实例"作为自己的一部分，在调用本体方法之前或之后会先调用代理的额外逻辑

步骤：
1. 将行为抽象出一个接口
2. 两种类分别实现该接口，一个作为被代理者，一个作为代理者
   被代理者作为本体实现自己的核心业务
   代理者将本体实例作为自己的一部分，实现的接口方法中除了调用核心业务外，还会做一些其他代理需要做的事情


*/

func Testxxxww(){

	fmt.Println("4 - 3 =") //结果：4 - 3 = 1
}
