1. Content-Disposition
解释：用于指定如何处理响应内容。它通常用于指示浏览器是否应该显示响应内容内联, 还是作为附件下载。
      wxy: 就是说一打开这个链接, 对应的内容是直接渲染在浏览器页面上, 还是作为一个文件被下载都本地
取值：
1). inline
   表示内容应该内联显示在浏览器窗口中。
   示例：Content-Disposition: inline:


2). attachment:
   表示内容应该作为附件下载，并且可以指定下载时的文件名。
   示例：Content-Disposition: attachment; filename="example.txt"
   实操：
   (1)在使用curl命令的时候, 就会直接保存到本地, 其实这个filename没有什么用
	  # curl http://localhost:8383/debug2 --output ./debug2
	  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
									 Dload  Upload   Total   Spent    Left  Speed
	100  9771    0  9771    0     0    832      0 --:--:--  0:00:11 --:--:--  2160
	
	root@lcmaas-wksp-g:~/tmp# ll
	-rw-r--r--  1 root root 9771 Sep 26 14:19 debug2

   (2)使用代码接收

1. Content-Type:
解释: 用于指示响应内容的媒体类型（MIME 类型）。它告诉客户端如何解释响应内容。以下是一些常见的 `Content-Type` 值：

1. 文本类型
   - `text/plain`: 纯文本
   - `text/html`: HTML 文档
   - `text/css`: CSS 样式表
   - `text/javascript`: JavaScript 脚本

2. 应用程序类型
   - `application/json`: JSON 数据
   - `application/xml`: XML 数据
   - `application/octet-stream`: 任意二进制数据
   - `application/pdf`: PDF 文档
   - `application/zip`: ZIP 压缩文件
   - `application/x-www-form-urlencoded`: 表单数据（URL 编码）


3. 图像类型
     Content-Type: image/jpeg
     Content-Type: image/png
     ...

4. 音频和视频类型
   - audio/mpeg: MPEG 音频
   - audio/ogg: OGG 音频
   - video/mp4: MP4 视频
   - video/ogg: OGG 视频

5. 多部分类型

   - multipart/form-data: 表单数据（包括文件上传）
   - multipart/byteranges: 多部分字节范围
   
   要点：
   1. 当使用'multipart'类型的'Content-Type'时，'boundary' 参数用于分隔多个部分内容。
      每个部分内容之间使用 `boundary` 字符串进行分隔，以便客户端能够正确解析和处理每个部分。

===========================================
使用multipart/form-data标识的http response, 其body中的内容如下：

--d73e603c0c6039d7653c6fe823b8b92895c26539aeb21aa925bec6f5c6de   ---第1个
Content-Disposition: form-data; name="field1"; filename="hello-world-0.0.1.tgz"
Content-Type: application/octet-stream

�\I
--d73e603c0c6039d7653c6fe823b8b92895c26539aeb21aa925bec6f5c6de   ---第2个
Content-Disposition: form-data; name="field1"; filename="hello-world-0.2.0.tgz"
Content-Type: application/octet-stream

�
--d73e603c0c6039d7653c6fe823b8b92895c26539aeb21aa925bec6f5c6de  ---第3个
Content-Disposition: form-data; name="field1"; filename="hello-world-calculate-0.2.1.tgz"
Content-Type: application/octet-stream

�:
--6e948143c10259bccbfce8ff25902542d092428c910311c66541fac4b02c  ---第4个
Content-Disposition: form-data; name="file"; filename="busybox-v1.tar"
Content-Type: application/octet-stream

274279bf...":["274279bf1f8f5bff62b606e780ce6b68c4355b1853fc31410f0e6f2516eef3bf/layer.tar"]}]




--6e948143c10259bccbfce8ff25902542d092428c910311c66541fac4b02c  ---第5个, 是普通表单形式, raw data则变成了字符串
Content-Disposition: form-data; name="wxy comments"

allinone
--6e948143c10259bccbfce8ff25902542d092428c910311c66541fac4b02c--    ---全部的结束符


解析：一个分成3个part, 每一个part由如下三部分构成
1)boundary: 为"--%s\r\n"格式
2)part的header, 包含两个部分
   Content-Disposition: 用来指导接收方如何处理这个part的, 包含两个字段, name和filename, 
                        接收方可以自行决定使用哪个字段的值来标识当前这个part的数据

   Content-Type: 表示这个part是什么类型的数据, 如果是来自file的, 则直接设置成application/octet-stream
3) raw data
4) 所有part的结束后, 还会有一个全部结束的boundary, 格式为: "\r\n--%s\r\n"
