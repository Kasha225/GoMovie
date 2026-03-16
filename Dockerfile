FROM golang:1.25.4

# Рабочая директория
WORKDIR /app

# Копируем go.mod и go.sum и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем приложение
RUN go build -o app .

# Копируем скрипт wait-for-it в контейнер
COPY wait-for-it.sh /wait-for-it.sh

# Делаем его исполняемым
RUN ["chmod", "+x", "/wait-for-it.sh"]

# Открываем порт приложения
EXPOSE 8080

# Команда запуска приложения с ожиданием базы
CMD ["/wait-for-it.sh", "db:5432", "--timeout=30", "--", "./app"]