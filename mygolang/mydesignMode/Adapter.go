package designMode

import (
	"fmt"
)

/*
参考：
https://github.com/senghoo/golang-design-pattern/tree/master/02_adapter
适配器模式用于转换一种接口适配另一种接口。
实际使用中Adaptee一般为接口，并且使用工厂函数生成实例。
在Adapter中匿名组合Adaptee接口，所以Adapter类也拥有SpecificRequest实例方法，又因为Go语言中非入侵式接口特征，其实Adapter也适配Adaptee接口。


核心原理：
创建一个适配器，用在独立模块之间独立开发的场景:
模块1要求一个功能实现接口A，但是对应的功能模块定义的接口却是B,
此时就需要一个适配层:
    对上，实现当前模块1的功能接口A，但并不负责实现所谓的"功能";
    对下，调用真正功能模块2的功能接口B;
	那么如何兼顾对上和对下呢？答: 实现接口A，内嵌接口B


步骤：
1. 上，模块开发中，制定了某个功能的接口1，对于适配器来说称为适配的目标接口
2. 下，真正实现了某个功能的另一个模块制定接口2，对于适配器来说称为被适配的目标接口
3. 适配器类,  实现适配的目标接口，接口内部调用被适配的目标实例的接口函数，

*/

//1. Target 是适配的目标接口
type Target interface {
	Request() string
}

//2. Adaptee 是被适配的目标接口
type Adaptee interface {
	SpecificRequest() string
}
//AdapteeImpl 是被适配的目标结构体，实现对应的接口
type adapteeImpl struct{}
func (*adapteeImpl) SpecificRequest() string {
	return "adaptee method"
}
//AdapteeFactory 是被适配接口的工厂函数，用来生成一个被是被适配的目标类实例
func AdapteeFactory() Adaptee {
	return &adapteeImpl{}
}


//3. Adapter即为适配层,转换Adaptee的接口为Target要求的接口的适配器:对外实现目标接口，内部调用被适配类实例接口方法
type adapter struct {
	Adaptee   //内嵌被适配目标接口，其隐含包括了一个同名的filed
}
func (a *adapter) Request() string {
	return a.SpecificRequest()   //****适配的核心: 在目标接口中调用被适配接口*****
}

//AdapterFactory 是Adapter的工厂函数: 给我一个向下的被适配类实例，我给你一个能够向上兼容的适配器
func AdapterFactory(adaptee Adaptee) Target {
	return &adapter{
		Adaptee: adaptee, //将底层被调用类实例作为自己的一个field
	}
}





func TestAdapter(){

	adaptee := AdapteeFactory()
	target := AdapterFactory(adaptee)
	res := target.Request()
	fmt.Println("res:",res) //结果：4 - 3 = 1
}
