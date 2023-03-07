package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"zim.cn/biz/proto"

	"zim.cn/base"
)

var pool *sync.Pool

type GobEncoder struct {
	enc *gob.Encoder
	buf *bytes.Buffer
}

func (e *GobEncoder) Reset() {
	e.buf.Reset()
}

func (e *GobEncoder) Encode(a any) ([]byte, error) {
	err := e.enc.Encode(a)
	if err != nil {
		return nil, err
	}
	b := e.buf.Bytes()
	return b, nil
}

func init() {
	pool = &sync.Pool{
		New: func() any {
			var buf = new(bytes.Buffer)
			e := &GobEncoder{
				enc: gob.NewEncoder(buf),
				buf: buf,
			}
			return e
		},
	}
}

func encodeGob(a any) []byte {
	//var buf bytes.Buffer
	enc := pool.Get().(*GobEncoder)
	defer pool.Put(enc)
	enc.Reset()
	b, err := enc.Encode(a)
	base.Raise(err)
	return b
}

func encodeJson(a any) []byte {
	b, _ := json.Marshal(a)
	return b
}

func decodeGob(data []byte, p any) error {
	var buf = bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(p)
	return err
}

func decodeJson(data []byte, p any) error {
	err := json.Unmarshal(data, p)
	return err
}

// gob的精髓在于流式编码, 单个编码是非常慢的. 也就是说只适用于socket连接的编码, 因为首次编码会额外携带类型信息.
func benchmark(typ int, n int) {
	a := &proto.Message{}
	data := []byte(`{"id":32,"type":101,"from_id":"u32380990889","from_user":{"id":"u32380990889","name":"test3","photo":"xx","ex":"ee","deleted":false},"to_id":"g9295541673","elem":{"text":"额外"},"revoked":false,"created_at":1659693057}`)
	json.Unmarshal(data, &a)
	gdata := encodeGob(a)
	var buf = bytes.NewBuffer(nil)
	dec := gob.NewDecoder(buf)
	//enc := gob.NewEncoder(buf)
	st := time.Now()
	for i := 0; i < n; i++ {
		if typ == 1 {
			b := encodeGob(a)
			_ = b
			//fmt.Println("b:", len(b))
		} else if typ == 2 {
			b := encodeJson(a)
			_ = b
			//fmt.Println("B:", len(b))
		} else if typ == 3 {
			//err := decodeGob(gdata, &a)
			//enc.Encode(a)
			buf.Write(gdata)
			fmt.Println("buflen:", len(gdata), buf.Len())
			//buf.Write(gdata)
			var q *proto.Message
			err := dec.Decode(&q)
			base.Raise(err)
			fmt.Println("q:", base.JsonString(q))

			//fmt.Println("C:", base.JsonString(a))
		} else if typ == 4 {
			err := decodeJson(data, &a)
			base.Raise(err)
		}
	}
	fmt.Println("gob took:", time.Since(st))
}

func main() {
	benchmark(3, 10)
}
