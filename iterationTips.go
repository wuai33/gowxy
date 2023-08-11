我们知道，在迭代的时候临时变量的声明周期仅限于当前次的迭代，所以如果使用临时变量作为函数的入参要千万小心
尤其是在函数中还有取地址的操作，因为很可能这个地址对应的值在迭代的过程中会被不断刷新
针对迭代中的临时变量设计了三种测试场景
场景1： 调用函数是结构体类型，临时变量作为值传递给函数
       此时函数入参完全得到一份值的拷贝, 于是可以放心使用，也就是取地址，
       因为取的是入参的，只要函数一次调用他就是一个独立的变量.
场景2： 调用函数是指针类型，临时变量的地址作为值传递给函数
       此时函数入参得到的是一个指向临时变量的地址，此时不能直接使用,
       因为一旦临时变量中的值变化了, 入参指向的内容也变化了.
场景3： 调用函数是指针类型，临时变量的地址作为值传递给函数，但是在函数中会将入参指向的内容拷贝一份再使用
       此时函数入参得到的是一个指向临时变量的地址，此时不能直接使用, 原因同上
       但是如果把要使用的内容拷贝一份，则就不怕临时变量因为迭代的原因发生变化了.
       
type SrcStruct struct {
	File                   string `json:"file"`
	ChartOverrideReference string `json:"chartOverrideReference,omitempty"`
}
type DstStruct struct {
	ChartOverrideReference *string // optional: "aic-vdu-0.300.13792.tgz"
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
