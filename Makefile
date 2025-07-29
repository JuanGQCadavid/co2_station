
mosquitto:
	docker run -it -d --name mosquitto \
		-p 1883:1883 \
		-v ./mosquitto/config:/mosquitto/config \
		-v ./mosquitto/data:/mosquitto/data \
		-v ./mosquitto/log:/mosquitto/log \
		eclipse-mosquitto


run:
	git pull && docker compose up --build -d

stop:
	docker compose down