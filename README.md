NewYear API
REST API для управления коллекционными предметами (монеты, медали и т.д.) с системой аутентификации и авторизации.

Технологии
Go 1.21+

Gin (веб-фреймворк)

GORM (ORM)

PostgreSQL 16

Docker & Docker Compose

JWT (golang-jwt/jwt)

bcrypt (golang.org/x/crypto)

OAuth 2.0 (Yandex ID)

Быстрый старт
1. Клонировать репозиторий
git clone <url>
cd newyear-api

docker-compose down

2. Создать файл .env
cp .env.example .env

Заполните своими значениями (особенно YANDEX_CLIENT_ID и YANDEX_CLIENT_SECRET).

3. Запустить через Docker Compose
docker-compose up --build

API будет доступен на http://localhost:4200

Переменные окружения (.env.example)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=newyear_db

PORT=4200

JWT_ACCESS_SECRET=your_access_secret_here
JWT_REFRESH_SECRET=your_refresh_secret_here
JWT_ACCESS_EXPIRATION=15m
JWT_REFRESH_EXPIRATION=168h

YANDEX_CLIENT_ID=your_yandex_client_id
YANDEX_CLIENT_SECRET=your_yandex_client_secret
YANDEX_CALLBACK_URL=http://localhost:4200/auth/oauth/yandex/callback

COOKIE_SECURE=false
OAUTH_SUCCESS_REDIRECT_URL=http://localhost:4200/auth/whoami

API Endpoints
Аутентификация (публичные)
POST /auth/register - Регистрация
POST /auth/login - Вход
POST /auth/refresh - Обновление токенов
POST /auth/forgot-password - Запрос сброса пароля
POST /auth/reset-password - Установка нового пароля
GET /auth/oauth/yandex - Вход через Яндекс
GET /auth/oauth/yandex/callback - Callback от Яндекса

Аутентификация (приватные, требуют токен)
GET /auth/whoami - Данные текущего пользователя
POST /auth/logout - Выход (текущая сессия)
POST /auth/logout-all - Выход со всех устройств

Коллекционные предметы (приватные, требуют токен)
GET /items - Получить свои предметы
GET /items/:id - Получить по ID
POST /items - Создать новый
PUT /items/:id - Полное обновление
PATCH /items/:id - Частичное обновление
DELETE /items/:id - Мягкое удаление

Тестирование через Postman
Все запросы к /items и приватным /auth эндпоинтам требуют предварительного входа через /auth/login.
Токены передаются автоматически через HttpOnly cookies — Postman подхватывает их сам.

1. Регистрация
POST http://localhost:4200/auth/register
Content-Type: application/json

{
"email": "test@example.com",
"password": "password123"
}

Ответ (201):

{
"message": "User registered successfully",
"user": {
"id": "uuid",
"email": "test@example.com",
"created_at": "2024-01-01T00:00:00Z"
}
}

2. Вход
POST http://localhost:4200/auth/login
Content-Type: application/json

{
"email": "test@example.com",
"password": "password123"
}

Ответ (200):

{
"message": "Login successful"
}

После входа Postman автоматически сохранит cookies access_token и refresh_token.

3. Проверка текущего пользователя
GET http://localhost:4200/auth/whoami

Ответ (200):

{
"id": "uuid",
"email": "test@example.com",
"created_at": "2024-01-01T00:00:00Z"
}

Ответ без токена (401):

{
"error": "Unauthorized - no token"
}

4. Создание предмета
POST http://localhost:4200/items
Content-Type: application/json

{
"name": "Ancient coin",
"year": 1925,
"country": "USSR",
"price": 5000,
"condition": "excellent"
}

Ответ (201):

{
"message": "created"
}

5. Получение списка (с пагинацией)
GET http://localhost:4200/items?page=1&limit=10

Ответ (200):

{
"data": [...],
"meta": {
"total": 100,
"page": 1,
"limit": 10,
"totalPages": 10
}
}

6. Получение по ID
GET http://localhost:4200/items/{id}

7. Полное обновление (PUT)
PUT http://localhost:4200/items/{id}
Content-Type: application/json

{
"name": "Updated coin",
"year": 2000,
"country": "Russia",
"price": 9999,
"condition": "perfect"
}

Попытка обновить чужой предмет вернёт 403 Forbidden.

8. Частичное обновление (PATCH)
PATCH http://localhost:4200/items/{id}
Content-Type: application/json

{
"price": 77777
}

9. Удаление (DELETE)
DELETE http://localhost:4200/items/{id}

Ответ: 204 No Content

