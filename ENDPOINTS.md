# API Endpoints

## Main Backend

### Аутентификация
- `POST /auth/login` - вход в систему по почте или номеру телефона, возвращает JWT и данные о пользователе.
- `GET /auth/me` - возвращает данные о текущем аутентифицированном пользователе.
- `POST /auth/change-password` - меняет пароль текущего пользователя.

### Admin
- `POST /admin/users` - создает нового пользователя платформы.
- `GET /admin/users` - возвращает список пользователей.

### Customer
- `GET /customer/dashboard/objects` - список объектов на панели управления клиента с фильтрами.
- `GET /customer/dashboard/foremen` - список мастеров на панели управления клиента.
- `GET /customer/foremen-list` - список доступных мастеров.
- `GET /customer/inspectors-list` - список доступных инспекторов.
- `POST /customer/objects` - создает новый объект.
- `POST /customer/objects/:id/activate` - отправляет данные об активации объекта на проверку инспектору.
- `GET /customer/objects/:id` - информация об объекте

*Виды работ*
- `GET /customer/objects/:id/work-items` - Список работ.
- `POST /customer/objects/:id/work-items` - Создание нового вида работы.
- `PUT /customer/objects/:id/work-items/:wid` - Обновление вида работ.
- `DELETE /customer/objects/:id/work-items/:wid` - Удаление вида работ.

*Работа с документами*
- `POST /customer/objects/:id/documents/upload` - Добавление нового документа.
- `GET /customer/objects/:id/documents` - Список документов.
- `DELETE /customer/objects/:id/documents/:docId` - Удаление документов.

### Foreman
- `GET /foreman/objects` - список объектов, назначенных текущему подрядчику.
- `GET /foreman/objects/:id` - сведения об объекте, задачи, поставки и прикрепленные материальные документы.
- `POST /foreman/objects/:id/work-reports` - отправляет отчеты о выполненных за день работах.

*Поставки материала*
- `GET /foreman/objects/:id/deliveries` - Список поставки материалов.
- `POST /foreman/objects/:id/deliveries` - Создает запись о ручной поставке материалов.

*Работа с документами*
- `POST /customer/objects/:id/documents/upload` - Добавление нового документа.
- `GET /customer/objects/:id/documents` - Список документов.
- `DELETE /customer/objects/:id/documents/:docId` - Удаление документов.

### Inspector
- `GET /inspector/dashboard/checks` - список назначенных проверок.
- `GET /inspector/dashboard/objects` - сводка по объектам на панели управления инспектора.
- `GET /inspector/objects` - список объектов, назначенных текущему инспектору.
- `GET /inspector/objects/:id` - сведения об объекте для рабочего процесса инспектора.
- `POST /inspector/objects/:id/activation-decision` - подтверждение или отклонение активации объекта.

- `POST /inspector/objects/:id/documents/upload`
- `GET /inspector/objects/:id/documents`
- `DELETE /inspector/objects/:id/documents/:docId`

*Поставки материала*
- `GET /foreman/objects/:id/deliveries` - Список поставки материалов.
- `POST /foreman/objects/:id/deliveries` - Создает запись о ручной поставке материалов.

*Работа с документами*
- `POST /customer/objects/:id/documents/upload` - Добавление нового документа.
- `GET /customer/objects/:id/documents` - Список документов.
- `DELETE /customer/objects/:id/documents/:docId` - Удаление документов.

## CV Microservice

### Service API
- `GET /health` - CV healthcheck.
- `POST /process-file` - принимает составные файлы и возвращает JSON-данные для распознавания и извлечения информации из одного документа.
