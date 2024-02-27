package algorithms

import (
	"fmt"
)

/* 冒泡排序 */
// 排序过程(升序)：
// 原始数据：
// 3, 2, 9, 5, 7, 1, 6

// 2, 3, 5, 7, 1, 6, 9
// 2, 3, 5, 1, 6, 7, 9
// 2, 3, 1, 5, 6, 7, 9
// 2, 1, 3, 5, 6, 7, 9
// 1, 2, 3, 5, 6, 7, 9
// 1, 2, 3, 5, 6, 7, 9

// 小小结：
// 每一轮(内)循环都是相邻元素之间的pk与置换：最大(小)置底
// 每一轮(内)循环因为已经将最大(小)的元素放在底下了，所以循环次数逐渐减少
// 结：冒泡沉底从头来
func bubbleSort(elems []int) {
	length := len(elems)
	for i := 0; i < length-1; i++ { // 外循环为排序趟数，len个数进行len-1趟
		for j := 0; j < length-1-i; j++ { // 内循环为每趟比较的次数，第i趟比较len-i次
			if elems[j] > elems[j+1] { // 被后置的元素将在下一轮和再后面的元素做比较
				temp := elems[j]
				elems[j] = elems[j+1]
				elems[j+1] = temp
			}
		}
		// display
		print(elems)
	}
}

/* 选择排序 */
// 原始数据：
// 3, 2, 9, 5, 7, 1, 6

// 1, 2, 9, 5, 7, 3, 6
// 1, 2, 9, 5, 7, 3, 6
// 1, 2, 3, 5, 7, 9, 6
// 1, 2, 3, 5, 7, 9, 6
// 1, 2, 3, 5, 6, 9, 7
// 1, 2, 3, 5, 6, 7, 9

// 小小结：
// 每一轮(内)循环找到其中最小(大)那个放置在前面应该的位置上
// 每一轮(内)循环为最小(大)找到位置后，可用位置就要后移一个了
// 结：选择置位接着来
func selectionSort(elems []int) {
	length := len(elems)
	for i := 0; i < length-1; i++ { /*外循环: 1)为选择参照物*/
		min := i
		for j := i + 1; j < length; j++ { /*内循环: 没一轮内层循环,选出最小的那个元素*/
			if elems[j] < elems[min] {
				min = j
			}
		}
		temp := elems[min] /*外循环: 2)将内循环选举出来的最小元素 和 外循环当前的参数元素做交换*/
		elems[min] = elems[i]
		elems[i] = temp
		// display
		print(elems)
	}
}

func print(data []int) {
	splitter := ""
	for _, d := range data {
		fmt.Printf("%s%d", splitter, d)
		splitter = ", "
	}
	fmt.Println()
}

func Sort() {
	data := []int{3, 2, 9, 5, 7, 1, 6}
	//bubbleSort(data)
	selectionSort(data)
}
