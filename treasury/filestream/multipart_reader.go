package filestream

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

var serverRoot = "/root/go/src/go/testdata"
var files = []string{
	"hello-world-0.0.1.tgz",
	"hello-world-0.2.0.tgz",
	"hello-world-calculate-0.2.1.tgz",
	"config-manager", // 190M
}

func setDebugServer(r *mux.Router) {
	r.HandleFunc("/debug", downloadHandler).Methods("GET")
}

// 如何利用goland的multipart 包，将多个文件封装到http请求的body中，从而提供给client下载使用
// [核心思想]：所谓multi part，就是存在多个parts，那么就可以将每一个file作为一个part，填充到http应答的body中，
//
// [关键操作]：
//	1. https header：首先，要让http的接收方知道这是一个multipart类型，所以要设置http的头
//	                 然后，要让接收方知道每一个part的边界，所以要设置boundary
//	2. http body: 包含了所有parts的内容，每一个part由boundary分隔，格式如下：
//			1)boundary: 为"--%s\r\n"格式
//			2)part的header, 包含两个部分
//			   Content-Disposition: 用来指导接收方如何处理这个part的, 包含两个字段, name和filename,
//                                  接收方可以自行决定使用哪个字段的值来标识当前这个part的数据

//				   Content-Type: 表示这个part是什么类型的数据,
//								如果是来自file的, 则直接设置成application/octet-stream
//	        3) raw data
//	        4) 所有part的结束后, 还会有一个全部结束的boundary, 格式为: "\r\n--%s\r\n"
//
// [编码核心步骤]：
//  1. 创建一个multipart.Writer实例，将http.ResponseWriter作为其参数
//  2. 设置http响应头，Content-Type: multipart/form-data; boundary=xxx
//  3. 为每一个file穿件一个file类型的part
//     1) 初始化一个part，包含的header信息
//     2) 从file中读取内容，写入到part中

func downloadHandler(w http.ResponseWriter, r *http.Request) {

	// 1. 初始化一个多部分写入器，写入的目的地是http的response
	mw := multipart.NewWriter(w)
	defer mw.Close()

	// 2. 设置响应头: 内容的类型，以及多个part的分隔符
	w.Header().Set("Content-Type", mw.FormDataContentType())

	// 3. 为每一个file创建一个part
	for _, fileName := range files {
		// 1) 创建一个file类型的part，即这个part中的内容是一个文件的内容
		//    每一个part都自己的header(不要与http那个header混淆)，包含了一些标识信息
		//    content-type： 对于file类型，其值就是application/octet-stream
		//    content-disposition: 对于file类型，其值是form-data; name="file"; filename="fileName"
		//                             包含两个参数，一个是name，一个是filename
		// wxy: 注意这个函数会自动将新建的part append到multipart.Writer中，即最后的boundary也会自动的后移
		part, err := mw.CreateFormFile("file", filepath.Base(fileName))
		if err != nil {
			return
		}

		// 2). 向part中填充真正的数据
		//    在这里是从指定文件中读取内容copy进去
		if err := writePart(part, fileName); err != nil {
			return
		}
	}

	// 4. 除此之外， 还可以添加一些其他的field，这些field是不是file类型的
	params := make(map[string]string)
	params["wxy comments"] = "allinone"
	for key, val := range params {
		_ = mw.WriteField(key, val)
	}
}

func writePart(part io.Writer, fileName string) error {
	filePath := filepath.Join(serverRoot, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	return nil
}
