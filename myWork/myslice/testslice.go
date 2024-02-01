package myslice

import (
	"fmt"
	"time"
)

type DataSourceRef struct {
	ID           int  `json:"id"`
	Name         string `json:"name"`
	Derived    	 bool   `json:"isderived"`
}


var cache *DataSourceCache
type DataSourceCache struct {
	//*sync.Mutex    //因为还有同步线程
	DataSourcesV   []DataSourceRef     //以namespace为key, 每个namespace下的
	DataSourcesP   []*DataSourceRef     //以namespace为key, 每个namespace下的
}

func InitDataSourceCache() {
	cache = &DataSourceCache{
		//Mutex: &sync.Mutex{},
		DataSourcesV: []DataSourceRef{},
		DataSourcesP: []*DataSourceRef{},
	}
}

func goroutine_add(){
	fmt.Println("goroutine 1: write")

	for i := 0; i < 20; i++ {
		cache.DataSourcesP = append(cache.DataSourcesP,  &DataSourceRef{
			ID:	i,
			Name:       		fmt.Sprintf("item-%d", i),
		})

		cache.DataSourcesV = append(cache.DataSourcesV,  DataSourceRef{
			ID:	i,
			Name:       		fmt.Sprintf("item-%d", i),
		})


	}

	fmt.Println("add result:")
	for _, v := range  cache.DataSourcesP{
		fmt.Println("item after add:%v",v)
	}

}

func goroutine_get1(){
	fmt.Println("goroutine : get1")

	start := false
	for ; ;  {
		get := []*DataSourceRef{}
		for _, v := range cache.DataSourcesP {
			get = append(get, v)
		}

		length :=  len(get)
		if length <= 0 && start{
			break
		}
		if length > 1 {
			start = true
		}
		fmt.Println("get1 result:", length)
		for _, v := range get {
			fmt.Println("item of get1:%v",v)
		}


		time.Sleep(100)
	}

}

func goroutine_get2(){
	fmt.Println("goroutine: get2")

	for ; ;  {

		for _, v := range cache.DataSourcesP {
			fmt.Printf("item of get2:%v, name:%s, id:%d\n",v, v.Name, v.ID )
			if cache.DataSourcesP[19].Name == "changed"{
				break
			}
		}

		time.Sleep(100)
	}

}
func goroutine_change(){
	fmt.Println("goroutine: change")

	for j:=0; j<10000 ; j++  {
		length :=  len(cache.DataSourcesP)

		if length > 10 {
			for i, v := range cache.DataSourcesP {
				if i > 10 {
					v.Name = fmt.Sprintf("changed-%d", j)
					v.ID = j*100
					fmt.Println("change:",v)
				}
				time.Sleep(10)
			}
		}


	}

}

func goroutine_change2(){
	fmt.Println("goroutine: change2")

	for j:=0; j<10000 ; j++  {
		length :=  len(cache.DataSourcesP)

		if length > 15 {
			for i, v := range cache.DataSourcesP {
				if i > 10 {
					v.Name = fmt.Sprintf("changed2-%d-1", j)
					v.ID = j*100+1
					fmt.Println("change2:",v)
				}
				time.Sleep(12)
			}
		}


	}

}


func goroutine_delete(i string){
	fmt.Println("goroutine 3: delete\n")

	for ;;{
		lenght := len(cache.DataSourcesP)
		fmt.Println("now length for delete:" + i, lenght)

		if lenght <= 0 {
			break
		}
		index := lenght/3
		list := cache.DataSourcesP
		list[index] = list[len(list)-1]   //用最后一个填充
		list[len(list)-1] = nil //把最后一个置空

		list = list[:len(list)-1]

		cache.DataSourcesP = list

		if index > 2 {
			cache.DataSourcesP[index-2].Name = "test"
		}


		fmt.Println("delete result:",)

		for _, v := range  cache.DataSourcesP{
			fmt.Println("item of delete "+ i, v)
		}

		time.Sleep(100)
	}


}

func TestMySlice(){
	InitDataSourceCache()

	goroutine_add()

	go goroutine_change2()
	go goroutine_get2()
	go goroutine_change()


	//go goroutine_delete("one")
	//go goroutine_delete("two")

	//part 1: 遍历的过程中修改内容

	/*
	for i, v := range  cache.DataSourcesP{
		 if i == 3 {
			cache.DataSourcesP[i].Name = "test for 3 of direct with value "
		}

		if i == 6 {
			v.Name = "test for 6 of iteration with value "
		}

	}

	 */


/*	list := cache.DataSourcesP

	list[5] = list[len(list)-1]
	list = list[:len(list)-1]
	cache.DataSourcesP = list


	fmt.Println("result:")
	for _, v := range  cache.DataSourcesP{
		fmt.Println("item of result:",v)
	}
*/
	for ; ;  {
		time.Sleep(1000)
	}
}