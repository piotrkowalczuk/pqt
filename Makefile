gen:
	@cd example && go install ./generator
	@cd example && go generate ./app/internal/model

run:
	@cd example && go run app/main.go

test:
	@./test.sh

