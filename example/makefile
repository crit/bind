build:
	@go build main.go

local: build
	@./main --name=John --age=23 --birthday=2006-01-06

debugger:
	@dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient
