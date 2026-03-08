module github.com/HernanSarmiento/test-auth-authorization-grpc/blog-service

go 1.25.5

replace github.com/HernanSarmiento/test-auth-authorization-grpc/api => ../api

require (
	github.com/HernanSarmiento/test-auth-authorization-grpc/api v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	google.golang.org/protobuf v1.36.11
	gorm.io/gorm v1.31.1
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/grpc v1.79.1 // indirect
)
