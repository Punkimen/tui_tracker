# Daily Tracker

Daily Tracker - терминальное приложение для отслеживания ежедневных привычек.
Приложение позволяет создавать привычки, отмечать значения по дням текущего
месяца, редактировать привычки и хранить историю в локальной SQLite-базе.

## Возможности

- создание привычек;
- поддержка типов привычек:
  - `progress` - прогресс;
  - `count` - количество;
  - `minutes` - минуты;
- ввод значений по дням месяца;
- редактирование, архивация и удаление привычек;
- локальное хранение данных в файле `tracker.db`.

## Требования

Перед запуском установите:

- Go версии `1.26.4` или совместимой версии, указанной в `go.mod`;
- терминал с поддержкой интерактивного TUI-интерфейса.

Проверить установленную версию Go:

```bash
go version
```

## Как запустить проект

1. Склонируйте репозиторий или скачайте проект на ПК:

```bash
git clone <URL_РЕПОЗИТОРИЯ>
cd daily_tracker
```

Если проект уже находится на компьютере, просто перейдите в папку проекта:

```bash
cd daily_tracker
```

2. Установите зависимости:

```bash
go mod download
```

3. Запустите приложение:

```bash
go run ./cmd/tracker
```

После запуска файл `tracker.db` автоматически создаётся в папке конфигурации
пользователя: `%AppData%\\daily-tracker` в Windows и обычно
`~/.config/daily-tracker` в Linux. В нём хранятся привычки и ежедневные записи.

## Релиз для Linux и Windows

При создании тега вида `v*` GitHub Actions запускает тесты, собирает готовые
архивы и создаёт GitHub Release с двумя файлами:

- `daily-tracker_linux_amd64.tar.gz` — для Linux x86_64;
- `daily-tracker_windows_amd64.zip` — для Windows x64.

Чтобы выпустить новую версию, закоммитьте изменения и отправьте тег в GitHub:

```bash
git tag v1.0.0
git push origin v1.0.0
```

Готовые файлы появятся во вкладке **Releases** репозитория. На Linux распакуйте
архив и запустите `./daily-tracker`; на Windows распакуйте ZIP и запустите
`daily-tracker.exe` из PowerShell или `cmd`.

Для локальной проверки релизной сборки выполните:

```bash
mkdir -p dist
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/daily-tracker ./cmd/tracker
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/daily-tracker.exe ./cmd/tracker
```

## Сборка исполняемого файла

Чтобы собрать приложение в отдельный бинарный файл:

```bash
go build -o daily-tracker ./cmd/tracker
```

Запуск собранного файла:

```bash
./daily-tracker
```

Для Windows можно собрать файл с расширением `.exe`:

```bash
go build -o daily-tracker.exe ./cmd/tracker
```

Запуск на Windows:

```powershell
.\daily-tracker.exe
```

## Управление

В приложении используются клавиши:

- `h` / `left` - перейти влево;
- `l` / `right` - перейти вправо;
- `j` / `down` - перейти вниз;
- `k` / `up` - перейти вверх;
- `enter` - выбрать кнопку или ячейку;
- `esc` - вернуться на предыдущий экран;
- `q` - выйти из приложения.

## Структура проекта

```text
cmd/tracker/main.go        точка входа приложения
internal/model             модели привычек и записей
internal/storage           работа с SQLite
internal/tracker           бизнес-логика трекера
internal/ui                терминальный интерфейс
go.mod                     описание Go-модуля и зависимостей
```

## Данные приложения

Все данные сохраняются локально в файле `tracker.db` в папке конфигурации
пользователя: `%AppData%\\daily-tracker` в Windows и обычно
`~/.config/daily-tracker` в Linux.

Если нужно начать с пустой базы, остановите приложение и удалите файл
`tracker.db`. При следующем запуске база будет создана заново.

## Проверка кода

Запустить тесты, если они будут добавлены в проект:

```bash
go test ./...
```
