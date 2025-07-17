package mycontext

import (
	"context"
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)


func TestContextAndStopCH(t *testing.T) {


	// 一: Context的基本用法

	/*
	0. root context是没有cancel函数的, 他当然也就没有Done() channel

	1. 每次派生一个可cancel Context的时候, 返回的是一个派生context以及对应于这个context的cancel函数
	    wxy: 新派生的context还cancel函数是对应关系的, 不要认为这个cancel是cancel子context的.

	    派生过程:
	    首先, 初始化一个cancelCtx类型的结构体实例：c
	        1). 将parent记录在自己的context中
	        2). 绑定和parent的关系, 即propagate cancel逻辑
	            - 如果parent是不能被cancel的context, 那么也不需要建立关系
				- 如果parent是能被cancel的标准context, 那么注册上去
				- 如果parent是一个可被close的channel, 那么就spawn一个goroutine去监听这个channel

		然后, 构造自己的cancel函数(wxy: 我不是只能被parent级联cancel, 我是可以自己cancel的)
			核心逻辑如下：
				d, _ := c.done.Load().(chan struct{})
				if d == nil {
					c.done.Store(closedchan)
				} else {
					close(d)
				}
				...
				for child := range c.children {
					child.cancel(false, err, cause)
				}
			说明:
			    1. 结束自己：
				   如果done已经初始化过了, 说明有人在监听我啥时cancel了, 那么就close这个channel
			       如果还没有, 以防万一初始化一个已经close的channel给他
				2. 结束child:
				   调用child的cancel函数, 传入因自己cancel的原因

		最后, done的lazy 创建: 有人调用Done()了
			func (c *cancelCtx) Done() <-chan struct{} {
				d := c.done.Load()
				if d == nil {
					d = make(chan struct{})
					c.done.Store(d)
    			}
				...
			}
			说明: 如果没有初始化过就初始化要给chan struct{}类型的channel给.done并返回给调用者
	*/
	func () {
		rootCtx := context.Background()
		parentCtx, parentCancel := context.WithCancel(rootCtx)
		sonCtx, _ := context.WithCancel(parentCtx)
		go func() {
			<-parentCtx.Done()
			fmt.Println("Parent context cancelled")
		}()
		go func() {
			<-sonCtx.Done()
			fmt.Println("Son context cancelled")
		}()
		time.Sleep(1 * time.Second) // Give some time for the parent context to be set up
		fmt.Println("Cancelling parent context")
		parentCancel()

		// 结果:
		// Cancelling parent context
		// Parent context cancelled
		// Son context cancelled
		// 说明: parent的cancel函数被调用了, parent和son的Done() channel都被关闭了

	}()



	// 二: 使用stop channel结合context的用法
	//  Using a stop channel to simulate a Kubernetes context and a custom channel context
	// This test demonstrates how to use a stop channel to create a context that can be used
	// in a Kubernetes environment, allowing for graceful shutdowns and cancellations.package main
	func() {

		stop := make(chan struct{})

		// 基于stop这个channel创建的要给k8自定义的context: channelContext
		go func(stopCh <-chan struct{}) {
			k8sCtx := wait.ContextForChannel(stopCh)
			for {
				_, ok := <-k8sCtx.Done()
				fmt.Println("Receive from K8s context done, ok:", ok)
				if !ok {
					return
				}
			}
		}(stop)

		go func(stopCh <-chan struct{}){
			// time.Sleep(1 * time.Second)
			for {
				_, ok := <-stopCh
				fmt.Println("Received from stop channel, ok:", ok)
				if !ok {
					return
				}
			}
		}(stop)

		stop <- struct{}{}
		fmt.Println("Sent to stop channel successfully")
		close(stop)
		fmt.Println("Closed!")

		time.Sleep(1 * time.Second) // Give some time to see the output
		fmt.Println("------------------------------")

		/*
		Received from stop channel, ok: true
		Sent to stop channel successfully
		Closed!
		Receive from K8s context done, ok: false
		Received from channel, ok: false

		或者

		Received from stop channel, ok: true
		Sent to stop channel successfully
		Closed!
		Received from channel, ok: false
		Receive from K8s context done, ok: false

		或者

		Sent to stop channel successfully
		Closed!
		Receive from K8s context done, ok: true
		Receive from K8s context done, ok: false
		Received from stop channel, ok: false

		*/

	// 三: stop channel的best practice
	/*

		1. 通常一个channel用作stop 信号使用的场景, 只允许receive和close, 所以常规用法如下(参考signal包)：
		func SetupSignalHandler(sig ...os.Signal) (stopCh <-chan struct{}) {
			stop := make(chan struct{})

			go func() {
				close(stop)
			}()

			return stop
		}
		说明: 通过返回值就可以看出来这个channel就不应该向里面send, 就只能receive和close

	*/
	}()

}
