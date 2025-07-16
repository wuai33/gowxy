
【结论】：

【直接错误】:
	panic: context: internal error: missing cancel error

	goroutine 143 [running]:
	context.(*cancelCtx).cancel(0x23f29c8?, 0x90?, {0x0?, 0x0?}, {0x0?, 0x0?})
		/usr/local/go/src/context/context.go:542 +0x2c5
	context.(*cancelCtx).propagateCancel.func2()
		/usr/local/go/src/context/context.go:516 +0xf1
	created by context.(*cancelCtx).propagateCancel in goroutine 181
		/usr/local/go/src/context/context.go:513 +0x3e6

【直接原因】
	// cancel closes c.done, cancels each of c's children, and, if
	// removeFromParent is true, removes c from its parent's children.
	// cancel sets c.cause to cause if this is the first time c is canceled.
	func (c *cancelCtx) cancel(removeFromParent bool, err, cause error) {
		if err == nil {
			panic("context: internal error: missing cancel error")
		}
		...

	func (c *cancelCtx) propagateCancel(parent Context, child canceler) {
		....

		1. 如果Done()是nil, 表示parent没有用于concel的channel, 所以肯定就是不能cancel的
		   那么, 也就不需要propagate cancel到当前context中来


		2. 如果pareant是可cancel的, 那么要把自己的会调添加到parent的child中
		   之后, parent cancel的时候也要记得childs
		// 如果context包含一个key是int类型, value是concelCtx
		p, ok := parent.Value(&cancelCtxKey).(*cancelCtx)
		....

		3. parent.(afterFuncer)


		4. 除此之外, parent并不是一个标准的可cancel context, 但是有Done(), 表示也是可以通知结束的
		   所以, 这里就额外启动了一个goroutine来监听这个parent是否close了
		   !!!!重要: bug也就在这里产生了
			go func() {
				select {
					case <-parent.Done():  --- Done函数返回的是.stopCh这个channel, 能够select出来表示channel中有内容了
						child.cancel(false, parent.Err(), Cause(parent))
					case <-child.Done():
					}
				}()
			}

	 根据如上的调用栈, 可以确定问题的直接原因是由于parent context的.Err()返回了空

【根本原因】:
 child context错误认为parent context已经cancel, 于是在关闭自己context的时候没有获取到parent context从cancel的原因

 具体分析：
 0. 在lcm的程序中, 一个stuct{}类型的空channel stop初始化后, 传递给了三个不同gouroutine处理

	1) 用于作为informter的channelContext的base
		go informer.Run(stop)
		...
		func (s *sharedIndexInformer) Run(stopCh <-chan struct{}) {
			s.RunWithContext(wait.ContextForChannel(stopCh))
		}

		type channelContext struct {
			stopCh <-chan struct{}
		}
		func (c channelContext) Done() <-chan struct{} { return c.stopCh }
		func (c channelContext) Err() error {
			select {
			case <-c.stopCh: ---如果还能从里面获取到内容, 说明是被cancel掉了, 所以就返回Cancel, 但由于stopCh是0缓存, 然后已经被获取出来了, 所以不会再继续被获取出来,  自然走default分支
				return context.Canceled
			default:
				return nil
			}
		}

		// watchWithResync runs watch with startResync in the background.
		func (r *Reflector) watchWithResync(ctx context.Context, w watch.Interface) error {
			cancelCtx, cancel := context.WithCancel(ctx)
			...
		}

		说明:
		Informer(k8s)支持传入一个chan struct{}类型的channel, 然后将其封装到一个自定义(k8s)的context中
        Informer内部的reflector还会据此派生可cancel类型的子context, 根据propagateCancel函数的实现原理,
		会启动一个goroutine专门监听parent的Done()

	2) 应用层代码中注册了handler函数, 用于monitor指定cr, 一旦发现cr符合指定条件就会添加内容到channel中:
		func (r *RANCNFReconciler) checkStatusAndSignal(obj interface{}, stopCh, engineRestart chan struct{}, action string) {
			...
			stopCh <- struct{}{}
		}

	3) 业务代码同时还启动了一个for循环, 不断从channel中获取内容, 一旦取得则关闭该channel
		defer close(stop)
		...
		func (r *RANCNFReconciler) handleRetryWithTimeout(...) {
			for {
				select {
					...
				case <-stop:
					log.Infof("Operation stopped for CR %s/%s", log.Scramble(namespacedName.Namespace), namespacedName.Name)
					return
				}
			}
		}

1. 当第一次超时的时候, for循环主要focus在了定时器超时这个分支上

2. 当install流程已经完成, 触发添加元素到stopCh中, 但是大循环没有能够及时捕获stop channel中的内容
   从而结束大循环, 执行stop channel的close

