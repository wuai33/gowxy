package myreflect

import (
	"fmt"
	"io"
	"os"
	"reflect"
)

/*
Go提供了一种机制, 允许变量在编译时并没有具体的类型, 而是在运行时更新变量, 然后检查他们的值以及调用他们的方法，
和他们支持的内在操作，这种机制被称为反射


一. 关键概念和函数
1.
Value: 一个结构体(struct)类型,
		实现了Stringer接口, 即String()方法;
		实现了Type()方法, 会返回其动态类型;
		实现了Kind()方法, 返回原始变量的Kind(种类)。
Type: 一个接口(interface)类型,
		实现了Stringer接口, 即String()方法, 为间接实现, 因为该接口本身要求实现
        本身要求实现Kind()方法, 返回原始变量的种类Kind？？？？？？？;

2.
func ValueOf(i interface{}) Value
func TypeOf(i interface{}) Type
解析:
形参, 为interface{}类型, 调用上述函数时, 会将入参做隐式的类型转换, 于是得到了包含两方面的信息的接口(interface{})类型值
	1)动态类型(具体类型)
	2)动态值
经过ValueOf的这个类型装换:
      会获取调用方传入的实参的动态值，返回的Value类型变量。之后，可调用String方法, 但需要注意的是,
      除非, 此Value类型的变量持有的是字符串类型的值,
      否则, String()方法只是返回具体的类型(?1), 可使用v%将结果做格式化

经过TypeOf的这个类型装换:
      会获取调用方传入的实参的动态类型，返回的Type接口类型变量。之后，可调用String方法, 但需要注意的是,
      返回的动态类型是具体的类型,如w的真正类型是*osFile而不是io.Writer, 可使用T%将结果做格式化

3. Kind()
种类, 例如是int?Ptr?Struct？

*/



func TestReflect() {
	fmt.Println("TestReflect=>")
	var i int = 3
	var w io.Writer = os.Stdout

	/*part 1: 反射的基础: TypeOf和ValueOf的作用是什么
	TypeOf: 表示数据类型, 分别是1)int类型;  2)struct类型之os.File类型
	        golang的数据类型分位基础数据类型(比如int)和复合类型(比如struct), 其中所谓复合类型依托基础类型,
	        详见: https://www.cnblogs.com/shuiguizi/p/11372635.html
			所以, 对于基础类型则直接为该类型，而对于复合类型, 则需要明确到是怎么样的基础类型的组合构成的复合类型.
		官方释意: TypeOf returns the reflection Type that represents the dynamic type of i.\

	ValueOf: 表示其值, 确切的说是这个变量对应的底层内存存放的啥,
			所以对于基础类型则直接是其面值，
			但是对于复合类型因为占用的是一整块内存, 那额底层一定有个指针指向他, 所以其值为"一个地址对应的内容"/"内容取地址"
		官方释意: ValueOf returns a new Value initialized to the concrete value stored in the interface i.

	.Type(): 有明确的值了，那么必有对应的类型，因此提供了一个通过值获取类型的方法.
			注: 有类型则只是有类型, 从类型得到值是不可能的, 也不符合逻辑....
		官方释意:
	*/

	t_int := reflect.TypeOf(i)
	t_struct := reflect.TypeOf(w)

	v_int := reflect.ValueOf(i)   //
	v_struct := reflect.ValueOf(w)

	tk_int := t_int.Kind()
	tk_struct := t_struct.Kind()
	vk_int := v_int.Kind()
	vk_struct := v_struct.Kind()

	vt_int := v_int.Type()
	vt_struct := v_struct.Type()

	fmt.Printf("   For int: type:%v, value:%v, kind of type:%v, kind of value:%v, type of value:%v \n", t_int, v_int, tk_int, vk_int, vt_int)
	fmt.Printf("For struct: type:%v, value:%v, kind of type:%v, kind of value:%v, type of value:%v \n",t_struct, v_struct, tk_struct, vk_struct, vt_struct)

	//结果：
	//1.如果是%t打印
	//   For int: type:&{%!t(uintptr=8) %!t(uintptr=0) %!t(uint32=4149441018) %!t(reflect.tflag=15) %!t(uint8=8) %!t(uint8=8) %!t(uint8=2) %!t(func(unsafe.Pointer, unsafe.Pointer) bool=0x4036c0) %!t(*uint8=0x4fea64) %!t(reflect.nameOff=879) %!t(reflect.typeOff=36224)}, value:%!t(int=3), kind of type:%!t(reflect.Kind=2), kind of value:%!t(reflect.Kind=2), type of value:*reflect.rtype
	//For struct: type:&{%!t(uintptr=8) %!t(uintptr=8) %!t(uint32=871609668) %!t(reflect.tflag=9) %!t(uint8=8) %!t(uint8=8) %!t(uint8=54) %!t(func(unsafe.Pointer, unsafe.Pointer) bool=0x4036c0) %!t(*uint8=0x4fea64) %!t(reflect.nameOff=7204) %!t(reflect.typeOff=0)}, value:&{%!t(*os.file=&{{{0 0 0} 576 {{0 0 0 0 0} 0 114 0 0 0xc000108280 <nil> {0 <nil>} {<nil> 0 <nil> 0 {0 <nil>} 0} <nil> <nil> 0 0 0 []} {{0 0 0 0 0} 0 119 0 0 0xc000108280 <nil> {0 <nil>} {<nil> 0 <nil> 0 {0 <nil>} 0} <nil> <nil> 0 0 0 []} {0} {0 0} [] [] [] 0 0 false true true true 4} /dev/stdout <nil> false})}, kind of type:%!t(reflect.Kind=22), kind of value:%!t(reflect.Kind=22), type of value:*reflect.rtype

	//2.如果是%T打印
	//   For int: type:*reflect.rtype, value:reflect.Value, kind of type:reflect.Kind, kind of value:reflect.Kind, type of value:*reflect.rtype
	//For struct: type:*reflect.rtype, value:reflect.Value, kind of type:reflect.Kind, kind of value:reflect.Kind, type of value:*reflect.rtype

	//3.如果是%v打印,  相当于直接使用println(实际上看该函数的内部实现, 就是增加的v这个arg)
	//   For int: type:int,      value:3,               kind of type:int, kind of value:int, type of value:int
	//For struct: type:*os.File, value:&{0xc000104000}, kind of type:ptr, kind of value:ptr, type of value:*os.File



	/*part 2:  断言与反射
	 value类型和interface{}类型都能持有任意的值, 不同的是:
	 interface{}: 一个空接口类型, 隐藏了值对应的表示方法和所有的公开方法, 因此只能断言成真正的类型才能访问其内部
	  value: 一个结构体类型, 可以有很多的内置方法, 进而可以通过这些内置的方法取得实际的值

	 */
	vInterface_int := v_int.Interface()   //返回一个interface{}类型的值
	reflect_int :=  vInterface_int.(int)
	vInterface_struct := v_struct.Interface()
	reflect_struct := vInterface_struct.(*os.File)


	fmt.Println(vInterface_int, reflect_int)      //结果:
	fmt.Println(vInterface_struct, reflect_struct)//结果:


}

type myStruct struct {
	id int
	name string
}

func TestMethod() {
	fmt.Println("TestMethod=>")
}