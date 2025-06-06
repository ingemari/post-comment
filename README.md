# Приложение для постов и комментариев

Это приложение на языке Go, использующее GraphQL для управления постами и комментариями с поддержкой иерархической структуры комментариев и уведомлений в реальном времени через подписки. Приложение по умолчанию использует in-memory хранилище и разработано с учетом легкости, расширяемости и возможности развертывания через Docker.

## Возможности
- Создание и получение постов с заголовками, содержимым и информацией об авторе.
- Добавление и получение комментариев с поддержкой вложенных ответов.
- Возможность включения или отключения комментариев для отдельных постов.
- Пагинация для комментариев и ответов.
- Уведомления в реальном времени о новых комментариях через GraphQL Subscriptions.
- In-memory хранилище для хранения данных.
- Модульные тесты для основного функционала.
- Поддержка Docker для простого развертывания.

## Запуск с Docker
1. **Соберите Docker-образ**:
   ```bash
   docker build -t post-comment-app .
   ```

2. **Запустите Docker-контейнер**:
   ```bash
   docker run -p 8080:8080 post-comment-app
   ```

3. GraphQL API будет доступен по адресу `http://localhost:8080`.

## GraphQL API
Приложение предоставляет GraphQL API со следующими основными операциями:

### Запросы (Queries)
- `posts`: Получение списка всех постов.
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
- `post(id: String!)`: Получение конкретного поста по ID.
  ```graphql
  query {
    post(id: "post-id") {
      id
      title
      content
      author
      comments {
        id
        text
        author
        replies {
          id
          text
          author
        }
      }
    }
  }
  ```

### Мутации (Mutations)
- `createPost(title: String!, content: String!, author: String!, allowComments: Boolean!)`: Создание нового поста.
  ```graphql
  mutation {
    createPost(title: "New Post", content: "This is a post", author: "John Doe", allowComments: true) {
      id
      title
      content
      author
      allowComments
    }
  }
  ```
- `addComment(postId: String!, parentId: String, author: String!, text: String!)`: Добавление комментария к посту или другому комментарию.
  ```graphql
  mutation {
    addComment(postId: "post-id", parentId: "parent-comment-id", author: "Jane Doe", text: "This is a comment") {
      id
      text
      author
      postId
      parentId
    }
  }
  ```

### Подписки (Subscriptions)
- `commentAdded(postId: String!)`: Подписка на уведомления о новых комментариях к конкретному посту.
  ```graphql
  subscription {
    commentAdded(postId: "post-id") {
      id
      text
      author
      postId
      parentId
    }
  }
  ```

## Тестирование
Для запуска модульных тестов выполните:
```bash
go test ./...
```

## Лицензия
MIT License. См. файл `LICENSE` для подробностей.
