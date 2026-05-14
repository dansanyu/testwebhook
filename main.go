package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
	"wuliu/proto/pb"

	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type ExpressResponse struct {
	Msg       string `json:"msg"`
	Status    string `json:"status"`
	ErrorCode string `json:"error_code"`
	Data      struct {
		Context []struct {
			Time string `json:"time"`
			Desc string `json:"desc"`
		} `json:"context"`
		Status          string `json:"status"`
		State           string `json:"state"`
		OfficialService struct {
			Com              string `json:"com"`
			ComName          string `json:"comName"`
			URLDesc          string `json:"urlDesc"`
			URL              string `json:"url"`
			Logo             string `json:"logo"`
			ServicePhoneDesc string `json:"servicePhoneDesc"`
			ServicePhone     string `json:"servicePhone"`
			Service          []struct {
				URL  string `json:"url"`
				Name string `json:"name"`
			} `json:"service"`
		} `json:"officalService"`
	} `json:"data"`
}

// 物流条目
type ExpressItem struct {
	AcceptTime    string `json:"AcceptTime"`
	AcceptStation string `json:"AcceptStation"`
}

type ExpressResult struct {
	Code      int           `json:"code"`
	WuliuCode string        `json:"wuliu_code"`
	Com       string        `json:"com"`
	Msg       string        `json:"msg"`
	Data      []ExpressItem `json:"data"`
}

func GetZTByBaidu(code, company, tokenV2, cookie string) ExpressResult {
	if company == "" {
		company = "zhongtong"
	}

	baseURL := "https://alayn.baidu.com/express/appdetail/get_detail"
	reqURL := fmt.Sprintf("%s?qid=adaef5e400054375&query_from_srcid=51151&tokenV2=%s&appid=4001&nu=%s&com=%s",
		baseURL,
		tokenV2,
		code,
		company,
	)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	fmt.Println(reqURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return ExpressResult{
			Code:      500,
			WuliuCode: code,
			Com:       company,
			Msg:       "创建请求失败: " + err.Error(),
			Data:      []ExpressItem{},
		}
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 Edge/17.17134")
	req.Header.Set("Host", "alayn.baidu.com")
	req.Header.Set("Cookie", cookie)

	resp, err := client.Do(req)
	if err != nil {
		return ExpressResult{
			Code:      500,
			WuliuCode: code,
			Com:       company,
			Msg:       "请求百度接口失败: " + err.Error(),
			Data:      []ExpressItem{},
		}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ExpressResult{
			Code:      500,
			WuliuCode: code,
			Com:       company,
			Msg:       "读取响应失败: " + err.Error(),
			Data:      []ExpressItem{},
		}
	}

	var data ExpressResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return ExpressResult{
			Code:      500,
			WuliuCode: code,
			Com:       company,
			Msg:       "解析 JSON 失败: " + err.Error(),
			Data:      []ExpressItem{},
		}
	}

	if data.Status == "0" && len(data.Data.Context) > 0 {
		list := make([]ExpressItem, len(data.Data.Context))
		for i, val := range data.Data.Context {
			ts, err := strconv.ParseInt(val.Time, 10, 64)
			if err != nil {
				// 解析失败，给一个默认值或跳过
				ts = 0
			}
			list[i] = ExpressItem{
				AcceptStation: val.Desc,
				AcceptTime:    time.Unix(ts, 0).Format("2006-01-02 15:04:05"),
			}
		}

		return ExpressResult{
			Code:      200,
			WuliuCode: code,
			Com:       company,
			Msg:       data.Msg,
			Data:      list,
		}
	}

	return ExpressResult{
		Code:      300,
		WuliuCode: code,
		Com:       company,
		Msg:       data.Msg,
		Data:      []ExpressItem{},
	}
}

var rdb *redis.Client

func redisDb(ctx context.Context) {
	// 读取 CA 文件
	caCert, err := ioutil.ReadFile("./ssl/ca.crt")
	if err != nil {
		log.Fatalf("无法读取 CA 文件: %v", err)
	}

	// 创建 CertPool
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatal("无法将 CA 添加到 CertPool")
	}

	// TLS 配置
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false, // true 会跳过证书验证，测试用
	}

	// Redis 客户端配置
	rdb = redis.NewClient(&redis.Options{
		Addr:      "148.135.81.245:6380",
		Password:  "123123",
		DB:        0,
		TLSConfig: tlsConfig,
	})

	// 测试连接
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis TLS 连接失败: %v", err)
	}
	fmt.Println("Redis TLS 连接成功:", pong)
}

func main() {
	ctx := context.Background()

	redisDb(ctx)

	r := gin.Default()

	r.GET("/wuliu", func(c *gin.Context) {
		company := c.Query("company") // 没有传则为空
		code := c.Query("code")       // 没有传则为空

		tokenv2 := rdb.Get(ctx, "wuliu:tokenv2")
		cookie := rdb.Get(ctx, "wuliu:cookie")
		codeRedis := rdb.Get(ctx, "wuliu:code")
		com := rdb.Get(ctx, "wuliu:com")
		if code == "" {
			code = codeRedis.Val()
		}

		if company == "" {
			company = com.Val()
		}
		fmt.Println(tokenv2.Val())
		result := GetZTByBaidu(code, company, tokenv2.Val(), cookie.Val())
		fmt.Printf("%+v\n", result)

		c.JSON(http.StatusOK, result)
	})
	r.Run(":8081")
	//lis, _ := net.Listen("tcp", ":50051")
	//log.Println("gRPC server listening on :50051")
	//grpcServer := grpc.NewServer()
	//pb.RegisterHelloServiceServer(grpcServer, &HelloService{})
	//grpcServer.Serve(lis)

	//配置客户端
	//config := openai.DefaultConfig("blz_gmfXITCkIldgSw_zq3bPgCZlYFFBZlbxE1mAd2xxlJ4")
	//config.BaseURL = "https://blazeai.boxu.dev/api/"
	//
	//// 创建客户端实例
	//client := openai.NewClientWithConfig(config)
	//
	//// 构建请求
	//resp, err := client.CreateChatCompletion(
	//	context.Background(),
	//	openai.ChatCompletionRequest{
	//		Model: "qwen3.6-plus-thinking",
	//		Messages: []openai.ChatCompletionMessage{
	//			{
	//				Role:    openai.ChatMessageRoleUser,
	//				Content: "猜一猜下句话是什么",
	//			},
	//		},
	//	},
	//)
	//
	//// 错误处理
	//if err != nil {
	//	log.Fatalf("请求失败: %v", err)
	//}
	//
	//// 输出响应内容
	//fmt.Println(resp.Choices[0].Message.Content)
}

// 创建一个结构体，实现 pb.HelloServiceServer 接口
type HelloService struct {
	pb.UnimplementedHelloServiceServer
}

// 实现 SayHello 方法
func (s *HelloService) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{
		Message: "Hello, " + req.Name,
	}, nil
}
