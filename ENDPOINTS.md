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

### Foreman
- `GET /foreman/objects` - список объектов, назначенных текущему подрядчику.
- `GET /foreman/objects/:id` - сведения об объекте, задачи, поставки и прикрепленные материальные документы.
- `POST /foreman/objects/:id/work-reports` - отправляет отчеты о выполненных за день работах.
- `POST /foreman/objects/:id/deliveries` - создает запись о ручной поставке материалов.
- `POST /foreman/objects/:id/deliveries/cv-upload` - загрузка доков, отправка их в CV, сохранение извлеченных данных о поставке, сохранение файла, создание `material_documents`.

### Inspector
- `GET /inspector/dashboard/checks` - список назначенных проверок.
- `GET /inspector/dashboard/objects` - сводка по объектам на панели управления инспектора.
- `GET /inspector/objects` - список объектов, назначенных текущему инспектору.
- `GET /inspector/objects/:id` - сведения об объекте для рабочего процесса инспектора.
- `POST /inspector/objects/:id/activation-decision` - подтверждение или отклонение активации объекта.

## CV Microservice

### Service API
- `GET /health` - CV healthcheck.
- `POST /process-file` - принимает составные файлы и возвращает JSON-данные для распознавания и извлечения информации из одного документа.
