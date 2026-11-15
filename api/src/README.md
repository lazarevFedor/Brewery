# Декомпозированная спецификация OpenAPI 3.0

Данная директория содержит модульную спецификацию API каталога продукции (`Product Catalog API`), разбитую на логические слои и доменные модули.

## Структура директории

```
api/src/
├── README.md                    # Инструкция по работе с модульной спецификацией
├── openapi.json                 # Главный входной файл спецификации (info, servers, tags, paths, components)
├── paths/                       # Эндпоинты (роуты), сгруппированные по бизнес-модулям
│   ├── auth/                    # Авторизация (/login)
│   ├── classes/                 # Управление классами изделий (/classes, /classes/{id}, ...)
│   │   └── parameters/          # Параметры классов (/classes/parameters, range, apply, exclude)
│   ├── aggregates/              # Агрегаты параметров (/aggregates, apply, ...)
│   ├── products/                # Продукты/изделия (/products, search, by_class, ...)
│   ├── operations/              # CRUD операции техпроцессов (/operations, /operations/{id}, /products/{id}/operations)
│   ├── equipment/               # CRUD операции над оборудованием/приборами (/equipment, /equipment/{id})
│   ├── materials/               # Состав изделия/материалы (/products/{id}/materials, norms, ...)
│   ├── history/                 # История изменений (/products/{id}/history, /history/{id})
│   └── enums/                   # Перечисления классов и значений (/enums, /enums/value)
└── components/                  # Переиспользуемые компоненты
    ├── schemas/                 # DTO / Схемы моделей данных
    │   ├── common/              # Общие схемы (ErrorResponse)
    │   ├── auth/                # Auth
    │   ├── classes/             # Class, ClassCreate, ClassUpdate, RangeParameter, ...
    │   ├── aggregates/          # ParameterAggregate, ParameterAggregateCreate, ...
    │   ├── products/            # Product, ProductCreate, ProductUpdate, ProductShort, ...
    │   ├── operations/          # OperationBase, Operation, OperationCreate, OperationUpdate, ProductTechCard
    │   ├── equipment/           # EquipmentBase, Equipment, EquipmentCreate, EquipmentUpdate, EquipmentShort
    │   ├── professions/         # Profession (справочник должностей/квалификаций)
    │   ├── materials/           # Material, MaterialCreate, MaterialNorm, MaterialTreeNode, ...
    │   ├── history/             # HistoryRecord, HistoryRecordCreate
    │   └── enums/               # EnumClass, EnumValue, ...
    ├── parameters/              # Переиспользуемые параметры запросов (Id, Limit, Offset, ...)
    ├── responses/               # Стандартные HTTP-ответы (400, 403, 404, 409, 422, 500)
    ├── examples/                # Примеры запросов/ответов
    └── securitySchemes/         # Авторизация (bearerAuth JWT)
```

## Как работать со спецификацией

1. **Добавление нового эндпоинта:**
   - Создайте или отредактируйте файл в соответствующей папке `paths/<module>/`.
   - Если эндпоинт новый, добавьте на него ссылку в `paths` файла `api/src/openapi.json`:
     `"/your/endpoint": { "$ref": "./paths/<module>/file.json" }`

2. **Добавление новой модели данных (Schema):**
   - Создайте файл в `components/schemas/<module>/ModelName.json`.
   - Зарегистрируйте его в `components.schemas` файла `api/src/openapi.json`:
     `"ModelName": { "$ref": "./components/schemas/<module>/ModelName.json" }`

3. **Сборка монолитного `api/openapi.json`:**
   ```bash
   make genapi
   ```
   Или напрямую через скрипт:
   ```bash
   python3 scripts/bundle_api.py
   ```

4. **Режим автосборки при изменениях (`--watch`):**
   ```bash
   python3 scripts/bundle_api.py --watch
   ```
