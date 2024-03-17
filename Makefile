REPO_URL="xpkg.upbound.io/upbound/function-cidr"
VERSION_TAG="v0.1.0"
PACKAGE_FILES="function-amd64.xpkg,function-arm64.xpkg"

help:                   ## Print help for targets with comments
			@printf "For more targets and info see comments in Makefile.\n\n"
			@grep -E '^[a-zA-Z0-9._-]+:.*## .*$$' Makefile | sort | \
				awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

all:                    docker-build-amd64 docker-build-arm64 xpkg-build-amd64 xpkg-build-arm64 xpkg-push

docker-build-amd64:	## Build AMD64 Docker Image
			docker build . --quiet --platform=linux/amd64 --tag runtime-amd64

docker-build-arm64:	## Build ARM64 Docker Image
			docker build . --quiet --platform=linux/arm64 --tag runtime-arm64

xpkg-build-amd64:	## Build AMD64 Composition Function XPKG
			crossplane xpkg build \
				--package-root=package \
				--embed-runtime-image=runtime-amd64 \
				--package-file=function-amd64.xpkg

xpkg-build-arm64:	## Build ARM64 Composition Function XPKG
			crossplane xpkg build \
				--package-root=package \
				--embed-runtime-image=runtime-arm64 \
				--package-file=function-arm64.xpkg

xpkg-push:		## Push XPKG Package Files, Requires Upbound login
			up xpkg push ${REPO_URL}:${VERSION_TAG} -f ${PACKAGE_FILES}

fn-build:               ## Build Function Code
			go generate ./...
			go build .

test:                   ## Run Code Tests
			go test -v -cover .

render:                 ## Render Examples, Requires make debug first
			@for file in examples/xr-*.yaml; do \
				echo "Rendering $$file..."; \
				crossplane beta render \
					"$$file" \
					apis/composition.yaml \
					examples/functions.yaml; \
			done

debug:                  ## Run CIDR Function For Rendering Examples
			go run . --insecure --debug
