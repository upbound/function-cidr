generate:
			go generate ./...

build:
			docker build . --quiet --platform=linux/amd64 --tag runtime-amd64
			crossplane xpkg build --package-root=package --embed-runtime-image=runtime-amd64 --package-file=function-amd64.xpkg

render:
			@for file in examples/xr-*.yaml; do \
			 	echo ""; \
				echo "Rendering $$file..."; \
				crossplane beta render \
					"$$file" \
					apis/composition.yaml \
					examples/functions.yaml; \
			done

render-pipeline:
			crossplane beta render examples/xr-cidrsubnet.yaml \
			  apis/composition-pipeline.yaml \
				examples/functions.yaml

render-pipeline-context:
			crossplane beta render examples/xr-cidrsubnet.yaml \
			  apis/composition-pipeline-context.yaml \
				examples/functions-pipeline-context.yaml \
				--extra-resources=examples/extraResources.yaml

debug:
			go run . --insecure --debug