3. 最终导致这个消息被context捕获了,
   在设计上, context的逻辑是只要能够从parent.Done()中select出来, 就表示是parent cancel/close了,
   于是进入child的cancel流程, 这个流程中会调用parent的Error, 而这个k8s提供的context
   但实际上, 能够select出来的原因并不是因为channel close掉, 而是因为channel中有item
   这个channel是0缓存, 所以调用Err()的时候, 直接返回nil了

```
    func (c channelContext) Done() <-chan struct{} { return c.stopCh }
	func (c channelContext) Err() error {
		select {
		case <-c.stopCh: ---如果还能从里面获取到内容, 说明是被cancel掉了, 所以就返回Cancel, 但由于stopCh是0缓存, 然后已经被获取出来了, 所以不会再继续被获取出来,  自然走default分支
			return context.Canceled
		default:
			return nil
		}
	}
```

4. 于是panic了, 因为golang的context是不允许的, 关键代码如下：
```
	// cancel closes c.done, cancels each of c's children, and, if
	// removeFromParent is true, removes c from its parent's children.
	// cancel sets c.cause to cause if this is the first time c is canceled.
	func (c *cancelCtx) cancel(removeFromParent bool, err, cause error) {
		if err == nil {
			panic("context: internal error: missing cancel error")
		}
		...
```

6. 为什么非必现
   1) 如果stop channel中的顺利被大循环捕获, 那么就可以结束大循环, 并close掉这个channel
   2) 发生了重试后, 拖慢了大循环的再次select



==================debug过程与发现===================================
经过debug, 这个问题的直接原因确实是stop还没有被cancel, 但是informer已经进入cancel时代, 直接错误如下：

```
// propagateCancel arranges for child to be canceled when parent is.
// It sets the parent context of cancelCtx.
func (c *cancelCtx) propagateCancel(parent Context, child canceler) {
	c.Context = parent

	done := parent.Done()
	...

	goroutines.Add(1)
	go func() {
		select {
		case <-parent.Done():  --- Done函数返回的是.stopCh这个channel, 能够select出来表示channel中有内容了
			child.cancel(false, parent.Err(), Cause(parent))
		case <-child.Done():
		}
	}()
}

```

说明:
1. Parent是如下这个类型
```
    type channelContext struct {
        stopCh <-chan struct{}
    }

    func (c channelContext) Done() <-chan struct{} { return c.stopCh }
```

2. Parent这个类型为什么错误信息为nil
```
    func (c channelContext) Err() error {
        select {
        case <-c.stopCh: ---如果还能从里面获取到内容, 说明是被cancel掉了, 所以就返回Cancel,
                            但由于stopCh是0缓存, 然后已经被获取出来了, 所以不会再继续被获取出来,  自然走default分支
            return context.Canceled
        default:
            fmt.Println("Not fetch from .stopCh(not stop), I will return nil")
            return nil
        }
    }
```

wxy: 但是这里有个问题, 为什么第一次没有问题，并且是非毕现的
     分析这个原因是: 如果不是因为有uninstall的goroutine进来打断接下来的从stop中select, 那么压根不会走到propagateCancel的go func中去
	 我分析这个函数的作用是, propagate的时候，除了新建一个context, 还要启动一个goroutine去监听parent是否结束
	 是否结束的标志就是能否从channel中select出来内容(golang的语法： 所有从关闭状态中读取时都会立即返回)，
	 通常

debug, 增加debug信息如下：
```
	go func() {
		select {
		case info, ok := <-parent.Done():
			fmt.Printf("====> Got info %v from parent context, ok?: %v\n", info, ok)
			child.cancel(false, parent.Err(), Cause(parent))
		case <-child.Done():
		}
	}()
```

Bad:

```
14>1 2025-07-15T15:18:20.519810Z [INFO] Event(Update) from monitor informer for Install RANCNF <rcp-log>default</rcp-log>/simulatorcnf(ResourceVersion from api reader: 5335846, ResourceVersion from informer: 5335846, Operation ID from initStaus: , Operation ID from informer: )
<14>1 2025-07-15T15:18:20.519843Z[INFO] Check status for CR <rcp-log>default</rcp-log>/simulatorcnf: Succeed
<14>1 2025-07-15T15:18:20.519865Z [INFO] Operation Install is not ongoing but Succeed, will stop this monitor informer
====> Got info {} from parent context, ok?: true
Not fetch from .stopCh(not stop), I will return nil
Not fetch from .stopCh(not stop), I will return nil
<14>1 2025-07-15T15:18:20.520200Z [INFO] Event(Update) from monitor informer for Install RANCNF <rcp-log>default</rcp-log>/simulatorcnf(ResourceVersion from api reader: 5335847, ResourceVersion from informer: 5335847, Operation ID from initStaus: , Operation ID from informer: )

```

