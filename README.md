Сервис, возвращающий количество дней до Нового года.



\## Запуск через Docker



```bash

\# Сборка образа

docker build -t newyear-api .



\# Запуск контейнера

docker run -p 4200:4200 newyear-api

