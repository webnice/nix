## Инструментарий проекта - Makefile.
## Версия: 10.11.2026.

DIR                  := $(strip $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))))
GOPATH               := $(GOPATH)
DATE                 := $(shell date -u +%Y%m%d.%H%M%S.%Z)
GOGENERATE            = $(shell if [ -f .gogenerate ]; then cat .gogenerate; fi)
TESTPACKETS           = $(shell if [ -f .testpackages ]; then cat .testpackages; fi)
BENCHPACKETS          = $(shell if [ -f .benchpackages ]; then cat .benchpackages; fi)
GO111MODULE          ?= $(GO111MODULE:on)
PACKAGES_LOCK_VER     = $(shell if [ -f .packages_lock_version ]; then cat .packages_lock_version; fi)
COVERALLS_REPO_TOKEN  = $(shell if [ -f .coveralls ]; then cat .coveralls; fi)

default: help

link:
.PHONY: link

## Загрузка зависимостей.
dep-init:
	@rm -rf ${DIR}/vendor 2>/dev/null; true
.PHONY: dep-init
dep: dep-init
	@go mod download
	@go mod tidy
	@go mod vendor
.PHONY: dep

update: upd
upd:
	@go clean -cache -modcache
	@go get -u ./...
	@for item in $(PACKAGES_LOCK_VER); do \
		go get -u "$${item}"; \
		true; \
	done
.PHONY: upd
.PHONY: update

## Кодогенерация (использовать только во время разработки).
## Все генерируемые пакеты перечислены в файле .gogenerate.
gen:
	@for PKGNAME in $(GOGENERATE); do GOPATH="$(DIR)" go generate $${PKGNAME}; done
.PHONY: gen

## Тестирование одного или нескольких пакетов, а также приложений, с отчётом о проценте покрытия кода тестами.
# Все тестируемые пакеты перечислены в файле .testpackages.
test: link
	@echo "mode: set" > coverage.log
	@for PACKET in $(TESTPACKETS); do \
		touch coverage-tmp.log; \
		unset GOPATH; go test -v -covermode=count -coverprofile=coverage-tmp.log $$PACKET; \
		if [ "$$?" -ne "0" ]; then exit $$?; fi; \
		tail -n +2 coverage-tmp.log | sort -r | awk '{if($$1 != last) {print $$0;last=$$1}}' >> coverage.log; \
		rm -f coverage-tmp.log; true; \
	done
	@if [ "$(COVERALLS_REPO_TOKEN)" != "" ]; then\
		echo "Загрузка в coveralls.io процента покрытия кода тестами."; \
		goveralls -repotoken $(COVERALLS_REPO_TOKEN) >/dev/null 2>&1; true; \
	fi
.PHONY: test

	## Отображение в браузере покрытия тестируемого кода по html отчёту (использовать только во время разработки).
cover: test
	unset GOPATH; go tool cover -html=$(DIR)/coverage.log
.PHONY: cover

## Тестирование производительности.
# Все пакеты, тестируемые на производительность, перечислены в файле .benchpackages.
bench:
	@for PACKET in $(BENCHPACKETS); do GOPATH=${GOPATH} go test -race -bench=. -benchmem $$PACKET; done
.PHONY: bench

## Тестирование качества кода.
# https://github.com/alecthomas/gometalinter/
# install: curl -L https://git.io/vp6lP | sh
lint: link
	gometalinter \
	--vendor \
	--deadline=15m \
	--cyclo-over=20 \
	--disable=aligncheck \
	--disable=gotype \
	--skip=vendor \
	--skip=src/vendor \
	--linter="vet:go tool vet -printf {path}/*.go:PATH:LINE:MESSAGE" \
	./...
.PHONY: lint

## Очистка проекта от временных файлов.
clean:
	@if [ -d ${DIR}/pkg ]; then\
		chown -R `whoami` ${DIR}/pkg/; true; \
		chmod -R 0777 ${DIR}/pkg/; true; \
	fi
	@rm -rf ${DIR}/src; true
	@rm -rf ${DIR}/bin; true
	@rm -rf ${DIR}/pkg; true
	@rm -rf ${DIR}/rpmbuild; true
	@rm -rf ${DIR}/*.log; true
	@rm -rf ${DIR}/*.lock; true
.PHONY: clean

## Справка по основным целям.
help:
	@echo "Usage: make [target]"
	@echo "  target is:"
	@echo "    gen                  - Кодогенерация с использованием go generate."
	@echo "    test                 - Запуск тестов проекта."
	@echo "    cover                - Запуск тестов проекта с отображением процента покрытия кода тестами."
	@echo "    bench                - Запуск тестов производительности."
	@echo "    lint                 - Запуск проверки кода с помощью gometalinter."
	@echo "    clean                - Очистка папки проекта от временных файлов."
.PHONY: help
