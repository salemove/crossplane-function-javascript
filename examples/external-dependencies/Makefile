.PHONY: render run build

render: build
	@crossplane beta render -r xr.yaml composition.yaml functions.yaml

run:
	$(MAKE) -C ../.. run

build: composition.yaml node_modules

composition.yaml: build.mjs src/index.js package.json package-lock.json
	@npm run build

node_modules:
	@npm i
