container_runtime := $(shell which podman || which docker)

$(info using ${container_runtime})

up: down
	${container_runtime} compose up --build -d

down:
	${container_runtime} compose down

clean:
	${container_runtime} compose down -v

run-tests: 
	${container_runtime} run --rm --network=host tests:latest

test:
	make clean
	make up
	@echo wait cluster to start && sleep 10
	make run-tests
	make clean
	@echo "test finished"

lint:
	make -C search-services lint

proto:
	make -C search-services protobuf

tools:
	go install github.com/yoheimuta/protolint/cmd/protolint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.4.0
	@echo "checking protobuf compiler, if it fails follow guide at https://protobuf.dev/installation/"
	@which -s protoc && echo OK || exit 1

tidy:
	make -C search-services tidy

unit:
	make -C search-services test
	mv search-services/cover.html .

swagger:
	${container_runtime} compose up swagger

help:
	@echo "Доступные команды:"
	@echo "  make up       		- Вкл"
	@echo "  make down     		- Выкл"
	@echo "  make clean    		- Очистка"
	@echo "  make test     		- Запустить тесты"
	@echo "  make run-tests		- Запустить тест-контейнер"
	@echo "  make lint     		- Запустить линтер"
	@echo "  make proto    		- Обновить proto файлы"
	@echo "  make tools    		- Установить зависимости"
	@echo "  make swagger  		- Генерация swagger документации"
	@echo "  make tidy      	- Почистить зависимости в проекте"
	@echo "  make unit    		- Проверка тестового покрытия кода с генерацией html страницы"
