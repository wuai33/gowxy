package designMode

import (
	"fmt"
)

//核心思想：工厂方法模式使用子类的方式延迟生成对象到子类中实现，子类的生成方式采用匿名嵌入式来实现
//具体的实现步骤为：
//1.定义基接口 和 生产基接口实例的基工厂接口
//2.定义基类，部分实现基接口的函数
//3.定制特色类，通过匿名嵌入的方式继承基类，同时"特色"实现基接口要求的剩下的函数，如此一来特色类完全实现了基接口
//4.定制生产特色类的工厂类，实现了基工厂接口
//5.使用：根据特色类实例，一方面可以调用基接口部分的通用函数，另一方面可以调用自己的特色函数
//wxy：基类并没有实现基接口，而是由特色类补齐从而实现，这就是所谓的延迟生成.....

//1.定义底层基接口，名叫"操作符"(Operator)
type Operator interface {
	SetA(int)
	SetB(int)
	Result() int
}

//2. 定义工厂接口，名叫"操作符工厂"(OperatorFactory),指定了一个"生产"函数
type OperatorFactory interface {
	Create() Operator
}

//3.定义底层基结构体： OperatorBase 并不是Operator 接口实现的类？因为他只实现了其中2个公共方法
type OperatorBase struct {
	a, b int
}
func (o *OperatorBase) SetA(a int) {
	o.a = a
}
func (o *OperatorBase) SetB(b int) {
	o.b = b
}

//4. 定制特色结构体 和 对应的工厂结构体，即"具体的操作符" 和 "具体操作符的工厂", 有如下的实现特点：
//   首先,各家里面都要封装基础类型的同名指针变量,所以可以认为他实现了OperatorBase类型的SetA和SetB方法
//   然后，各家自己实现了Result()方法
//   最后，一综合，他算是实现了最底层的Operator接口

//3.1加法操作符: PlusOperator   和   该操作符的生产类：PlusOperatorFactory
type PlusOperator struct {
	*OperatorBase
}
func (o PlusOperator) Result() int {
	return o.a + o.b
}

type PlusOperatorFactory struct{}
func (PlusOperatorFactory) Create() Operator {
	return &PlusOperator{
		OperatorBase: &OperatorBase{},
	}
}

//3.2减法操作符: MinusOperator   和   该操作符的生产类：MinusOperatorFactory
//MinusOperator Operator 的实际减法实现
type MinusOperator struct {
	*OperatorBase
}
func (o MinusOperator) Result() int {
	return o.a - o.b
}

type MinusOperatorFactory struct{}
func (MinusOperatorFactory) Create() Operator {
	return &MinusOperator{
		OperatorBase: &OperatorBase{},
	}
}

//定义计算函数，根据传进来的"特色工厂"，首先生产出来"特色操作", 然后为其设置值，最后调用"特色"的计算函数
func compute(factory OperatorFactory, a, b int) int {
	op := factory.Create()
	op.SetA(a)
	op.SetB(b)
	return op.Result()
}

func TestFactoryMethod(){
	var factory OperatorFactory  	//0.工厂接口句柄

	factory = PlusOperatorFactory{} //1.句柄指向加法工厂
	fmt.Println("1 + 2 =",  compute(factory, 1, 2)) //结果：1 + 2 = 3

	factory = MinusOperatorFactory{} //2.句柄指向减法工厂
	fmt.Println("4 - 3 =",  compute(factory, 4, 3)) //结果：4 - 3 = 1
}

//简单的工厂模式
/*
设计架构为：
首先，定义基接口，即产品的模板
然后，定义各种特色性结构体，即按照模板定制产品设计图
最后，定义工厂函数，根据要求生产出对应的产品

*/