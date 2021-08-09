fmt:
	go fmt ./...

image:
	docker build -t umi0410/streaming-chat:latest .

push:
	docker push umi0410/streaming-chat:latest
deploy:
	kubectl apply -f k8s
	kubectl rollout restart deployment server client

fwd:
	sudo -E kubefwd services -n default
cicd: fmt image push deploy

login:
	go run *.go --serverAddr server:55555 client