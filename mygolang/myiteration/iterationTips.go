package myiteration

import (
	"fmt"
	"testing"
)

type SrcStruct struct {
	File                   string `json:"file"`
	ChartOverrideReference string `json:"chartOverrideReference,omitempty"`
}

//////////////////////////////////////////Part 1//////////////////////////////////////////
// 我们知道, 在迭代的时候临时变量的声明周期仅限于当前次的迭代，
// 但是如果在当前的迭代周期内将元素append到新的数组(slice)中, 会不会发生指向错误呢
// 结果： 任何情况下，即无论被append的是值类型，还是指针类型, 在append的当下，都是将元素的一个分完全拷贝添加到新的slice中
//       所以, 即使是脱离了迭代的作用域, 已经append的内容并不会受影响.

type Profile struct {
	NodeConfigName string
	Identity       SrcStruct
}

func TestAppendRannicProfile(t *testing.T) {
	resetProfiles1 := make([]Profile, 0)
	resetProfiles2 := make([]Profile, 0)
	resetProfiles3 := make([]*Profile, 0)
	profile1 := Profile{
		Identity: SrcStruct{
			File: "1.1.1.1",
		},
		NodeConfigName: "node1",
	}
	profile2 := Profile{
		Identity: SrcStruct{
			File: "2.2.2.2",
		},
		NodeConfigName: "node1",
	}

	profile3 := Profile{
		Identity: SrcStruct{
			File: "3.3.3.3",
		},
		NodeConfigName: "node2",
	}

	// 将值数组的元素append到值数组中，每个元素都是完全的一份拷贝, 不会发生指向错误
	ranNicDevices1 := []Profile{profile1, profile2, profile3}

	for _, candidate := range ranNicDevices1 {
		resetProfiles1 = append(resetProfiles1, candidate)
	}

	// 结果：
	// p1:  {node1 {1.1.1.1 }}
	// p1:  {node1 {2.2.2.2 }}
	// p1:  {node2 {3.3.3.3 }}
	for _, p := range resetProfiles1 {
		fmt.Println("p1: ", p)
	}

	// 将指针数组的元素append到值数组中，每个元素得到的也是一份完全的拷贝，不会发生指向错误
	ranNicDevices2 := []*Profile{&profile1, &profile2, &profile3}
	for _, candidate := range ranNicDevices2 {
		resetProfiles2 = append(resetProfiles2, *candidate)
	}

	// 结果：
	// p2:  {node1 {1.1.1.1 }}
	// p2:  {node1 {2.2.2.2 }}
	// p2:  {node2 {3.3.3.3 }}
	for _, p := range resetProfiles2 {
		fmt.Println("p2: ", p)
	}

	// 将指针数组的元素append到指针数组中，append的也是一份完全的拷贝， 不会发生指向错误
	for _, candidate := range ranNicDevices2 {
		resetProfiles3 = append(resetProfiles3, candidate)
	}

	// 结果：
	// p3:  &{node1 {1.1.1.1 }}
	// p3:  &{node1 {2.2.2.2 }}
	// p3:  &{node2 {3.3.3.3 }}
	for _, p := range resetProfiles3 {
		fmt.Println("p3: ", p)
	}

}

///////////////////////////////////////////Part 2////////////////////////////////////////////
// 我们知道，在迭代的时候临时变量的声明周期仅限于当前次的迭代，所以如果使用临时变量作为函数的入参要千万小心
// 尤其是在函数中还有取地址的操作，因为很可能这个地址对应的值在迭代的过程中会被不断刷新
// 针对迭代中的临时变量设计了三种测试场景
// 场景1： 调用函数是结构体类型，临时变量作为值传递给函数
//        此时函数入参完全得到一份值的拷贝, 于是可以放心使用，也就是取地址，
//        因为取的是入参的，只要函数一次调用他就是一个独立的变量.
// 场景2： 调用函数是指针类型，临时变量的地址作为值传递给函数
//        此时函数入参得到的是一个指向临时变量的地址，此时不能直接使用,
//        因为一旦临时变量中的值变化了, 入参指向的内容也变化了.
// 场景3： 调用函数是指针类型，临时变量的地址作为值传递给函数，但是在函数中会将入参指向的内容拷贝一份再使用
//        此时函数入参得到的是一个指向临时变量的地址，此时不能直接使用, 原因同上
//        但是如果把要使用的内容拷贝一份，则就不怕临时变量因为迭代的原因发生变化了.

type DstStruct struct {
	ChartOverrideReference *string
}

func TestCnfPkgManifestData(t *testing.T) {
	data1 := SrcStruct{
		File:                   "test1",
		ChartOverrideReference: "referChart1",
	}
	data2 := SrcStruct{
		File:                   "test2",
		ChartOverrideReference: "referChart2",
	}
	data3 := SrcStruct{
		File:                   "test3",
		ChartOverrideReference: "referChart3",
	}
	testDatas := []SrcStruct{data1, data2, data3}
	testCharts1 := make([]DstStruct, 0)
	testCharts2 := make([]DstStruct, 0)
	testCharts3 := make([]DstStruct, 0)

	for _, d := range testDatas {
		var chart DstStruct
		set1(&chart, d)
		testCharts1 = append(testCharts1, chart)
	}
	// test 1: referChart1
	// test 1: referChart2
	// test 1: referChart3
	for _, c := range testCharts1 {
		fmt.Printf("test 1: %s \n", *c.ChartOverrideReference)
	}

	for _, d := range testDatas {
		var chart DstStruct
		set2(&chart, &d)
		testCharts2 = append(testCharts2, chart)
	}
	// test 2: referChart3
	// test 2: referChart3
	// test 2: referChart3
	for _, c := range testCharts2 {
		fmt.Printf("test 2: %s \n", *c.ChartOverrideReference)
	}

	for _, d := range testDatas {
		var chart DstStruct
		set3(&chart, &d)
		testCharts3 = append(testCharts3, chart)
	}
	// test 3: referChart1
	// test 3: referChart2
	// test 3: referChart3
	for _, c := range testCharts3 {
		fmt.Printf("test 3: %s \n", *c.ChartOverrideReference)
	}
}

func set1(chart *DstStruct, chartData SrcStruct) {
	chart.ChartOverrideReference = &chartData.ChartOverrideReference
}

func set2(chart *DstStruct, chartData *SrcStruct) {
	chart.ChartOverrideReference = &chartData.ChartOverrideReference
}

func set3(chart *DstStruct, chartData *SrcStruct) {
	tmpData := chartData.ChartOverrideReference
	chart.ChartOverrideReference = &tmpData
}
