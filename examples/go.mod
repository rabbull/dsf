module github.com/rabbull/dsf/example

go 1.17

replace github.com/rabbull/dsf => ../../dsf

require (
	github.com/bytedance/sonic v1.0.0
	github.com/go-redis/redis/v8 v8.11.4
	github.com/rabbull/dsf v0.0.1
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20211229061535-45e1f0233683 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
)
