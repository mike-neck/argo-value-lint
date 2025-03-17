.PHONY: build
build:
	go build -o build/argo-wf-lint ./*.go



goScripts := $(wildcard scripts/*.go)

.PHONY: $(patsubst scripts/%.go,run-%,$(goScripts))
define RunScript
$(2): $(1)
	@echo $(@)
	@[[ -d "data" ]] || mkdir -p "data"
	@go run $(1) | tee "data/$(2).json"

endef

$(foreach src,$(goScripts),$(eval $(call RunScript,$(src),$(patsubst scripts/%.go,run-%,$(src)))))
