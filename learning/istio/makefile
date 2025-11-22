install:
	istioctl install --set profile=demo -y

check:
	kubectl get pods -n istio-system

label:
	kubectl label namespace default istio-injection=enabled

see-label:
	kubectl get namespace --show-labels

restart:
	kubectl rollout restart deployment

verify-sidecar:
	kubectl exec -it url-shortener-canary-deployment-6ddd644b57-hvhqg -c istio-proxy -- curl localhost:15000/config_dump

jaeger-addon:
	kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.28/samples/addons/jaeger.yaml

install-jaeger-extension-provider:
	istioctl install -f jaeger.yaml --skip-confirmation

enable-tracing:
	kubectl apply -f tracing.yaml	

unlabel:
	kubectl label namespace default istio-injection-

uninstall:
	istioctl uninstall --purge -y


