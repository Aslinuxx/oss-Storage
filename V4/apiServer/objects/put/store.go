package put

import (
	"Storage/V4/apiServer/locate"
	"Storage/lib/utils"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

//object对象的具体内容, hash值, size

func storeObject(r io.Reader, hash string, size int64) (int, error) {
	//put请求加上去重和校验后,第一步就是查看是否重复; 重复的话只需进行逻辑更新(增加版本号即可)
	if locate.Exist(url.PathEscape(hash)) {
		return http.StatusOK, nil
	}

	//建立临时数据流
	stream, e := putStream(url.PathEscape(hash), size)
	if e != nil {
		return http.StatusInternalServerError, e
	}

	//这里掉了TempPutStream.Write方法;
	//需要研究下数据流的写入
	reader := io.TeeReader(r, stream) // 这里替代了io.Copy()
	//这个就类似linux中的tee 命令 将 c = tee(A,B) 将A的内容输出到C,同时写入文件B;

	//计算写完的数据的散列值是否一致
	d := utils.CalculateHash(reader)
	if d != hash {
		stream.Commit(false) //不一致就将临时文件删除;
		return http.StatusBadRequest, fmt.Errorf("object hash mismatch, calculated=%s, requested=%s", d, hash)

	}
	stream.Commit(true) //一致就将临时文件转正;
	return http.StatusOK, nil

}
