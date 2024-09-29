package multipart

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestDownloadAndSave(t *testing.T) {
	if err := parseAndSaveFiles(); err != nil {
		panic(err)
	}
}

// 如何利用goland的multipart 包，从http应答的body中读取得到多个文件并保存到本地，从而提供给client下载使用
// [核心思想]：所谓multi part，就是存在多个parts，我们和server端约定好，每一个part都将是一个file
//
// [关键操作]：
//	1. https header：首先，要知道这是一个multipart类型
//	                 然后，要知道每一个part的边界
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
//  1. 解析http响应头，Content-Type: multipart/form-data; boundary=xxx
//  2. 创建一个multipart.Reader实例，将response的body和boundary传递给multipart.Reader
//  3. 读取每一个part，解析保存成为一个个本地file
func parseAndSaveFiles() error {
	url := "http://localhost:8383/debug"
	// 发送 GET 请求到远程服务器
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	// debug
	// body, _ := io.ReadAll(resp.Body)
	// fmt.Printf("Response body: %d bytes: %s ", len(body), string(body))

	// 1. 读取http的响应头，获取Content-Type，得两部分信息
	//    1) 这是一个multipart/form-data类型的响应
	//    2）那么必定需要boundary， 解析出boundary
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return fmt.Errorf("invalid content type: %s", contentType)
	}

	// 解析 Content-Type 中的 boundary
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("failed to parse media type: %v", err)
	}
	boundary := params["boundary"]
	if boundary == "" {
		return fmt.Errorf("no boundary found in Content-Type")
	}

	// 2. 使用multipart包提供的NewReader封装response中body，准备读取并解析
	mr := multipart.NewReader(resp.Body, boundary)

	// 3. 根据boundary，分割出来各个part，然后处理每个part
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			// 所有部分读取完毕，退出循环
			fmt.Println("All parts read.")
			break
		}
		if err != nil {
			return fmt.Errorf("error reading next part: %v", err)
		}

		err = readPart(part, "")
		if err != nil {
			return fmt.Errorf("failed to save file: %v", err)
		}
	}
	return nil
}

// saveFile 保存文件到本地路径
func readPart(part *multipart.Part, rename string) error {
	fmt.Println("Processing part:", part.Header)
	defer part.Close()

	fmt.Printf("Processing Part: %s, FileName: %s\n", part.FormName(), part.FileName())
	if part.FileName() == "" {
		// 附加: 对于不是file类型的part，则为普通的表单字段， 直接读取
		data, err := io.ReadAll(part)
		if err != nil {
			return fmt.Errorf("failed to read form field: %v", err)
		}
		fmt.Printf("Form field %s: %s\n", part.FormName(), string(data))
		return nil
	}

	if rename == "" {
		rename = part.FileName()
	}
	out, err := os.Create("./" + rename)
	if err != nil {
		return err
	}
	defer out.Close()

	// 将文件内容写入到本地文件
	_, err = io.Copy(out, part)
	if err != nil {
		return err
	}

	return nil
}
