

// 知识点1：阻塞select(不会浪费CPU)和 非阻塞select(可能会浪费CPU)
// 1) 阻塞式select
/*
	for{
		select {
		case <-parentCtx.Done():
			...
		case feedbackInfo, ok := <-ch:
			...
		default:
			time.Sleep(10 * time.Second)
		}
	}
*
// 2) 非阻塞式select
/*
	for{
		select {
		case <-parentCtx.Done():
			...
		case <-ctx.Done():
			...
		case feedbackInfo, ok := <-ch:
			...
		}
	}
*/
// 小小结:
// 首先，select是否阻塞是select的语法，之所以外套一层for循环, 是为了说明问题
// 在阻塞式select中, 如果在两个channel中都没有select出来则会持续保持你休眠，并不会占用CPU
// 在非阻塞式select中，因为有default的存在，所以并不会阻塞，那么就会发生循环继续从而造成大量消耗CPU的问题
//        也因此，在非阻塞式select中，需要在default中主动放弃cpu
//
// 结合实际情况，完全不需要这个等待，直接使用阻塞式就可以了, 原因是会导致event不会被及时处理,
// 尤其是第一个event, 除非在本函数调用之前，否则第一个event会在sleep后才会被处理


func (r *RANCNFReconciler) HandleFeedbacks(parentCtx context.Context, ch <-chan interface{}) {
	ctx, cancel := context.WithCancel(parentCtx)
	go func (ctx context.Context)  {
		for {
			select {
			// 知识点3: 如果parentCtx被Cancel掉，这里会立马感知，甚至可能还早于主线程中的<-parentCtx.Done()
			//         如果想要主动让这里得到感知，则需要主动将当前ctx的cancel掉，即执行主goroutine的defer cancel()
			case <-ctx.Done():
				log.Infof("I will exit this goroutine due to ctx done.")
				return
			}
		}
	}(ctx)

	defer cancel()
	for {
		log.Infof("One loop of select start")
		select {

		// 知识点2：context和propagated context之间的关系
		//         关于从实验结果来看，parentCtx.Done()会先于ctx.Done()触发
		//         且可以无限次的从他们中select出来(数据?)
		//         Golang: ctx.Done()的触发是因为parentCtx的cance)
		//
		//         在此后的循环中, 可以不断的从parentCtx.Done()/ctx.Done()中获取到数据
		//         打印的log如下：
				// 2025-04-02T02:48:23.417695+00:00 - - One loop of select start
				// 2025-04-02T02:48:28.421786+00:00 - - Stopping HandleFeedbacks due to parent context cancellation.
				// 2025-04-02T02:48:28.421844+00:00 - - One loop of select finished
				// 2025-04-02T02:48:28.421869+00:00 - - One loop of select start
				// 2025-04-02T02:48:28.421877+00:00 - - Stopping HandleFeedbacks due to parent context cancellation.
				// 2025-04-02T02:48:28.421883+00:00 - - One loop of select finished
				// 2025-04-02T02:48:28.421888+00:00 - - One loop of select start
				// 2025-04-02T02:48:28.421895+00:00 - - Stopping HandleFeedbacks due to parent context cancellation.
				// 2025-04-02T02:48:28.421913+00:00 - - One loop of select finished
				// 2025-04-02T02:48:28.421918+00:00 - - One loop of select start
				// 2025-04-02T02:48:28.421924+00:00 - - Stopping HandleFeedbacks due to derived context cancellation.
				// 2025-04-02T02:48:28.421929+00:00 - - One loop of select finished
				// ...
				// 2025-04-02T02:48:28.422893+00:00 - - Stopping HandleFeedbacks due to derived context cancellation.
				// 2025-04-02T02:48:28.422898+00:00 - - One loop of select finished
				// 2025-04-02T02:48:28.422903+00:00 - - One loop of select start
				// 2025-04-02T02:48:28.422909+00:00 - - Stopping HandleFeedbacks due to parent context cancellation.
				// 2025-04-02T02:48:28.422933+00:00 - - One loop of select finished
				// 2025-04-02T02:48:28.42294+00:00  - - One loop of select start
				// 2025-04-02T02:48:28.422946+00:00 - - Stopping HandleFeedbacks due to parent context cancellation.
				// 2025-04-02T02:48:28.422962+00:00 - - One loop of select finished
				// 2025-04-02T02:48:28.422983+00:00 - - One loop of select start
				// 2025-04-02T02:48:28.422991+00:00 - - Failed to get feedback! Maybe the channel generating feedback has been closed.
		case <-parentCtx.Done():
			log.Info("Stopping HandleFeedbacks due to parent context cancellation.")
			time.Sleep(1 * time.Second)
			return
		// case <-ctx.Done():
		// 	log.Info("Stopping HandleFeedbacks due to derived context cancellation.")
		// 	time.Sleep(10 * time.Second)
			// return
		case feedbackInfo, ok := <-ch:
			if !ok {
				log.Errorln("Failed to get feedback! Maybe the channel generating feedback has been closed.")
				time.Sleep(3 * time.Second)
				return
			}
		default: // 决定了是否是阻塞式select
		// log.Info("Start Wait for 10s and return")
		// time.Sleep(10 * time.Second)
		// log.Info("End 10s wait and return")
		}
		log.Infof("One loop of select finished")
	}
}
