# Daily Tracker: контекст для агентов

Этот файл нужен как быстрый вход в проект для будущих агентов. Он описывает текущее устройство кода, основные контракты и места, на которые стоит смотреть перед изменениями.

## Назначение проекта

Daily Tracker - терминальное Go-приложение для отслеживания ежедневных привычек. Пользователь создает привычки, видит таблицу дней текущего месяца и заполняет значения по датам. Данные хранятся локально в SQLite-файле `tracker.db`.

## Стек

- Go module: `daily-tracker`
- Go version в `go.mod`: `1.26.4`
- TUI: `charm.land/bubbletea/v2`
- Стили TUI: `github.com/charmbracelet/lipgloss`
- SQLite-драйвер: `github.com/glebarez/go-sqlite`

## Запуск и проверка

```bash
go run ./cmd/tracker
go test ./...
go build -o daily-tracker ./cmd/tracker
```

Приложение открывает базу через `storage.New("tracker.db")`, поэтому файл БД создается в текущей рабочей директории процесса запуска.

## Структура

- `cmd/tracker/main.go` - точка входа: создает SQLite storage, `tracker.Tracker`, Bubble Tea app и запускает TUI.
- `internal/model/habit.go` - доменные модели `Habit`, `Entry`, типы привычек.
- `internal/storage/storage.go` - интерфейс хранилища, от которого зависит бизнес-логика.
- `internal/storage/sqlite.go` - SQLite-реализация хранилища и миграция таблиц.
- `internal/tracker/tracker.go` - бизнес-логика поверх `storage.Storage`.
- `internal/ui/app.go` - главный экран с таблицей привычек по дням месяца.
- `internal/ui/form.go` - форма создания и редактирования привычки.
- `internal/ui/formHabit.go` - экран выбора привычки для обновления/удаления.
- `internal/ui/choose.go` - модальное меню выбора действия.
- `internal/ui/styles.go` - lipgloss-стили.
- `internal/ui/navigationHints.go` - рендер подсказок клавиш.

## Доменная модель

`model.Habit`:

- `ID int64`
- `Name string`
- `Type HabitType`
- `CreatedAt time.Time`
- `Goal *float64`
- `StartDate time.Time`
- `EndDate *time.Time`

`model.Entry`:

- `ID int64`
- `HabitID int64`
- `Date time.Time`
- `Value float64`

Типы привычек:

- `progress`
- `count`
- `minutes`

## Схема SQLite

Миграция находится в `internal/storage/sqlite.go` и выполняется при `storage.New`.

`habits`:

- `id INTEGER PRIMARY KEY AUTOINCREMENT`
- `name TEXT NOT NULL`
- `type TEXT NOT NULL`
- `goal REAL`
- `start_date TEXT NOT NULL`
- `end_date TEXT`
- `created_at TEXT NOT NULL`

`entries`:

- `id INTEGER PRIMARY KEY AUTOINCREMENT`
- `habit_id INTEGER NOT NULL REFERENCES habits(id)`
- `date TEXT NOT NULL`
- `value REAL NOT NULL`

Даты сохраняются строками в формате `2006-01-02`.

## Поток данных

1. `main.go` создает `storage.SqliteDB`.
2. `tracker.New(db)` принимает хранилище через интерфейс `storage.Storage`.
3. `ui.CreateApp(t)` создает `AppModel` и начальное состояние UI.
4. `AppModel.Init()` запускает `loadData()`.
5. `loadData()` читает привычки и записи текущего месяца через `tracker`.
6. Действия пользователя в TUI вызывают методы `tracker`, а после изменений UI обычно перезагружает данные через `loadData()` или `loadHabits()`.

## Бизнес-логика

`internal/tracker/tracker.go`:

