package busi

import (
	"fmt"
	"time"

	v1 "github.com/Spacescore/observatory-task/internal/busi/api/v1"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *HttpServer) registerV1(r *gin.Engine) {
	apiv1 := r.Group("/api/v1")
	{
		apiv1.GET("/ping", v1.Ping)
	}
}

func (s *HttpServer) RegisterRoutes(r *gin.Engine) {
	// r.Use(utils.Cors())
	r.Use(cors.Default())
	r.GET("/meta-task/swagger/*any", swagHandler)

	s.registerV1(r)
}

func (s *HttpServer) Start() {
	// // if Flags.Mode == "prod" {
	gin.SetMode(gin.ReleaseMode)
	// // }

	// r := gin.Default()
	r := gin.New()
	// LoggerWithFormatter middleware will write the logs to gin.DefaultWriter
	// By default gin.DefaultWriter = os.Stdout
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// your custom format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	r.Use(gin.Recovery())
	s.RegisterRoutes(r)
	r.Run(s.addr)
}

type HttpServer struct {
	addr string
}

func NewHttpServer(addr string) *HttpServer {
	return &HttpServer{addr}
}

func HttpServerStart(addr string) {
	NewHttpServer(addr).Start()
}
