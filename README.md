# ShortLink
___

## Краткое описание
___
Сервис сокращения ссылок на Golang.

В качестве хранилища информации выступает map  или postgreSQL в зависимости от флага запуска программы.

В файле "linkFromDB.txt" содержатся найстройка подключения к бд

## Пример

Запускаем main.go файл с флагом "-d"
```sh
go run main.go -d
```

Отправляем Post-запрос с оригинаьной ссылкой и получаем в ответе короткую.
```sh
Request:
curl -X POST -d "http://cjdr17afeihmk.biz/kdni9/z9d112423421" http://localhost:8080/
Response:
http://localhost:8080/88d2d0f8fe07c98da23165c7a8a7acae
```
Отправляем Get-запрос с короткой ссылкой и получаем в ответе оригинальную.
```sh
Request:
curl -X GET http://localhost:8080/88d2d0f8fe07c98da23165c7a8a7acae
Response:
http://cjdr17afeihmk.biz/kdni9/z9d112423421
```