Soft Delete — запись остаётся в БД, но помечается как удалённая (поле deleted_at).

10. Обновление токенов
POST http://localhost:4200/auth/refresh

Используется refresh_token из cookie. Выдаёт новую пару токенов.

11. Выход
POST http://localhost:4200/auth/logout

Отзывает текущие токены и очищает cookies.

12. Выход со всех устройств
POST http://localhost:4200/auth/logout-all

Отзывает все токены пользователя во всех сессиях.

13. Сброс пароля
Шаг 1 — запросить сброс:

POST http://localhost:4200/auth/forgot-password
Content-Type: application/json

{
"email": "test@example.com"
}

Ответ (200):

{
"message": "If the account exists, reset instructions have been sent."
}

Токен сохраняется в папку tmp-mails/ (симуляция отправки письма для разработки).
Откройте файл из папки tmp-mails/ и скопируйте токен.

Шаг 2 — установить новый пароль:

POST http://localhost:4200/auth/reset-password
Content-Type: application/json

{
"token": "токен-из-файла-tmp-mails",
"new_password": "newpassword123"
}

Ответ (200):

{
"message": "Password reset successful"
}

14. Вход через Яндекс (OAuth 2.0)
Откройте в браузере:

GET http://localhost:4200/auth/oauth/yandex

Вас перенаправит на страницу входа Яндекса. После входа Яндекс вернёт на callback, и вы получите JWT cookies автоматически.

Безопасность
Пароли хешируются через bcrypt с автоматической уникальной солью. Исходный пароль восстановить невозможно.

Access и Refresh токены хранятся в БД в хешированном виде (SHA-256). Даже при утечке БД токены нельзя использовать напрямую.

HttpOnly cookies — JavaScript на клиенте не имеет доступа к токенам (защита от XSS).

SameSite=Lax — защита от CSRF атак.

state параметр в OAuth — защита от CSRF при авторизации через Яндекс.

Владение ресурсом — пользователь может редактировать и удалять только свои предметы.

Logout/LogoutAll — токены отзываются в БД, повторное использование невозможно.

Soft Delete
После удаления запись остаётся в БД, но помечается как удалённая.

Проверка в БД:

docker exec -it newyear-api_db psql -U postgres -d newyear_db -c "SELECT id, name, deleted_at FROM collectibles;"

Только удалённые:

docker exec -it newyear-api_db psql -U postgres -d newyear_db -c "SELECT id, name, deleted_at FROM collectibles WHERE deleted_at IS NOT NULL;"

Проверка что пароли разных пользователей имеют разные хеши:

docker exec -it newyear-api_db psql -U postgres -d newyear_db -c "SELECT id, email, password_hash FROM users;"



http://localhost:4200/auth/oauth/yandex



http://localhost:4200/api/docs/index.html


# 1. Запускаем всё
docker-compose up --build -d

# Кеширование Redis
1. Логирование Cache HIT/MISS

# Смотреть логи с фильтром
docker-compose logs -f app | findstr "Cache"

# Результат:
# Cache MISS for key: wp:items:user:9a485105:page:1:limit:10
# Cache HIT for key: wp:items:user:9a485105:page:1:limit:10
# Cache MISS for key: wp:users:profile:9a485105

2. Проверка кеша (GET /items)
# 1. Выполните POST /auth/login, затем GET /items
# В логах: Cache MISS

# 2. Повторите GET /items
# В логах: Cache HIT

# 3. Создайте POST /items (новый предмет)
# В Redis: старый ключ удалён

# 4. Снова GET /items
# В логах: Cache MISS (кеш инвалидирован, новые данные)



3. Проверка JTI токенов
# После POST /auth/login:
docker exec -it wp_labs_redis redis-cli -a redis_secure_password_change_in_prod

KEYS wp:auth:user:*:access:*
# Результат: wp:auth:user:9a485105:access:550e8400-e29b-41d4-a716-446655440000

TTL wp:auth:user:9a485105:access:550e8400...
# Результат: 900 (15 минут)

# После POST /auth/logout:
KEYS wp:auth:user:*:access:*
# Результат: (empty array) - ключ удалён


4. Итоговые команды для демонстрации

# Мониторинг логов (отдельное окно)
docker-compose logs -f app | findstr "Cache"

# Подключение к Redis
docker exec -it wp_labs_redis redis-cli -a redis_secure_password_change_in_prod

# Просмотр всех ключей кеша
KEYS wp:*

# Просмотр TTL (времени жизни)
TTL wp:items:user:9a485105:page:1:limit:10
TTL wp:users:profile:9a485105  # 3600 (1 час)

