package designMode

import (
	"fmt"
)

//核心思想：抽象工厂模式用于生成产品族的工厂，所生成的对象是有关联的。
//如果抽象工厂退化成生成的对象无关联则成为工厂函数模式。
//比如本例子中使用RDB和XML存储订单信息，抽象工厂分别能生成相关的主订单信息和订单详情信息。
//如果业务逻辑中需要替换使用的时候只需要改动工厂函数相关的类就能替换使用不同的存储方式了。
//具体的实现步骤为：
//首先，需求是这样的，一个产品比如订单在构成/生产的时候, 是需要各种辅佐，
//     比如订单的组成包括概要信息和详细信息(二者都要), 比如要为订单提供保存方式是RDB还是XML(二选一)
//然后，根据订单的构成定义2种接口，分别为概要信息接口和详细信息接口；
//     根据2种保存方式分别是RDB和XML，分别实现如上2种接口；
//     于是，从4个维度，一共有4个订单类
//最后，定义1个工厂接口，接口中要求能够生产出来订单的概要信息部分 和 详细信息部分
//     根据订单的保存方式有两种工厂的实现类，即RDB方式和XML方式的工厂类
//wxy：为什么称为抽象工厂呢？他有什么特点？
//答:     工厂不再只能生产一个产品，而是能生产多个产品，往往是多个不同等级的产品，这些产品也常常是具有关联性，
//    实现的方式是: 首先，工厂接口要求可以生产不同等级产品：
//                其次，不同等级的产品对应不同等级的产品接口；
//                最后，从"工厂类型"和"产品等级"二个维度进行产品的实例化

//1.定义基接口1：称为main即主订单信息 和其两种实现类分别称为RDB和XML(代表存储订单信息的方式)
type OrderMainDAO interface {
	SaveOrderMain()
}
//1.1
type RDBMainDAO struct{}
func (*RDBMainDAO) SaveOrderMain() {
	fmt.Print("rdb main save\n")
}
//1.2
type XMLMainDAO struct{}
func (*XMLMainDAO) SaveOrderMain() {
	fmt.Print("xml main save\n")
}

//2.定义基接口2:称为detail即详细订单信息, 和其两种实现类分别称为RDB和XML(代表存储订单信息的方式)
type OrderDetailDAO interface {
	SaveOrderDetail()
}
//2.1
type RDBDetailDAO struct{}
func (*RDBDetailDAO) SaveOrderDetail() {
	fmt.Print("rdb detail save\n")
}
//2.2
type XMLDetailDAO struct{}
func (*XMLDetailDAO) SaveOrderDetail() {
	fmt.Print("xml detail save")
}

//3, 定义工厂基接口1：即生产订单的工厂接口，包含两个实现类分别称为RDB和XML(代表存储订单信息的方式)
type DAOFactory interface {
	CreateOrderMainDAO() OrderMainDAO
	CreateOrderDetailDAO() OrderDetailDAO
}
//3.1, 称为RDB方式订单信息生产工厂即可以生产如上两种接口类型的RDB实现类实例
type RDBDAOFactory struct{}
func (*RDBDAOFactory) CreateOrderMainDAO() OrderMainDAO {
	return &RDBMainDAO{}
}
func (*RDBDAOFactory) CreateOrderDetailDAO() OrderDetailDAO {
	return &RDBDetailDAO{}
}
//3.2, 称为DAO方式订单信息生产工厂即可以生产如上两种接口类型的DAO实现类实例
type XMLDAOFactory struct{}
func (*XMLDAOFactory) CreateOrderMainDAO() OrderMainDAO {
	return &XMLMainDAO{}
}
func (*XMLDAOFactory) CreateOrderDetailDAO() OrderDetailDAO {
	return &XMLDetailDAO{}
}

//4. 如何从这两个维度(main or detail,  RDB or XML)获取订单信息
//对于一个工厂来说：调用工厂的创建方法生产出来两种基类实例，并调用其业务函数
func getMainAndDetail(factory DAOFactory) {
	factory.CreateOrderMainDAO().SaveOrderMain()
	factory.CreateOrderDetailDAO().SaveOrderDetail()
}


//对于全局来说: 分别调用如下两个工厂生成逻辑
//1.一个工厂句柄，指向RDB类型工厂实例
func ExampleRdbFactory() {
	var factory DAOFactory
	factory = &RDBDAOFactory{}
	getMainAndDetail(factory)    //结果:rdb main save
}                                //     rdb detail save

//2.一个工厂句柄，指向XML类型工厂实例
func ExampleXmlFactory() {
	var factory DAOFactory
	factory = &XMLDAOFactory{}   //结果:xml main save
	getMainAndDetail(factory)    //    xml detail save
}

func TestFactoryAbstract(){
	ExampleRdbFactory()
	ExampleXmlFactory()
}