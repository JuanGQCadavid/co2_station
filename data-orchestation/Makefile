run-drift:
	export mqtt_uri="mqtt://localhost:1883" && go run cmd/drift/main.go

run-exporter:
	source .env.local && go run cmd/exporter/main.go

run-dashboard:
	source .env.local && cd cmd/dashboard/ && go run main.go

