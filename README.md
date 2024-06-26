# Документация к использованию приложения
## Запуск приложения
### Запуск приложения локально
Находясь в корне проекта следует выполнить:
```
go run cmd/server.go
```
Перед запуском можно локально создать файл .env и указать там переменные окружения:
1. APP_PORT - по умолчанию 8080.
2. HOST_PORT - по умолчанию 8080.
3. DB_STORE - по умолчанию false. Выбор использования приложения через in-memory или PostgreSQL реализцию хранения данных. Для использования хрфнения в бд необходимо указать true, при любом другом вводе будет false. 
4. DATABASE_URL - для подключения к базе, загружается если DB_STORE выбрано true.

### Используя docker run
Приложение имеет docker-image по [ссылке](https://hub.docker.com/repository/docker/lenev/ozon-task-image/general).
Для запуска необходимо установить изображение:

```docker push lenev/ozon-task-image:<tagname>``` (tagname - latest)

Далее запустить командой:

```docker run -p 8080:8080 -d --rm docker push lenev/ozon-task-image:<tagname>```

Можно указать переменные окружения, используя -e VAR=

Если запустить приложение с DB_STORE=true необходимо указать DATABASE_URL до развёрнутой базы данных PostgreSQL. Если база планируется быть запущенной в докере, то необходимо настроить соединение между контейнерами черех network.
### Используя docker-compose up
Данное приложение содержит настроенный docker-compose.yml файл, а потому можно просто выполнить следующую команду из корневой директории проекта:
```docker-compose up```
После будут созданы образы приложения и postgres с уже настроенным подключением между ними.
## Поддерживаемые запросы в GraphQL
### Mutation:
1. ```createPost(input: CreatePostInput!): Post!``` - создаёт пост с данными, которые необходимы для ввода. Возвращает пост. Необходим уже созданный пользователь.
2. ```createComment(input: CreateCommentInput!): Comment!``` - создаёт комментарий для поста или другого комментария. Возврашает комментарий. Необходимы созданные пользователь и пост.
3. ```createUser(username: String!): User!``` - создаёт пользователя по username. Возвращает пользователя.
### Query:
1. ```getPosts(limit: Int, offset: Int): PostPage!``` - возвращает поле PostPage, которая содержит общее количество постов и список постов, без комментариев под ними. Поддерживает пагинацию.
2. ```getPost(id: ID!, limit: Int, offset: Int): Post!``` - возвращает пост с комментариями по ID поста. Содержит поле CommPage, в котором находятся список комментариев и количество комментариев. У дальнейших ответов уже называется ReplyPage. Поддерживает пагинацию для комментариев и ответов.
3. ```getComments(commId: ID!, limit: Int, offset: Int): Comment!``` - возвращает комментарий, по ID комментария, с ответами на него. Содержит поле ReplyPage, в котором находятся список ответов и количество ответов. Поддерживает пагинацию для ответов.
### Subscription:
1. ```commentAdded(postId: ID!): Comment!``` - позволяет пользователю подписаться на уведомления о создании комментария под постом по ID поста. Выполняется асинхронно без необходимости повторного запроса.
# Особенности работы приложения
## Ограничение вложенности при получении
В данном приложении ограничена максимальная вложенность комментариев до 5 при выполнении getPost и getComments. Сделано с целью если где-то будет слишком большая вложенность.
Если необходимо получить дальнейшие по вложенности комментари, которые будут отвечать на последний, следует выполнить:

```getComments(commId: "<id последнего комментария>") {}```
## Учёт проблемы n+1
В приложении для работы с бд используется пакет gorm, который представляет из себя ORM для Golang. Данная библиотека автоматизирует запросы так, что проблема n+1 не возникает.

Например, если выполнять запрос на getPosts, то у постов содержатся авторы. Проблема n+1 заключалась бы в том, что получив список постов, потом было бы необходимо для каждого поста получить связанного по userID пользователя, написавшего этот пост. То есть запрос на посты + столько запросов к пользователям, сколько было постов.

За счёт использования готовой реализации получения данных из бд, предоставляемой gorm, данная проблема не возникнет.
## Переменные окружения
Присутствуют необходимые для запуска переменные окружения. Через них можно выбрать способ хранения данных, а также указать порты. Переменные имеют значения по умолчанию, за исключением DATABASE_URL. Если переменная DB_STORE=true, то DATABASE_URL необходимо задать в окружении.
При запуске докера можжно указать через -e.
