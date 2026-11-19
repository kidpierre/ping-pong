.PHONY: images deploy logs-ping logs-pong count interval restart test lint clean

images:
	docker build --build-arg SERVICE=ping -t ping:1.0 .
	docker build --build-arg SERVICE=pong -t pong:1.0 .

deploy: images
	minikube image load ping:1.0 --overwrite=true
	minikube image load pong:1.0 --overwrite=true
	kubectl apply -f k8s/
	kubectl rollout restart deploy/ping deploy/pong
	kubectl rollout status deploy/ping
	kubectl rollout status deploy/pong

logs-ping:
	kubectl logs -f deploy/ping

logs-pong:
	kubectl logs -f deploy/pong

count:
	@curl -s http://$$(minikube ip):30080/count
	@curl -s http://$$(minikube ip):30081/count

interval:
	@test -n "$(PING)$(PONG)" || { echo "usage: make interval PING=500ms PONG=1s"; exit 1; }
	$(if $(PING),kubectl patch configmap ping-pong-config --type merge -p '{"data":{"ping-interval":"$(PING)"}}')
	$(if $(PONG),kubectl patch configmap ping-pong-config --type merge -p '{"data":{"pong-interval":"$(PONG)"}}')
	$(if $(PING),kubectl rollout restart deploy/ping && kubectl rollout status deploy/ping)
	$(if $(PONG),kubectl rollout restart deploy/pong && kubectl rollout status deploy/pong)

restart:
	kubectl rollout restart deploy/ping deploy/pong

test:
	go test -race ./...

lint:
	go vet ./...

clean:
	kubectl delete -f k8s/ --ignore-not-found
