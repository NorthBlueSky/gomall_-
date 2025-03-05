// Code generated by hertz generator.

package main

import (
	"context"
	"os"
	"time"

	"github.com/cloudwego/biz-demo/gomall/app/frontend/biz/router"
	frontendBizUtils "github.com/cloudwego/biz-demo/gomall/app/frontend/biz/utils"
	"github.com/cloudwego/biz-demo/gomall/app/frontend/conf"
	"github.com/cloudwego/biz-demo/gomall/app/frontend/infra/rpc"
	"github.com/cloudwego/biz-demo/gomall/app/frontend/middleware"
	frontendUtils "github.com/cloudwego/biz-demo/gomall/app/frontend/utils"
	"github.com/cloudwego/biz-demo/gomall/common/mtl"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/cors"
	"github.com/hertz-contrib/gzip"
	"github.com/hertz-contrib/logger/accesslog"
	hertzlogrus "github.com/hertz-contrib/logger/logrus"
	prometheus "github.com/hertz-contrib/monitor-prometheus"
	hertzobslogrus "github.com/hertz-contrib/obs-opentelemetry/logging/logrus"
	hertztracing "github.com/hertz-contrib/obs-opentelemetry/tracing"
	"github.com/hertz-contrib/pprof"
	"github.com/hertz-contrib/sessions"
	"github.com/hertz-contrib/sessions/redis"
	"github.com/joho/godotenv"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	ServiceName  = frontendUtils.ServiceName
	MetricsPort  = conf.GetConf().Hertz.MetricsPort
	RegistryAddr = conf.GetConf().Hertz.RegistryAddr
)

func main() {
	_ = godotenv.Load()
	p := mtl.InitTracing(ServiceName)
	defer p.Shutdown(context.Background())
	// init dal
	// dal.Init()
	consul, registryInfo := mtl.InitMetric(ServiceName, MetricsPort, RegistryAddr)
	defer consul.Deregister(registryInfo)
	rpc.Init()
	address := conf.GetConf().Hertz.Address

	tracer, cfg := hertztracing.NewServerTracer()

	h := server.New(server.WithHostPorts(address), server.WithTracer(prometheus.NewServerTracer(
		"",
		"",
		prometheus.WithDisableServer(true),
		prometheus.WithRegistry(mtl.Registry),
	)), tracer)

	h.Use(hertztracing.ServerMiddleware(cfg))
	registerMiddleware(h)

	// add a ping route to test
	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong"})
	})

	router.GeneratedRegister(h)
	h.LoadHTMLGlob("template/*")
	h.Static("/static", "./")

	// ProtectGroup := h.Group("/protect")
	// ProtectGroup.Use(middleware.AuthMiddleware())

	// ProtectGroup.GET("/about", func(c context.Context, ctx *app.RequestContext) {
	// 	ctx.HTML(consts.StatusOK, "about", frontendUtils.WarpResponse(c, ctx, utils.H{"Title": "About"}))
	// })

	h.GET("/about", func(c context.Context, ctx *app.RequestContext) {
		hlog.CtxInfof(c, "CloudWeGo shop about page")
		ctx.HTML(consts.StatusOK, "about", frontendBizUtils.WarpResponse(c, ctx, utils.H{"Title": "About"}))
	})

	h.GET("/sign-in", func(c context.Context, ctx *app.RequestContext) {
		data := utils.H{
			"Title": "Sign In",
			"Next":  ctx.Query("next"),
		}
		ctx.HTML(consts.StatusOK, "sign-in", data)
	})

	h.GET("/sign-up", func(c context.Context, ctx *app.RequestContext) {
		ctx.HTML(consts.StatusOK, "sign-up", utils.H{"Title": "Sign Up"})
	})
	h.Spin()
}

func registerMiddleware(h *server.Hertz) {
	store, _ := redis.NewStore(10, "tcp", conf.GetConf().Redis.Address, "", []byte(os.Getenv("SESSION_SECRET")))
	h.Use(sessions.New("qitian-shop", store))
	// log
	logger := hertzobslogrus.NewLogger(hertzobslogrus.WithLogger(hertzlogrus.NewLogger().Logger()))
	// logger := hertzlogrus.NewLogger()
	hlog.SetLogger(logger)
	hlog.SetLevel(conf.LogLevel())
	var flushInterval time.Duration
	if os.Getenv("GO_ENV") == "online" {
		flushInterval = time.Minute
	} else {
		flushInterval = time.Second
	}
	asyncWriter := &zapcore.BufferedWriteSyncer{
		WS: zapcore.AddSync(&lumberjack.Logger{
			Filename:   conf.GetConf().Hertz.LogFileName,
			MaxSize:    conf.GetConf().Hertz.LogMaxSize,
			MaxBackups: conf.GetConf().Hertz.LogMaxBackups,
			MaxAge:     conf.GetConf().Hertz.LogMaxAge,
		}),
		FlushInterval: flushInterval,
	}
	hlog.SetOutput(asyncWriter)
	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		asyncWriter.Sync()
	})

	// pprof
	if conf.GetConf().Hertz.EnablePprof {
		pprof.Register(h)
	}

	// gzip
	if conf.GetConf().Hertz.EnableGzip {
		h.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	// access log
	if conf.GetConf().Hertz.EnableAccessLog {
		h.Use(accesslog.New())
	}

	// recovery
	h.Use(recovery.Recovery())

	// cores
	h.Use(cors.Default())

	middleware.Register(h)
}