- `AddHabit` выставляет `StartDate` и `CreatedAt` через `time.Now()` и сохраняет привычку.
- `GetHabits(date)` возвращает привычки, активные для месяца даты.
- `UpdateHabit` валидирует `id`, имя и тип, затем читает существующую привычку и обновляет имя/тип.
- `ArchiveHabit` ставит `end_date = now`, но текущий UI вместо архивации использует удаление.
- `DeleteHabit` вызывает удаление привычки.
- `GetEntries(date)` возвращает записи месяца.
- `GetEntryByCurrentDate(habitID, date)` ищет запись по привычке и точной дате среди записей месяца.
- `AddEntry`, `UpdateEntry`, `DeleteEntry` проксируют операции в storage.

## Поведение UI

Главный экран (`AppModel`) показывает таблицу привычек и дней текущего месяца:

- верхние кнопки: `Create Habit`, `Update Habits`;
- фокус переключается между кнопками и таблицей;
- таблица поддерживает горизонтальный и вертикальный viewport по размеру окна;
- курсор по умолчанию стоит на текущем дне;
- для `count` нажатие `enter` по ячейке переключает значение между `1` и `0`;
- для остальных типов `enter` переводит ячейку в режим ввода числа, повторный `enter` сохраняет значение;
- `esc` отменяет ввод в ячейке.

Форма привычки (`FormModel`):

- используется и для создания, и для редактирования;
- поле имени редактируется вручную через обработку keypress;
- тип выбирается между `progress`, `count`, `minutes`;
- при сохранении новой привычки вызывается `Tracker.AddHabit`;
- при редактировании вызывается `Tracker.UpdateHabit`.

Экран `FormHabitModel`:

- показывает привычки сеткой;
- `enter` открывает модальное меню с действиями `Update Habit` и `Delete Habit`;
- удаление физически удаляет привычку из БД.

## Клавиши

- `h` / `left` - влево
- `l` / `right` - вправо
- `j` / `down` - вниз
- `k` / `up` - вверх
- `enter` - выбрать/сохранить/начать ввод
- `esc` - назад или отменить текущий ввод
- `q` - выход

## Важные нюансы текущей реализации

- В `entries` нет уникального ограничения на пару `(habit_id, date)`. UI пытается обновлять существующую запись через `GetEntryByCurrentDate`, но storage сам дубликаты не запрещает.
- `DeleteHabit` удаляет только строку из `habits`; в схеме нет явного `ON DELETE CASCADE`, поэтому поведение записей зависит от настроек SQLite foreign keys. Сейчас код не включает `PRAGMA foreign_keys = ON`.
- `ArchiveHabit` есть в бизнес-логике и storage, но в UI не подключен.
- `goal` хранится в модели и БД, но форма сейчас не дает пользователю вводить цель.
- `GetDaysFromMoth(month time.Month)` принимает `month`, но фактически использует `time.Now()` и игнорирует параметр.
- `GetHabitByID` использует `Query` без `defer rows.Close()` и возвращает пустой `Habit`, если запись не найдена.
- В `renderTable` при отрисовке каждой видимой ячейки вызывается `Tracker.GetEntryByCurrentDate`, который снова читает entries месяца из storage. На больших таблицах это может быть дорого; в `AppModel.entry` уже есть загруженные entries, но они почти не используются для рендера.
- В репозитории есть бинарник `main` и локальная база `tracker.db`; обычно их не стоит менять в задачах по коду.

## Рекомендации для следующих изменений

- Сохранять зависимость бизнес-логики от `storage.Storage`, не привязывать `tracker` к SQLite напрямую.
- При изменениях хранения сначала обновлять интерфейс `storage.Storage`, затем SQLite-реализацию, затем `tracker` и UI.
- Для новых проверок удобно добавлять тесты на `internal/tracker` через fake storage, потому что `Tracker` уже зависит от интерфейса.
- Для изменений в TUI проверять минимум `go test ./...` и ручной запуск `go run ./cmd/tracker`, потому что автоматических UI-тестов сейчас нет.
- Не удалять пользовательский `tracker.db` без явной просьбы.