Good：
```
cp-log>/simulatorcnf: Succeed
<14>1 2025-07-15T15:13:22.698753Z [INFO] Operation Install is not ongoing but Succeed, will stop this monitor informer
<14>1 2025-07-15T15:13:22.698769Z [INFO] Operation stopped for CR <rcp-log>default</rcp-log>/simulatorcnf
====> Got info {} from parent context, ok?: false
<14>1 2025-07-15T15:13:22.857734Z [INFO] CR <rcp-log>default</rcp-log>/simulatorcnf come in reconcile with resourceVersion: 5335500
<14>1 2025-07-15T15:13:22.857779Z [INFO] CR <rcp-log>default</rcp-log>/simulatorcnf is in terminating

```



=====================代码调用栈=============================
```
func (r *RANCNFReconciler) startActionTimer(
	cancel context.CancelFunc,
	namespacedName types.NamespacedName,
	ctx context.Context,
	actionTimeout time.Duration,
	action string,
	status rancnfv1alpha1.RANCNFConditionType,
	actionFunc func(context.Context, ctrl.Request, *rancnfv1alpha1.RANCNF) error,
) {
	stop := make(chan struct{})
	engineRestart := make(chan struct{})
	actionTimer := r.initializeActionTimer(actionTimeout)
	log.Infof("Start action(%s) timer with timeout %v for CR %s/%s", action, actionTimeout, log.Scramble(namespacedName.Namespace), namespacedName.Name)
	go r.monitorAndNotifyRANCNFChanges(stop, engineRestart, namespacedName.Namespace, action, ctx)
	defer close(stop)
	defer close(engineRestart)
	defer cancel()
	defer actionTimer.Stop()
	r.handleRetryWithTimeout(ctx, namespacedName, actionTimeout, action, status, actionFunc, actionTimer, stop, engineRestart)
}

```

分析:
stop：在这里初始化的stop channel, 一旦handleRetryWithTimeout执行完毕就会就会被stop
engineRestart： 在这里初始化的stop channel, 一旦handleRetryWithTimeout执行完毕就会就会被stop
ctx/cancel: 是上一层传进来的, 一旦handleRetryWithTimeout执行完毕就会就会被执行cancel, 如此依赖ctx.Done()就会被select出来内容


```
func (r *RANCNFReconciler) monitorAndNotifyRANCNFChanges(stop, engineRestart chan struct{}, namespace, action string, ctx context.Context) {
	...
	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			r.handleInformerAddEvent(obj, stop, engineRestart, action, ctx)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			r.handleInformerUpdateEvent(newObj, stop, engineRestart, action)
		},
		DeleteFunc: func(obj interface{}) {
			r.handleInformerDeleteEvent(obj, stop, engineRestart, action)
		},
	})
	if err != nil {
		log.Errorf("Failed to add event handler: %v", err)
		stop <- struct{}{}
	}

	go informer.Run(stop)
}

```
--------------------------------

```
func (r *RANCNFReconciler) handleInformerUpdateEvent(newObj interface{}, stop chan struct{}, engineRestart chan struct{}, action string) {
	resource := newObj.(*unstructured.Unstructured)
	var rancnf rancnfv1alpha1.RANCNF
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(resource.Object, &rancnf)
	if err == nil {
		log.Infof("Event(Update) from monitor informer for %s RANCNF %s/%s(ResourceVersion: %s)", action, log.Scramble(rancnf.Namespace), rancnf.Name, rancnf.ResourceVersion)
		r.checkStatusAndSignal(rancnf, stop, engineRestart, action)
	} else {
		log.Errorf("Failed to convert unstructured to RANCNF: %v", err)
	}
}

```

```

func (r *RANCNFReconciler) checkStatusAndSignal(obj interface{}, stopCh, engineRestart chan struct{}, action string) {
	cr := obj.(rancnfv1alpha1.RANCNF)
	status := cr.Status.Operation.Status
	log.Infof("Check status for CR %s/%s: %s", log.Scramble(cr.Namespace), cr.Name, status)
	if status != rancnfv1alpha1.OperationOngoing && status != rancnfv1alpha1.OperationEngineShutDown {
		stopCh <- struct{}{}
		log.Infof("Operation %s is not ongoing but %s, will stop this monitor informer", action, status)
		return
	}

	if status == rancnfv1alpha1.OperationEngineShutDown {
		log.Infof("Operation status indicates engine shut down, restarting operation %s", action)
		cr.Status.Operation.Status = rancnfv1alpha1.OperationOngoing
		if err := r.updateStatus(context.Background(), &cr, cr.Status); err != nil {
			log.Errorf("Failed to update status for CR(resourceVersion: %s) %s/%s: %v", cr.ResourceVersion, log.Scramble(cr.Namespace), cr.Name, err)
			stopCh <- struct{}{}
			return
		}
		engineRestart <- struct{}{}
	}
}

```
-------------------------------------------------

