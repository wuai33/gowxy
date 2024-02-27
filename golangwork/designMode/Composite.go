package designMode

import (
	"fmt"
)
/*
参考链接：https://github.com/senghoo/golang-design-pattern/blob/master/13_composite/composite.go
组合模式统一对象和对象集，使得使用相同接口使用对象和对象集。
组合模式常用于树状结构，用于统一叶子节点和树节点的访问，并且可以用于应用某一操作到所有子节点。

wxy:
component: 成分; 部件; 组成部分;
composite: 复合材料; 合成物; 混合物;
这种模式一般用来规范行为的，即对于一些有从属关系树状逻辑，这种所谓的"从属关系"就可以抽象出来了

核心思想：
的当整体的架构是一个树状:有树根，树枝，树叶三部分组成，而他们相互之间有着共同的从属关系,
即树根 包含多个树枝，每个树枝上有多个树叶，整体是一个"发散"的形态，
其中，树根 和 树枝都是有下级的，即需要包含另一部分，所以认为是"集成部件"，
	 树叶没有下级，是一个"原子部件"
那么，就可以使用组合模式，将整个的行为关系抽象到一个接口上，然后不同的部件对应去实现


具体的实现步骤为：
1.定义"从属关系"的接口;
2.定义"从属关系"的结构体，实现对应的接口。其中完全可以复用的方法直接实现，需要定制化的则可以是虚方法的方式呈现;
3.根据在整个逻辑中的不同"定位"分别实现这种从属关系接口，在这里通过内嵌结构体的方式直接继承公共方法，然后分别实现特有方法;
4.实例化各个部位，然后利用关系函数(即"从属关系"中的方法)来和自己的"下属"或"上级"建立关系。
*/


//1. 定义"关联性"接口，规划了从属关系的整体形态
type Relevance interface {
	Parent() Relevance
	SetParent(Relevance)
	Name() string
	SetName(string)
	AddChild(Relevance)
	Print(string)
}

const (
	atomNode = iota   //"原子部件类型"
	compositeNode	  //"集成部件类型"
)

//2.定义"相关性"结构体, 除了实现"相关性"的接口外, 还具备两个元素：自己的名字 和 自己的上级
type relevance struct {
	parent Relevance
	name   string
}
func (c *relevance) Parent() Relevance {
	return c.parent
}
func (c *relevance) SetParent(parent Relevance) {
	c.parent = parent
}
func (c *relevance) Name() string {
	return c.name
}
func (c *relevance) SetName(name string) {
	c.name = name
}
//特色方法，待实现
func (c *relevance) AddChild(Relevance) {}
func (c *relevance) Print(string) {}

func NewComponent(kind int, name string) Relevance {
	var c Relevance
	switch kind {
	case atomNode:
		c = NewLeaf()
	case compositeNode:
		c = NewComposite()
	}

	c.SetName(name)
	return c
}

//3.具象化1: 原子部件类型的"叶子"结构体,  内嵌一个"关系"结构体
type Leaf struct {
	relevance
}
//其中，Print方法有override
func (c *Leaf) Print(pre string) {
	fmt.Printf("%s-%s\n", pre, c.Name())
}


func NewLeaf() *Leaf {
	return &Leaf{}
}



//3. 具象化2: 集成部件类型的"集成"结构体:  自身的"关系网"  + 子部件的"关系网"
//   重写AddChild方法和Print方法
type Composite struct {
	relevance
	childs []Relevance   //树枝  or  叶子
}
func (c *Composite) AddChild(child Relevance) {
	child.SetParent(c)
	c.childs = append(c.childs, child)
}
func (c *Composite) Print(pre string) {
	fmt.Printf("%s+%s\n", pre, c.Name())
	pre += " "
	for _, comp := range c.childs {
		comp.Print(pre)
	}
}

func NewComposite() *Composite {
	return &Composite{
		childs: make([]Relevance, 0),
	}
}


func TestComposite(){
	//1.创建四个"集成部件"，一个作为根, 其余作为树枝
	root := NewComponent(compositeNode, "root")
	branch1 := NewComponent(compositeNode, "branch1")
	branch2 := NewComponent(compositeNode, "branch2")
	branch3 := NewComponent(compositeNode, "branch3")

	//2.创建3个"原子部件"，均作为叶子
	leaf1 := NewComponent(atomNode, "leaf1")
	leaf2 := NewComponent(atomNode, "leaf2")
	leaf3 := NewComponent(atomNode, "leaf3")

	//3.为1个根添加2个树枝，为其中给1个树枝再添加1个树枝，为其中2个树枝分别添加1个和2个叶子
	root.AddChild(branch1)
	root.AddChild(branch2)
	branch1.AddChild(branch3)

	branch1.AddChild(leaf1)
	branch2.AddChild(leaf2)
	branch2.AddChild(leaf3)

	root.Print("")
}
