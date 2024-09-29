package channelandgoroutine

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)



func counter(out chan<- int) {
	for x := 0; x < 100; x++ {
		out <- x    //将1-100的值依次填入out channel中
	}
	close(out)
}

func squarer(out chan<- int, in <-chan int) {
	for v := range in {   //从in中不断读取数据
		out <- v * v      //计算后添加到out中
	}
	//close(out)
}

func printer(in <-chan int) {
	for v := range in {       fmt.Println(v)    }
	fmt.Print("print over!")
}

type Manager struct {
	internalStop        <-chan struct{}
	internalStopper     chan<- struct{}
}

func NewManager() *Manager {
	stop := make(chan struct{})
	return &Manager{
		internalStop:          stop,
		internalStopper:       stop,
	}
}

type Monitor struct {
	clustersLock sync.RWMutex
	managerStopperSetter bool
	managerStopper     chan<- struct{}
}


func (mo *Monitor) SetMGRStopper(stopper chan<- struct{}) {
	mo.clustersLock.Lock()
	defer mo.clustersLock.Unlock()
	mo.managerStopperSetter = true
	mo.managerStopper = stopper
}


func (mo *Monitor) StopMGRStopper() {
	mo.clustersLock.Lock()
	defer mo.clustersLock.Unlock()
	if mo.managerStopperSetter {
		close(mo.managerStopper)
		mo.managerStopperSetter = false
	}
}

func testChannelClose(){
	test := make(chan int)    //元素是结构体类型Monitor,
	go func(send chan<- int) {
		send <- 1
		send <- 0
		send <- 2
		close(send)
	}(test)

	go func(receive <-chan int) {
		for i := 0; i< 5; i++ {
			v, ok := <-receive
			fmt.Printf("value:%v, ok:%v\n",v, ok)
		}
	}(test)
}


func TestChannel(){
	testChannelClose()
	/*
	naturals := make(chan int)      //双向channel
	squares := make(chan int)       //双向channel
	go counter(naturals)
	go squarer(squares, naturals)
	go printer(squares)

	var monitor Monitor

	systemStopCh := SetupSignalHandler()
	go func(stop <-chan struct{}){
		for {
			time.Sleep(time.Second * 2)   //4.等待
			m := NewManager()             //1.5先初始化manager，并将stopper交给monitor
			//monitor.managerStopper = m.internalStopper
			monitor.SetMGRStopper(m.internalStopper)

			//阻塞着，知道外面的Stopper
			select {
			case <-m.internalStop:        //3.被关闭，重新启动一次循环
				fmt.Println("Receive internal stop signal received, so begin to another [crd contrller] loop!")
			case <-stop:
				fmt.Println("Receive stop signal received, so close crd controller and return!")
			}
		}
	}(systemStopCh)

	go func(){
		time.Sleep(time.Second * 3)   //2.monitor关闭channel
		//close(monitor.managerStopper)
		monitor.StopMGRStopper()

		time.Sleep(time.Second * 1)
		//close(monitor.managerStopper)  //5.再次关闭
		monitor.StopMGRStopper()

		time.Sleep(time.Second * 2)
		//close(monitor.managerStopper)  //5.再次关闭
		monitor.StopMGRStopper()

	}()


	close(squares)

	 */
	for{
		time.Sleep(time.Second * 600)   //10min后
		break
	}
}
var onlyOneSignalHandler = make(chan struct{})
var shutdownSignals = []os.Signal{os.Interrupt}

func SetupSignalHandler() (stopCh <-chan struct{}) {
	//wxy: 为了确保一个main进程中只有一个系统级别stop方法，用这个操作来保证，
	//     因为如果有多处调用本方法则会因为重复close channel导致失败， 很巧妙
	close(onlyOneSignalHandler) // panics when called twice

	//初始化一个双向无缓存channel, 作为实现stop的信道
	stop := make(chan struct{})
	c := make(chan os.Signal, 2)          //再初始化一个可以存放2个元素(系统信号类型)的channel
	signal.Notify(c, shutdownSignals...)  //双向channel此时会在传递过程中被强壮成单向：只接收数据
	go func() {
		<-c   		   //第一次从channle中读取数据，如果成功说明产生了我监听的信号
		close(stop)    //于是, 关闭目标channel, 即stop channel
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop   //虽然是双向，但是经过返回值的强转，即外面结束到本channel后，只能从channel中读取数据
}