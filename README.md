# PR Reviewer Assignment Service

## Инструкция по запуску
1. Клон репозитория

```
git clone https://github.com/evakaiing/PR-Reviewer-Assignment-Service.git
cd PR-Reviewer-Assignment-Service
```
2. Настройка переменных окружения
```
Добавьте нужные параметры окружения в .env согласно .env.example
```
2. Запуск с Docker
```
docker-compose up
```

Это запустит:
- PostgreSQL на localhost:5432
- Приложение на http://localhost:8080

3. Проверка работы
```
curl http://localhost:8080/health
```

## Тестирование

```
go test ./... -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html
```

### Проблемы, с которыми столкнулась
- Изначально коммитила не по мере выполнения задач. Решение: стала коммитить по мере реализации отдельных логических блоков. К сожалению, получилось не везде.

- При первой реализации метод team.Add выполнял только обновление существующих пользователей через UPDATE. В API указано: создать команду с участниками (создаёт/обновляет пользователей). Решение: добавила запрос с ON CONFLICT для вставки, если пользователя нет, и обновления. 

`hash-commit: a4d9dcbf714f888eae76e65bdfecc842a93f1c4b`

- Метод GetReview изначально возвращал PullRequest, по спецификации требовалось PullRequestShort. Решение: изменила возвращаемый тип, упростила SQL запрос, так как нам не нужны assigned_reviewers.

`hash-commit: a4d9dcbf714f888eae76e65bdfecc842a93f1c4b`

- При проектировании Reassign упустила replaced_by в response. Решение: изменила сигнатуру метода Reassign в repository и исправила тесты.

`hash-commit: 33b0e3eb0c63cae20dbcf5bf5bbdb23fd1c756ad`

- Не вся логика покрыта тестами. Решение: оставила TODO.