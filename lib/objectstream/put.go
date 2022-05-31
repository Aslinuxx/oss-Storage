package objectstream


import (
	"fmt"
	"io"
	"net/http"
)


//实现apiServer与指定的dataServer进行通信; 并实现将用户put的object对象的具体数据内容put到dataserver上;
// 即实现apiserver与dataServer的put请求; 即将objects的具体内容存入 apiserver与dataserver进行http.requeset的正文中;
// 从而构造出在dataserver来看就是一个全新的客户端的put请求

type PutStream struct{
	writer *io.PipeWriter
	c chan error

}

//传入参数是数据存储服务器的ip地址和object对象的名称
func NewPutStream(server, object_name string) *PutStream {
	reader, writer := io.Pipe() // 生成一对 reader和writer;
	// 这里writer和reader是是管道互联的; 相当于如果往writer中写入什么值,也能从reader中读出; 可能内部是共享内存?
	// 所以这个在之前store.go中我们使用io.Copy(PutStream,r.Body)时候将object对象的具体数据内容写入到writer中后;
	// 则reader也就是该数据内容


	c := make(chan error) // channel初始化

	// 下面是将reader的内容 put到dataserver;
	// 此时apiserver就是一个http的客户端,去put请求dataserver;这与单机版的一致
	go func(){
		// 生成一个新的request请求; Put请求, URL是"http://"+server+"/objects/"+object_name, 正文是reader(即object对象的具体数据内容)
		request, _ := http.NewRequest("PUT", "http://"+server+"/objects/"+object_name, reader)
		client := http.Client{} //新建一个http客户端; go中不进行初始化也可以使用;默认使用内部值
		r,e := client.Do(request)
		if e == nil && r.StatusCode != http.StatusOK{
			e = fmt.Errorf("dataServer return http code %d",r.StatusCode)
		}
		c <- e
	}()

	return &PutStream{writer,c}

}


func (w *PutStream) Write(p []byte) (n int, err error){

	return w.writer.Write(p)
}

func (w *PutStream) Close() error {
	w.writer.Close()

	return <-w.c //将error管道中的数据读出
}



























