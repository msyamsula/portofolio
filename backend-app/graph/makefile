build:
	docker build --platform linux/amd64,linux/arm64 -t graph .

run:
	docker run -it --name graph --network=test --rm \
	-p 10000:10000 \
	-e JAEGER_HOST=jaeger \
	-e PORT=10000 \
	-e CALLBACK_URI=http://localhost:10000 graph