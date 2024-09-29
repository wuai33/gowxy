package designMode

import (
	"fmt"
)

//1.首先定义基接口，名叫API
type API interface {
	Say(name string) string
}


//3.定义工厂:根据入参定制出对应的产品实例
func SimpleFactory(t int32) API {
	if t == 1 {
		return &hiAPI{}
	} else if t == 2 {
		return &helloAPI{}
	}
	return nil
}

//2.1 定义产品1对应的结构体，需要实现基接口
type hiAPI struct{}
func (*hiAPI) Say(name string) string {
	return fmt.Sprintf("Hi, %s", name)
}

//2.2 定义产品2对应的结构体，需要实现基接口
type helloAPI struct{}
func (*helloAPI) Say(name string) string {
	return fmt.Sprintf("Hello, %s", name)
}


func TestSimpleFactory(){

	//TestType1 test get hiapi with factory
	api := SimpleFactory(1)
	res1 := api.Say("Tom")
	fmt.Println("the res1:", res1)

	api = SimpleFactory(2)
	res2 := api.Say("Tom")
	fmt.Println("the res2:", res2)

}

//简单的工厂模式
/*
设计架构为：
首先，定义基接口，即产品的模板
然后，定义各种特色性结构体，即按照模板定制产品设计图
最后，定义工厂函数，根据要求生产出对应的产品

*/