package myInterface

import (
	"fmt"

)

type Store interface {
	Context() int
}
type Defaults struct {
	ID  int
	Store         Store
}

func TestInterface(){
	myDefaults := Defaults{ID:1}
	if myDefaults.Store == nil {
		fmt.Println("myDefaults.Store is nil")
	}else{
		fmt.Println("myDefaults.Store is not nil")
	}



}