```
func (r *RANCNFReconciler) handleRetryWithTimeout(
	ctx context.Context,
	namespacedName types.NamespacedName,
	actionTimeout time.Duration,
	action string,
	status rancnfv1alpha1.RANCNFConditionType,
	actionFunc func(context.Context, ctrl.Request, *rancnfv1alpha1.RANCNF) error,
	actionTimer *time.Timer,
	stop, engineShutdown chan struct{},
) {
	hasRetried := false

	for {
		select {
		case <-actionTimer.C:
			log.Infof("Action %s timeout for CR %s/%s", action, log.Scramble(namespacedName.Namespace), namespacedName.Name)
			if !r.handleRetry(ctx, namespacedName, action, actionFunc, status, &hasRetried) {
				return
			}
			r.resetTimer(actionTimer, actionTimeout)

		case <-engineShutdown:
			log.Infof("Engine shutdown with graceful termination for CR %s/%s", log.Scramble(namespacedName.Namespace), namespacedName.Name)
			if !r.handleRetry(ctx, namespacedName, action, actionFunc, status, &hasRetried) {
				return
			}
			r.resetTimer(actionTimer, actionTimeout)

		case <-ctx.Done():
			log.Infof("Context cancelled for CR %s/%s", log.Scramble(namespacedName.Namespace), namespacedName.Name)
			return

		case <-stop:
			log.Infof("Operation stopped for CR %s/%s", log.Scramble(namespacedName.Namespace), namespacedName.Name)
			return
		}
	}
}

```

