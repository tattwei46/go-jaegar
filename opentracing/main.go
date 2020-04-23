package opentracing

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	"log"
	"net/http"
	"time"
)

func Init(service string){
	udpTransport, _ := jaeger.NewUDPTransport("localhost:5775", 0)
	reporter := jaeger.NewRemoteReporter(udpTransport)
	sampler := jaeger.NewConstSampler(true)
	tracer, _ := jaeger.NewTracer(service, sampler, reporter)
	opentracing.SetGlobalTracer(tracer)
}

func IntroduceSpan(ctx context.Context, spanName string)(opentracing.Span, context.Context){
	return opentracing.StartSpanFromContext(ctx, spanName)
}

func Serialize(ctx context.Context, req *http.Request){
	req = req.WithContext(ctx)
	if span := opentracing.SpanFromContext(ctx); span != nil {
		opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(req.Header))
	}
}

func Deserialize(r *http.Request, spanName string)(opentracing.Span, *http.Request){
	time.Sleep(250*time.Millisecond)
	var serverSpan opentracing.Span
	appSpecificOperationName := spanName
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		log.Fatal(err)
	}
	serverSpan = opentracing.StartSpan(appSpecificOperationName,
		ext.RPCServerOption(wireContext))
	ctx := opentracing.ContextWithSpan(context.Background(), serverSpan)
	r = r.WithContext(ctx)
	return serverSpan, r
}

func HttpMiddleware(serverName string, h http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		span, newCtx := opentracing.StartSpanFromContext(r.Context(), serverName)
		defer func(){
			span.Finish()
		}()
		r = r.WithContext(newCtx)
		h.ServeHTTP(w, r)
	})
}
