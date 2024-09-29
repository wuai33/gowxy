package designMode

import "fmt"

/*
参考：
https://github.com/senghoo/golang-design-pattern/tree/master/20_decorator
装饰模式使用对象组合的方式动态改变或增加对象行为。
Go语言借助于匿名组合和非入侵式接口可以很方便实现装饰模式。
使用匿名组合，在装饰器中不必显式定义转调原对象方法。

wxy:
Concrete: 具体的(而非想象或猜测的);有形的;实在的。
		  在这里,因为接口是"虚"的,而结构体是"实"的,所以这里用这个单词来修饰结构体
装饰模式的核心是：基本体即被装饰者 和 装饰完的基本体

核心原理：
decorator,意为装饰，既然是装饰则包含了两个层面的意思：首先要有本体作为被装饰者,  然后要有装饰物
所以这种模式首先将本体抽象出来一个接口，然后实现该接口得到一个具象即本体，
最后花样实现接口并将本体实例作为自己的基础即被装饰者，然后通过重写接口方法的方式，在内部调用本体的接口的方法基础上增加花样即装饰物

步骤：
1. 将事物的基本样抽象出一个接口
2. 低配实现该接口作为基本款
3. 花样实现该接口将基本款的实例作为自己的一部分,相当于继承基本款, 然后在重写接口方法的时候添加上"装饰物"



*/

//定义基接口，要求实现一个称为计算的方法
type Base interface {
	Calc() int32
}

//定义基结构体/基类,实现接口(方法)
type ConcreteBase struct{}
func (*ConcreteBase) Calc() int32 {
	fmt.Println("(Base)I am ConcreteBase")
	return 100
}

//定义装饰器1(乘法)，结构体中除了自己的field外，内嵌了Base结构体，相当于继承因此他也实现了Base接口
type MulDecorator struct {
	Base  //结构体的内嵌接口，表示Component Component
	num int32
}

//装饰器1(乘法), 该结构体也实现了Calc()方法, 即实现了Base接口。其内部会调用基类的接口方法。
func (d *MulDecorator) Calc() int32 {
	fmt.Println("(Decorator)I am MulDecorator")
	return d.num * d.Base.Calc()    //调用的是被装饰的实例的Calc()函数 * 自己的num
}
//装饰器1(乘法)的生成函数: 根据入参1(Component接口类型), 入参2, 得到装饰后的具象实例
func WarpMulDecorator(b Base, num int32) Base {
	return &MulDecorator{
		Base: b,
		num:  num,
	}
}


//定义装饰器2(加法)，结构体中同样包含了基本接口和一个自己的field
type AddDecorator struct {
	Base   //结构体的内嵌，表示Component Component
	num int32
}

//装饰器2(加法), 同样实现了接口的要求以及调用父类的接口方法
func (d *AddDecorator) Calc() int32 {
	fmt.Println("(Decorator)I am AddDecorator")
	return d.num + d.Base.Calc()
}
//装饰器2(加法)的生成函数
func WarpAddDecorator(c Base, num int32) Base {
	return &AddDecorator{
		Base: c,
		num:       num,
	}
}

//定义装饰器3，结构体中同样包含了基接口和一个自己的field， 没有实现接口函数
type InheritBase struct {
	Base
	num int32
}

func TestDecorator() {
	//1.初始化1个基结构体ConcreteBase的实例
	var  decorator Base = &ConcreteBase{}

	//2.生成两个装饰器实例，都可以被基接口所指向
	decorator = WarpAddDecorator(decorator, 11) //装饰基实例，得到的是加法装饰器的实例
	decorator = WarpMulDecorator(decorator, 8)  //装饰(经过装饰的)加法类实例，得到乘法装饰器的实例


	//3. 执行计算: .Calc() == 乘法.Calc() == 乘法.Calc() * num ==  (Base.Calc() + num_add) * num_multi
	res := decorator.Calc()
	fmt.Printf("res %d\n", res)      //结果为: 8 * (100 + 11) = 888

	/*1’.传入接口 */
	var decorator1 Base
	decorator1 = &InheritBase{
		Base: decorator1,
		num:       3,
	}
	//fmt.Printf("res1: %d\n", decorator1.Calc())   //运行异常: 空指针引发非法访问


	/*1’’.直接继承 */
	var decorator2 Base = &ConcreteBase{}
	decorator2 = &InheritBase{
		Base: decorator2,
		num:       3,
	}

	fmt.Printf("res2: %d\n", decorator2.Calc())   //运行结果: 100, 即直接调用InheritBase的父类ConcreteBase的Calc()


	//编译报错: decorator.num, decorator2.num undefined (type Base has no field or method num)
	//		   因为decorator或者decorator2是接口Base类型的(选择器), 尽管
	//		   第一次给他的接收器是Base接口的实现类ConcreteBase, 第二次是ConcreteBase的继承类InheritBase
	//		   但是, 选择器本身是没有num元素，只有Calc()方法，而Calc()方法则由接收器决定
	//wxy: 这部分的原理可以参考？？？？？章节
	//fmt.Printf("decorator2.num: %d，decorator2.num: %d\n", decorator.num, decorator2.num)


}

//运行结构如下：
/*
(Decorator)I am MulDecorator
(Decorator)I am AddDecorator
(Base)I am ConcreteComponent
res1 80
(Base)I am ConcreteComponent
res2 0,c2.num
*/

