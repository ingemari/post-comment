# Приложение для постов и комментариев

GraphQL-приложение на Go для управления постами и комментариями с иерархической структурой и реальным временем через подписки. Поддерживает in-memory и PostgreSQL хранилища.

## Возможности
- Создание и просмотр постов.
- Добавление и просмотр иерархических комментариев.
- Запрет комментариев для постов.
- Пагинация комментариев и ответов.
- Уведомления о новых комментариях через GraphQL Subscriptions.
- Хранилища: in-memory или PostgreSQL (через `STORAGE_TYPE`).
- Потокобезопасность.
- Unit-тесты.
- Docker-развертывание.

## GraphQL API
### Запросы
- `posts`:
  ```graphql
  query {
    posts {
      id
      title
      content
      author
      allowComments
    }
  }
  ```
- `post`:
  ```graphql
  query {
    post(id: "post-id") {
      id
      title
      comments(limit: 2, offset: 0) {
        id
        text
        replies(limit: 2, offset: 0) {
          id
          text
        }
      }
    }
  }
  ```

### Мутации
- `createPost`:
  ```graphql
  mutation {
    createPost(title: "Test", content: "Content", author: "Author", allowComments: true) {
      id
    }
  }
  ```
- `addComment`:
  ```graphql
  mutation {
    addComment(postId: "post-id", author: "User", text: "Comment") {
      id
      text
    }
  }
  ```

### Подписки
- `commentAdded`:
  ```graphql
  subscription {
    commentAdded(postId: "post-id") {
      id
      text
    }
  }
  ```

## Тестирование
```bash
go test ./...
```

## Структура проекта
```
post-comment-app/
├── graph/
│   ├── generated.go
│   ├── resolver.go
│   ├── schema.resolvers.go
│   ├── schema.resolvers_test.go
├── storage/
│   ├── storage.go
│   ├── inmemory.go
│   ├── postgres.go
│   ├── postgres_test.go
├── server.go
├── schema.sql
├── Dockerfile
├── docker-compose.yml
├── .env
├── .env.example
├── .gitignore
├── README.md
```

## Конфигурация
- `PORT`: Порт (по умолчанию 8080).
- `STORAGE_TYPE`: `inmemory` или `postgres`.
- `DATABASE_URL`: Строка подключения PostgreSQL.
- `TEST_DATABASE_URL`: Строка подключения для тестов.

## Лицензия
MIT License.