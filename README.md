# object-pool-service

### To generate `proto` module
```bash
protoc --go_out=service/proto --go_opt=paths=source_relative \
    --go-grpc_out=service/proto --go-grpc_opt=paths=source_relative \
    --js_out=library=client/src/proto
    object_pool.proto
```
