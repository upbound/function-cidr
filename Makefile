render:
			@for file in examples/xr-*.yaml; do \
				echo "Rendering $$file..."; \
				crossplane beta render \
					"$$file" \
					apis/composition.yaml \
					examples/functions.yaml; \
			done
