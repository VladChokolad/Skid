# 🚀 Быстрый запуск проекта Skid

Короткая памятка: от команд до проверки.

## 1. Требования
- Docker Desktop (PostgreSQL в контейнере)
- Go 1.26+
- VS Code + Live Server (для фронтенда)

---

## 2. База данных (Docker)

```bash
# Поднять БД (контейнер skid-db, порт 5432)
docker compose up -d db

# Убедиться, что контейнер запущен
docker ps   # должен быть виден skid-db
```

Создать таблицы (8 шт.) — выполнить CREATE TABLE в psql:
```bash
docker exec -it skid-db psql -U postgres -d skid
```

Вставить все запросы (применять строго по порядку — с учётом внешних ключей):

```sql
-- 1. Пользователи
CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    phone         VARCHAR(20),
    profile_image TEXT,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2. Анонимные пользователи
CREATE TABLE IF NOT EXISTS anonymous_users (
    id            SERIAL PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    phone         VARCHAR(20),
    access_token  VARCHAR(255) NOT NULL UNIQUE,
    last_activity TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at    TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 3. Вечеринки
CREATE TABLE IF NOT EXISTS parties (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    party_image TEXT,
    owner_id    INTEGER NOT NULL,
    invite_code VARCHAR(50) NOT NULL UNIQUE,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 4. Участники
CREATE TABLE IF NOT EXISTS participants (
    id                   SERIAL PRIMARY KEY,
    party_id             INTEGER NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    user_or_anonymous_id INTEGER,
    name                 VARCHAR(100) NOT NULL,
    is_admin             BOOLEAN NOT NULL DEFAULT FALSE,
    is_anonymous         BOOLEAN NOT NULL DEFAULT FALSE,
    is_placeholder       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at           TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 5. Иконки покупок
CREATE TABLE IF NOT EXISTS purchase_icons (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    icon TEXT NOT NULL
);

-- 6. Покупки
CREATE TABLE IF NOT EXISTS purchases (
    id                SERIAL PRIMARY KEY,
    party_id          INTEGER NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    buyer_id          INTEGER NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    name              VARCHAR(255) NOT NULL,
    description       TEXT,
    purchase_icon_id  INTEGER REFERENCES purchase_icons(id),
    price             DECIMAL(10,2) NOT NULL,
    split_type        INTEGER NOT NULL,
    date_of_purchase  TIMESTAMP,
    created_at        TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 7. Долги
CREATE TABLE IF NOT EXISTS debts (
    id             SERIAL PRIMARY KEY,
    purchase_id    INTEGER NOT NULL REFERENCES purchases(id) ON DELETE CASCADE,
    participant_id INTEGER NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    split_value    DECIMAL(10,2) NOT NULL
);

-- 8. Переводы
CREATE TABLE IF NOT EXISTS payments (
    id                  SERIAL PRIMARY KEY,
    party_id            INTEGER NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    from_participant_id INTEGER NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    to_participant_id   INTEGER NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    amount              DECIMAL(10,2) NOT NULL,
    note                TEXT,
    is_confirmed        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMP NOT NULL DEFAULT NOW()
);
```

> ⚠️ Имена колонок (`split_value`, `last_activity`) подобраны под реальные SQL-запросы в Go-коде, а не под диаграмму `Sqlsheme` — иначе INSERT/UPDATE будут падать.

Проверка, что таблицы есть:
```bash
docker exec -it skid-db psql -U postgres -d skid -c "\dt"
```

> ⚠️ **Важно:** если на порту 5432 уже что-то висит (старый PostgreSQL), останови его, иначе бэкенд попадёт не в ту БД.

### Как проверить и освободить порт 5432

**1. Кто слушает порт 5432?** В PowerShell:
```powershell
netstat -ano | findstr ":5432" | findstr "LISTENING"
```
В выводе последняя колонка — это **PID** процесса. Например `TCP 0.0.0.0:5432 ... 19008`.

**2. Что за процесс?** По PID:
```powershell
Get-Process -Id <PID> | Select-Object ProcessName, Path
```
- Если это `postgres.exe` из `C:\Program Files\PostgreSQL\...` — это локальный PostgreSQL, его нужно остановить.
- Если это `com.docker.backend` — это прокси Docker, порт уже занят контейнером, ничего делать не нужно.

**3. Остановить локальный PostgreSQL (Windows-сервис):**
```powershell
# Показать все сервисы postgres
Get-Service | Where-Object {$_.Name -like "*postgres*"} | Select-Object Name, Status

# Остановить нужный (например postgresql-x64-18)
Stop-Service postgresql-x64-18
# Или через сеть
net stop postgresql-x64-18
```

**4. Если это процесс не из списка сервисов** (например запущен вручную) — завершить по PID:
```powershell
Stop-Process -Id <PID> -Force
```

**5. Проверить, что порт освободился:**
```powershell
netstat -ano | findstr ":5432"
```
Пусто (кроме прокси Docker `com.docker.backend`) → порт свободен, поднимай контейнер `docker compose up -d db` и проверяй `\dt`.

> Почему это важно: бэкенд ходит на `localhost:5432`. Если порт держит **старый** PostgreSQL без таблиц, регистрация падает с `pq: отношение "users" не существует` — а в Docker-базе таблицы при этом есть.

---

## 3. Бэкенд (Go, порт 8080)

```bash
cd Backend

# 1) Файл .env рядом с main.go (без него не запустится):
#    DB_HOST=localhost
#    DB_PORT=5432
#    DB_USER=postgres
#    DB_PASSWORD=твой_пароль
#    DB_NAME=skid
#    SERVER_PORT=8080
#    JWT_SECRET=любая_строка   ← ОБЯЗАТЕЛЬНО
#    FRONTEND_URL=localhost:5500

# 2) Скачать зависимости и запустить
go mod tidy
go run .
```

Проверка:
```bash
curl http://localhost:8080/echo
```
В логах: `Сервер запущен на порту: 8080`.

---

## 4. Фронтенд (порт 5500)

```bash
cd Frontend
python -m http.server 5500
```
Или открой папку `Frontend` через **Live Server** в VS Code (порт 5500).

Открыть: `http://localhost:5500/main.html`

---

## 5. Проверка всего стека
1. БД запущена, `\dt` показывает 8 таблиц.
2. Бэкенд отвечает на `/echo`.
3. В браузере консоль: `🚀 Local Development Mode`, `API URL: http://localhost:8080`.
4. Регистрация/логин работают.

Всё зелёное → проект работает локально.