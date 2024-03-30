# grpc-server
grpc服务端， 实现了grpc拦截器，链路追踪等功能，配合sxp-server使用
grpc_middleware支持链式拦截器，可以使用官方提供的：
- Auth： grpc_auth
- Logging： grpc_ctxtags，grpc_zap， grpc_logrus，grpc_kit
- Monitoring: grpc_prometheus, grpc_opentracing
- Client: grpc_retry
- Server: grpc_validator,grpc_recovery,ratelimit

除了这些middleware，我们还可以实现定制的拦截器，实现自己的功能,本项目自己实现了一个token校验的拦截器供参考。

