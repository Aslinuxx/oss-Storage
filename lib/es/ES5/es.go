//该包是对ES包的再封装

package ES5

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//在apiservers中我们使用到了
// put请求: es 中增加新版本记录
// get请求: es中根据 object_name, version 获取整个元数据
// locate请求: 同get请求
// version请求: 搜索指定object的所有版本; 以及没有指定object对象时候默认搜索全部对象的所有版本

//这就是元数据; 相当于objects表
type Metadata struct {
	Name    string
	Version int
	Size    int64
	Hash    string
}

type hit struct {
	Source Metadata `json:"_source"`
}

// 这里是对搜索结果进行格式处理,使之与getMetadata返回的 Metadata结构体数据一致
//因为调用不同的访问url访问ES,返回的结果不同
type searchResult struct {
	Hits struct {
		Total int
		Hits  []hit
	}
}

// 根据object_name, 指定的版本versionId 获取对象的元数据
func getMetadata(name string, versionId int) (meta Metadata, e error) {

	//制作向ES服务器访问的http请求
	url := fmt.Sprintf("http://%s/metadata/objects/%s_%d/_source",
		os.Getenv("ES_SERVER"), name, versionId)
	//访问ES服务器,并或者搜索结果
	r, e := http.Get(url)
	if e != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf("fail to get %s_%d: %d", name, versionId, r.StatusCode)
		return
	}

	result, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(result, &meta) //反序列化; 将json格式解码转成Metadata结构体结构
	return

}

//根据对象名称搜索该对象的最新元数据版本
func SearchLatestVersion(name string) (meta Metadata, e error) {

	url := fmt.Sprintf("http://%s/metadata/_search?q=name:%s&size=1&sort=version:desc",
		os.Getenv("ES_SERVER"), url.PathEscape(name))

	r, e := http.Get(url)
	if e != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf("fail to search latest metadata: %d", r.StatusCode)
		return
	}

	result, _ := ioutil.ReadAll(r.Body)
	var sr searchResult
	json.Unmarshal(result, &sr)
	if len(sr.Hits.Hits) != 0 {
		meta = sr.Hits.Hits[0].Source
	}
	return

}

// 根据object_name, 指定的版本versionId 获取对象的元数据
// 如果客户端没有指定version版本;默认搜索最新的版本
func GetMetadata(name string, version int) (Metadata, error) {
	if version == 0 {
		return SearchLatestVersion(name)
	}

	return getMetadata(name, version)

}

//put完对象后,对应的需要在ES中更新其版本;
// 先获取当前object_name的最新版本; 再插入一条新版本数据字段
func AddVersion(name string, size int64, hash string) error {
	version, e := SearchLatestVersion(name) // 这里的version是Metadata结构体
	if e != nil {
		return e
	}
	return PutMetadata(name, version.Version+1, size, hash)
}

//插入一条指定对象的最新版本
func PutMetadata(name string, version int, size int64, hash string) error {
	doc := fmt.Sprintf(`{"name":"%s","version":%d,"size":%d,"hash":"%s"}`,
		name, version, size, hash)

	client := http.Client{}
	url := fmt.Sprintf("http://%s/metadata/objects/%s_%d?op_type=create",
		os.Getenv("ES_SERVER"), name, version)
	//这里是put新版本到ES
	request, _ := http.NewRequest("PUT", url, strings.NewReader(doc))
	r, e := client.Do(request)

	if e != nil {
		return e
	}
	if r.StatusCode == http.StatusConflict {
		return PutMetadata(name, version+1, size, hash) //版本已存在则加1再试
	}
	if r.StatusCode != http.StatusCreated {
		result, _ := ioutil.ReadAll(r.Body)
		return fmt.Errorf("fail to put metadata: %d %s", r.StatusCode, string(result))
	}
	return nil

}

//搜索指定对象的所有版本
func SearchAllVersions(name string, form, size int) ([]Metadata, error) {
	url := fmt.Sprintf("http://%s/metadata/_search?sort=name,version&from=%d&size=%d",
		os.Getenv("ES_SERVER"), form, size)
	if name != "" { // 如果没指定object_name 默认搜索所有的对象的所有版本
		url += "&q=name:" + name
	}
	r, e := http.Get(url)
	if e != nil {
		return nil, e
	}
	metas := make([]Metadata, 0)
	result, _ := ioutil.ReadAll(r.Body)
	var sr searchResult
	json.Unmarshal(result, &sr)
	for i := range sr.Hits.Hits {
		metas = append(metas, sr.Hits.Hits[i].Source)
	}
	return metas, nil
}

func DelMetadata(name string, version int) {
	client := http.Client{}
	url := fmt.Sprintf("http://%s/metadata/objects/%s_%d",
		os.Getenv("ES_SERVER"), name, version)
	request, _ := http.NewRequest("DELETE", url, nil)
	client.Do(request)
}

type Bucket struct {
	Key         string
	Doc_count   int
	Min_version struct {
		Value float32
	}
}

type aggregateResult struct {
	Aggregations struct {
		Group_by_name struct {
			Buckets []Bucket
		}
	}
}

func SearchVersionStatus(min_doc_count int) ([]Bucket, error) {
	client := http.Client{}
	url := fmt.Sprintf("http://%s/metadata/_search", os.Getenv("ES_SERVER"))
	body := fmt.Sprintf(`
        {
          "size": 0,
          "aggs": {
            "group_by_name": {
              "terms": {
                "field": "name",
                "min_doc_count": %d
              },
              "aggs": {
                "min_version": {
                  "min": {
                    "field": "version"
                  }
                }
              }
            }
          }
        }`, min_doc_count)
	request, _ := http.NewRequest("GET", url, strings.NewReader(body))
	r, e := client.Do(request)
	if e != nil {
		return nil, e
	}
	b, _ := ioutil.ReadAll(r.Body)
	var ar aggregateResult
	json.Unmarshal(b, &ar)
	return ar.Aggregations.Group_by_name.Buckets, nil
}

func HasHash(hash string) (bool, error) {
	url := fmt.Sprintf("http://%s/metadata/_search?q=hash:%s&size=0", os.Getenv("ES_SERVER"), hash)
	r, e := http.Get(url)
	if e != nil {
		return false, e
	}
	b, _ := ioutil.ReadAll(r.Body)
	var sr searchResult
	json.Unmarshal(b, &sr)
	return sr.Hits.Total != 0, nil
}

func SearchHashSize(hash string) (size int64, e error) {
	url := fmt.Sprintf("http://%s/metadata/_search?q=hash:%s&size=1",
		os.Getenv("ES_SERVER"), hash)
	r, e := http.Get(url)
	if e != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf("fail to search hash size: %d", r.StatusCode)
		return
	}
	result, _ := ioutil.ReadAll(r.Body)
	var sr searchResult
	json.Unmarshal(result, &sr)
	if len(sr.Hits.Hits) != 0 {
		size = sr.Hits.Hits[0].Source.Size
	}
	return
}