=====================  完整log以及分析 =======================
```
<14>1 2025-07-12T19:36:35.193282Z [INFO] Get latest CR <rcp-log>default</rcp-log>/simulatorcnf(resourceVersion: 7940) from controller-runtime with Phase: Preparing, CNFBuild: , Condition: [{Installed False 0 2025-07-12 19:36:33 +0000 UTC Instantiating Helm install simulatorcnf-cluster-preparation(Revision: 1, Chart: simulatorcnf-cluster-preparation-1.0.0, mainChart: false) successfully, operation id: 1752348972530665423} {Tested False 0 2025-07-12 19:36:12 +0000 UTC PrepareInstallation -} {Prepared True 0 2025-07-12 19:36:32 +0000 UTC Prepared Prepare CNF successfully} {Ready Unknown 0 2025-07-12 19:36:12 +0000 UTC PrepareInstallation -}]
<14>1 2025-07-12T19:36:35.204783Z [INFO] Update CR <rcp-log>default</rcp-log>/simulatorcnf(resourceVersion: 7949) status with Phase: Preparing, cnfSWBuild:  successfully
<14>1 2025-07-12T19:36:35.205656Z [INFO] Event(Update) from monitor informer for Install RANCNF <rcp-log>default</rcp-log>/simulatorcnf(ResourceVersion from api reader: 7949, ResourceVersion from informer: 7949, Operation ID from initStaus: 1752348972530665423, Operation ID from informer: 1752348972530665423)
<14>1 2025-07-12T19:36:35.205685Z [INFO] Check status for CR <rcp-log>default</rcp-log>/simulatorcnf: Ongoing

/// 设置install的超时事件很短, 导致超时重试了, 此后死循环会在如下的分支有一个很长事件的操作：
///      case <-actionTimer.C: ....
/// 然后, 重置了定时器
<14>1 2025-07-12T19:37:02.543533Z [INFO] Action Install timeout for CR <rcp-log>default</rcp-log>/simulatorcnf
<14>1 2025-07-12T19:37:02.551397Z [INFO] Handle action Install retry after 13ss, CR <rcp-log>default</rcp-log>/simulatorcnf with phase: Preparing, operation status: Ongoing
<14>1 2025-07-12T19:37:15.552446Z [INFO] Previous operation incomplete after timeout, continue to handle it, operation type=Install, CR: <rcp-log>default</rcp-log>/simulatorcnf
<14>1 2025-07-12T19:37:15.552505Z [INFO] Start Query CNF for operation Install on CR: <rcp-log>default</rcp-log>/simulatorcnf
<14>1 2025-07-12T19:37:15.552695Z [INFO] Waiting for Service:<rcp-log>cnf-lcm-operator</rcp-log>/cnf-config-manager-25r1-1-0-0-2-gdfa56bc6 for asset cnf-config-manager-25r1-1-0-0-2-gdfa56bc6 until ready
<14>1 2025-07-12T19:37:15.556120Z [INFO] Get service:cnf-config-manager-25r1-1-0-0-2-gdfa56bc6, IP=10.104.6.55, port=[{httpserver TCP <nil> 8080 {0 8080 } 0}], selector=map[app:cnf-config-manager-25r1-1-0-0-2-gdfa56bc6], status=&ServiceStatus{LoadBalancer:LoadBalancerStatus{Ingress:[]LoadBalancerIngress{},},Conditions:[]Condition{},}
<14>1 2025-07-12T19:37:15.556526Z [INFO] cnf-config-manager-25r1-1-0-0-2-gdfa56bc6 for CNFBuild vCUCNF25R1_1.0.0-2-gdfa56bc6 is running. It can be reused
<14>1 2025-07-12T19:37:15.556573Z [INFO] Increment cnf-config-manager-25r1-1-0-0-2-gdfa56bc6 SW build usage for namespace <rcp-log>default</rcp-log>
<14>1 2025-07-12T19:37:15.556581Z [INFO] Request SW build vCUCNF25R1_1.0.0-2-gdfa56bc6 for CR <rcp-log>default</rcp-log>/simulatorcnf successfully
<14>1 2025-07-12T19:37:15.556691Z [INFO] ImageRegistry is not specified in CR, using the default value: rcp-docker-testing-virtual.artifactory-espoo2.int.net.nokia.com/default
<14>1 2025-07-12T19:37:22.515508Z [INFO] Received query result of CNF with status ongoing on CR: <rcp-log>default</rcp-log>/simulatorcnf
<14>1 2025-07-12T19:37:22.515556Z [INFO] Operation is ongoing for CR: <rcp-log>default</rcp-log>/simulatorcnf, operation: Install, is dryRun: false, some feedback may lost due to controller restart
<14>1 2025-07-12T19:37:23.235935Z [INFO] Got event feedback: CR=<rcp-log>default</rcp-log>/simulatorcnf, Event type=CnfLcmOperatorNotification, Event key=Performed a CNF query from the engine, TimeStamp=2025-07-12T19:37:22.515544931Z, Identifier=1752348972530665423

<14>1 2025-07-12T19:37:23.240052Z [INFO] Get latest CR <rcp-log>default</rcp-log>/simulatorcnf(resourceVersion: 7949) from controller-runtime with Phase: Preparing, CNFBuild: , Condition: [{Installed False 0 2025-07-12 19:36:34 +0000 UTC Instantiating Helm install simulatorcnf-prerequisite(Revision: 1, Chart: simulatorcnf-prerequisite-1.0.0, mainChart: false) successfully, operation id: 1752348972530665423} {Tested False 0 2025-07-12 19:36:12 +0000 UTC PrepareInstallation -} {Prepared True 0 2025-07-12 19:36:32 +0000 UTC Prepared Prepare CNF successfully} {Ready Unknown 0 2025-07-12 19:36:12 +0000 UTC PrepareInstallation -}]
<14>1 2025-07-12T19:37:23.249236Z [INFO] Update CR <rcp-log>default</rcp-log>/simulatorcnf(resourceVersion: 8046) status with Phase: Preparing, cnfSWBuild:  successfully
<14>1 2025-07-12T19:37:23.250386Z [INFO] Event(Update) from monitor informer for Install RANCNF <rcp-log>default</rcp-log>/simulatorcnf(ResourceVersion from api reader: 8046, ResourceVersion from informer: 8046, Operation ID from initStaus: 1752348972530665423, Operation ID from informer: 1752348972530665423)
<14>1 2025-07-12T19:37:23.250411Z [INFO] Check status for CR <rcp-log>default</rcp-log>/simulatorcnf: Ongoing
<14>1 2025-07-12T19:37:53.270482Z [INFO] Got event feedback: CR=<rcp-log>default</rcp-log>/simulatorcnf, Event type=Instantiated, Event key=Helm install simulatorcnf, TimeStamp=2025-07-12T19:37:52Z, Identifier=1752348972530665423

<14>1 2025-07-12T19:37:53.276839Z [INFO] Get latest CR <rcp-log>default</rcp-log>/simulatorcnf(resourceVersion: 8046) from controller-runtime with Phase: Preparing, CNFBuild: , Condition: [{Installed False 0 2025-07-12 19:36:34 +0000 UTC Instantiating Helm install simulatorcnf-prerequisite(Revision: 1, Chart: simulatorcnf-prerequisite-1.0.0, mainChart: false) successfully, operation id: 1752348972530665423} {Tested False 0 2025-07-12 19:36:12 +0000 UTC PrepareInstallation -} {Prepared True 0 2025-07-12 19:36:32 +0000 UTC Prepared Prepare CNF successfully} {Ready Unknown 0 2025-07-12 19:37:22 +0000 UTC CnfLcmOperatorNotification Got CNF status ongoing for operation Install}]
<14>1 2025-07-12T19:37:53.303759Z [INFO] Update CR <rcp-log>default</rcp-log>/simulatorcnf(resourceVersion: 8136) status with Phase: Preparing, cnfSWBuild:  successfully
<14>1 2025-07-12T19:37:53.303805Z [INFO] Got event feedback: CR=<rcp-log>default</rcp-log>/simulatorcnf, Event type=Testing, Event key=Test helm release started, TimeStamp=2025-07-12T19:37:52Z, Identifier=1752348972530665423

<14>1 2025-07-12T19:37:53.306713Z [INFO] Event(Update) from monitor informer for Install RANCNF <rcp-log>default</rcp-log>/simulatorcnf(ResourceVersion from api reader: 8136, ResourceVersion from informer: 8136, Operation ID from initStaus: 1752348972530665423, Operation ID from informer: 1752348972530665423)
<14>1 2025-07-12T19:37:53.306832Z [INFO] Check status for CR <rcp-log>default</rcp-log>/simulatorcnf: Ongoing
<14>1 2025-07-12T19:37:53.308707Z [INFO] Get latest CR <rcp-log>default</rcp-log>/simulatorcnf(resourceVersion: 8136) from controller-runtime with Phase: Preparing, CNFBuild: , Condition: [{Installed True 0 2025-07-12 19:37:52 +0000 UTC Instantiated Helm install simulatorcnf(Revision: 1, Chart: simulatorcnf-1.0.0-2-gdfa56bc6, mainChart: true) successfully, operation id: 1752348972530665423} {Tested False 0 2025-07-12 19:36:12 +0000 UTC PrepareInstallation -} {Prepared True 0 2025-07-12 19:36:32 +0000 UTC Prepared Prepare CNF successfully} {Ready Unknown 0 2025-07-12 19:37:22 +0000 UTC CnfLcmOperatorNotification Got CNF status ongoing for operation Install}]
<14>1 2025-07-12T19:37:53.330501Z [INFO] Update CR <rcp-log>default</rcp-log>/simulatorcnf(resourceVersion: 8137) status with Phase: Preparing, cnfSWBuild:  successfully
<14>1 2025-07-12T19:37:53.334057Z [INFO] Event(Update) from monitor informer for Install RANCNF <rcp-log>default</rcp-log>/simulatorcnf(ResourceVersion from api reader: 8137, ResourceVersion from informer: 8137, Operation ID from initStaus: 1752348972530665423, Operation ID from informer: 1752348972530665423)
<14>1 2025-07-12T19:37:53.334235Z [INFO] Check status for CR <rcp-log>default</rcp-log>/simulatorcnf: Ongoing
<14>1 2025-07-12T19:38:03.338634Z [INFO] Got event feedback: CR=<rcp-log>default</rcp-log>/simulatorcnf, Event type=Tested, Event key=Test helm release simulatorcnf successfully, TimeStamp=2025-07-12T19:38:02Z, Identifier=1752348972530665423

<14>1 2025-07-12T19:38:03.344830Z [INFO] Get latest CR <rcp-log>default</rcp-log>/simulatorcnf(resourceVersion: 8137) from controller-runtime with Phase: Preparing, CNFBuild: , Condition: [{Installed True 0 2025-07-12 19:37:52 +0000 UTC Instantiated Helm install simulatorcnf(Revision: 1, Chart: simulatorcnf-1.0.0-2-gdfa56bc6, mainChart: true) successfully, operation id: 1752348972530665423} {Tested False 0 2025-07-12 19:37:52 +0000 UTC Testing Test [simulatorcnf-test-connection] will be executed, operation id: 1752348972530665423} {Prepared True 0 2025-07-12 19:36:32 +0000 UTC Prepared Prepare CNF successfully} {Ready Unknown 0 2025-07-12 19:37:22 +0000 UTC CnfLcmOperatorNotification Got CNF status ongoing for operation Install}]
<14>1 2025-07-12T19:38:03.358572Z [INFO] Update CR <rcp-log>default</rcp-log>/simulatorcnf(resourceVersion: 8172) status with Phase: Preparing, cnfSWBuild:  successfully
<14>1 2025-07-12T19:38:03.358639Z [INFO] Got event feedback: CR=<rcp-log>default</rcp-log>/simulatorcnf, Event type=finished, Event key=Install helm charts of CNF instance successfully, TimeStamp=2025-07-12T19:38:02Z, Identifier=1752348972530665423

<14>1 2025-07-12T19:38:03.359068Z [INFO] Event(Update) from monitor informer for Install RANCNF <rcp-log>default</rcp-log>/simulatorcnf(ResourceVersion from api reader: 8172, ResourceVersion from informer: 8172, Operation ID from initStaus: 1752348972530665423, Operation ID from informer: 1752348972530665423)
<14>1 2025-07-12T19:38:03.359146Z [INFO] Check status for CR <rcp-log>default</rcp-log>/simulatorcnf: Ongoing
<14>1 2025-07-12T19:38:03.362934Z [INFO] Get latest CR <rcp-log>default</rcp-log>/simulatorcnf(resourceVersion: 8172) from controller-runtime with Phase: Preparing, CNFBuild: , Condition: [{Installed True 0 2025-07-12 19:37:52 +0000 UTC Instantiated Helm install simulatorcnf(Revision: 1, Chart: simulatorcnf-1.0.0-2-gdfa56bc6, mainChart: true) successfully, operation id: 1752348972530665423} {Tested True 0 2025-07-12 19:38:02 +0000 UTC Tested Test helm release simulatorcnf successfully, and the log of test pod is: [], operation id: 1752348972530665423} {Prepared True 0 2025-07-12 19:36:32 +0000 UTC Prepared Prepare CNF successfully} {Ready Unknown 0 2025-07-12 19:37:22 +0000 UTC CnfLcmOperatorNotification Got CNF status ongoing for operation Install}]
<14>1 2025-07-12T19:38:03.362974Z [INFO] Set status.Operation.Status to Succeed, eventType=finished
<14>1 2025-07-12T19:38:03.382718Z [INFO] Update CR <rcp-log>default</rcp-log>/simulatorcnf(resourceVersion: 8174) status with Phase: Running, cnfSWBuild: vCUCNF25R1_1.0.0-2-gdfa56bc6 successfully
<14>1 2025-07-12T19:38:03.382752Z [INFO] Operation for CR <rcp-log>default</rcp-log>/simulatorcnf has been finished
<14>1 2025-07-12T19:38:03.382821Z [INFO] Decrement cnf-config-manager-25r1-1-0-0-2-gdfa56bc6 SW build usage for namespace <rcp-log>default</rcp-log>
<14>1 2025-07-12T19:38:03.382843Z [INFO] Got status feedback: Namespace=<rcp-log>default</rcp-log>, CR Name=simulatorcnf, Status=finished

<14>1 2025-07-12T19:38:03.383837Z [INFO] Event(Update) from monitor informer for Install RANCNF <rcp-log>default</rcp-log>/simulatorcnf(ResourceVersion from api reader: 8174, ResourceVersion from informer: 8174, Operation ID from initStaus: 1752348972530665423, Operation ID from informer: 1752348972530665423)
<14>1 2025-07-12T19:38:03.383873Z [INFO] Check status for CR <rcp-log>default</rcp-log>/simulatorcnf: Succeed

//// monitor感知到install的operation已经成功了, 然后向stop channel中添加内容了,
///  之后会等待死循环在下一轮中被select出来后结束当前goroutine  ---这里是问题的关键
<14>1 2025-07-12T19:38:03.383929Z [INFO] Operation Install is not ongoing but Succeed, will stop this monitor informer

/// Delete事件触发的CR更新事件进来了, 此时CR的Resource Version还是8211
<14>1 2025-07-12T19:38:23.611390Z [INFO] CR <rcp-log>default</rcp-log>/simulatorcnf come in reconcile with resourceVersion: 8211
<14>1 2025-07-12T19:38:23.611428Z [INFO] CR <rcp-log>default</rcp-log>/simulatorcnf is in terminating
<14>1 2025-07-12T19:38:23.620284Z [INFO] Update status for CR <rcp-log>default</rcp-log>/simulatorcnf(new resourceVersion: 8212) successfully, current status phase is Preparing, current operation is {Uninstall Ongoing 1752349103611452449}, current build is vCUCNF25R1_1.0.0-2-gdfa56bc6, current conditions are [{Uninstalled False 0 2025-07-12 19:38:23 +0000 UTC PrepareUninstallation -} {Installed True 0 2025-07-12 19:37:52 +0000 UTC Instantiated Helm install simulatorcnf(Revision: 1, Chart: simulatorcnf-1.0.0-2-gdfa56bc6, mainChart: true) successfully, operation id: 1752348972530665423} {Prepared False 0 2025-07-12 19:38:23 +0000 UTC PrepareUninstallation -} {Ready Unknown 0 2025-07-12 19:38:23 +0000 UTC PrepareUninstallation -}]
<14>1 2025-07-12T19:38:23.620358Z [INFO] Uninstalling CNF for CR: <rcp-log>default</rcp-log>/simulatorcnf
<14>1 2025-07-12T19:38:23.620431Z [INFO] Waiting for Service:<rcp-log>cnf-lcm-operator</rcp-log>/cnf-config-manager-25r1-1-0-0-2-gdfa56bc6 for asset cnf-config-manager-25r1-1-0-0-2-gdfa56bc6 until ready
<14>1 2025-07-12T19:38:23.620470Z [INFO] Start action(Uninstall) timer with timeout 50m0s for CR <rcp-log>default</rcp-log>/simulatorcnf
<14>1 2025-07-12T19:38:23.622723Z [INFO] Get service:cnf-config-manager-25r1-1-0-0-2-gdfa56bc6, IP=10.104.6.55, port=[{httpserver TCP <nil> 8080 {0 8080 } 0}], selector=map[app:cnf-config-manager-25r1-1-0-0-2-gdfa56bc6], status=&ServiceStatus{LoadBalancer:LoadBalancerStatus{Ingress:[]LoadBalancerIngress{},},Conditions:[]Condition{},}
<14>1 2025-07-12T19:38:23.623001Z [INFO] cnf-config-manager-25r1-1-0-0-2-gdfa56bc6 for CNFBuild vCUCNF25R1_1.0.0-2-gdfa56bc6 is running. It can be reused
<14>1 2025-07-12T19:38:23.623044Z [INFO] Increment cnf-config-manager-25r1-1-0-0-2-gdfa56bc6 SW build usage for namespace <rcp-log>default</rcp-log>
<14>1 2025-07-12T19:38:23.623051Z [INFO] Request SW build vCUCNF25R1_1.0.0-2-gdfa56bc6 for CR <rcp-log>default</rcp-log>/simulatorcnf successfully
<14>1 2025-07-12T19:38:23.624504Z [INFO] Event(Add) from monitor informer for Uninstall RANCNF <rcp-log>default</rcp-log>/simulatorcnf(ResourceVersion from api reader: 8212, ResourceVersion from informer: 8212, Operation ID from initStaus: 1752349103611452449, Operation ID from informer: 1752349103611452449)
<14>1 2025-07-12T19:38:23.624527Z [INFO] Check status for CR <rcp-log>default</rcp-log>/simulatorcnf: Ongoing
<14>1 2025-07-12T19:38:23.624657Z [INFO] CNF uninstallation for CR <rcp-log>default</rcp-log>/simulatorcnf is ongoing, will wait for the event from lcm-engine
<14>1 2025-07-12T19:38:24.399008Z [INFO] Got event feedback: CR=<rcp-log>default</rcp-log>/simulatorcnf, Event type=Preparing, Event key=Preparing, TimeStamp=2025-07-12T19:38:23.620355299Z, Identifier=1752349103611452449

//// 至此没有打印"Operation stopped for CR, 表示handleRetryWithTimeout一直没有退出, 也因此无法执行到 'defer close(stop)'
//// 也因此, informer还是没有被停止, 也所以informer继续处理事件
//// 但是, 这里根据log可以看出来, 此时的CR其实是delete事件触发的那个CR, ResourceVersion
     wxy: 注意, 这里代码改了, 没有使用apiReader获取CR

//// 也就是说, 这里再次因为不是原子操作导致了两条指令之间发生了问题
<14>1 2025-07-12T19:38:24.399485Z [INFO] Event(Update) from monitor informer for Install RANCNF <rcp-log>default</rcp-log>/simulatorcnf(ResourceVersion from api reader: 8211, ResourceVersion from informer: 8211, Operation ID from initStaus: 1752348972530665423, Operation ID from informer: 1752348972530665423)
<14>1 2025-07-12T19:38:24.399516Z [INFO] Check status for CR <rcp-log>default</rcp-log>/simulatorcnf: Succeed
//// 执行这条命令的时候, 再次向stop中添加了内容, 能添加进去, 说明上一个已经被select出来, 只是还没来得及打印
<14>1 2025-07-12T19:38:24.399548Z [INFO] Operation Install is not ongoing but Succeed, will stop this monitor informer
panic: context: internal error: missing cancel error

goroutine 143 [running]:
context.(*cancelCtx).cancel(0x23f29c8?, 0x90?, {0x0?, 0x0?}, {0x0?, 0x0?})
	/usr/local/go/src/context/context.go:542 +0x2c5
context.(*cancelCtx).propagateCancel.func2()
	/usr/local/go/src/context/context.go:516 +0xf1
created by context.(*cancelCtx).propagateCancel in goroutine 181
	/usr/local/go/src/context/context.go:513 +0x3e6

```
