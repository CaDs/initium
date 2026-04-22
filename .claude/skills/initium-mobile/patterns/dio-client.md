# Dio client pattern

The Dio client is built in `providers/api_provider.dart`. A refresh-token
interceptor uses a `Completer<void>` to serialize concurrent 401s — this
prevents spurious logouts when multiple requests race and all get 401.

## Repository skeleton — single-object fetch

Every repo method returns `Future<(T?, DomainError?)>` — positional records,
`null`-on-failure. Do not throw from repos. Every repo has a private
`_mapError(DioException)` helper.

```dart
class OrderRepositoryImpl implements OrderRepository {
  OrderRepositoryImpl(this._api);
  final ApiClient _api;

  @override
  Future<(Order?, DomainError?)> get(String id) async {
    try {
      final response = await _api.dio.get('/api/orders/$id');
      final dto = OrderDto.fromJson(response.data as Map<String, dynamic>);
      return (dto.toDomain(), null);
    } on DioException catch (e) {
      return (null, _mapError(e));
    }
  }

  DomainError _mapError(DioException e) {
    if (e.response?.statusCode == 401) return const Unauthorized();
    return UnknownError(e.message ?? 'Unknown error');
  }
}
```

## Repository skeleton — list fetch (envelope)

Initium list endpoints wrap arrays in an envelope: `GET /api/orders` returns
`{"orders": [...]}`, not a bare array. Unwrap via the envelope field before
mapping:

```dart
@override
Future<(List<Order>?, DomainError?)> list() async {
  try {
    final response = await _api.dio.get('/api/orders');
    final envelope = response.data as Map<String, dynamic>;
    final items = (envelope['orders'] as List)
        .map((e) => OrderDto.fromJson(e as Map<String, dynamic>).toDomain())
        .toList();
    return (items, null);
  } on DioException catch (e) {
    return (null, _mapError(e));
  }
}
```

Treating `response.data` as `List<dynamic>` directly will throw at runtime
because the real payload is the envelope. Also: register both the item schema
(e.g. `Order`) AND the envelope schema (e.g. `OrderList`) AND any request body
(e.g. `CreateOrderRequest`) in `mobile/tool/dto_manifest.yaml` — each needs
its own hand-written DTO + manifest entry.

## Timestamps

OpenAPI declares created/updated timestamps as `format: date-time`. The mapper
parses to `DateTime`; the domain entity stores `DateTime`, not `String`. Do
not leak the wire format into the domain layer.

```dart
// dto
final String createdAt; // raw ISO-8601 string from JSON

// mapper
Order toDomain() => Order(
      id: id,
      createdAt: DateTime.parse(createdAt),
      // ...
    );

// domain entity
class Order {
  final DateTime createdAt;
  // ...
}
```

## Rules

- Never instantiate a new `Dio` instance. Always use `apiClientProvider`.
- DTOs are hand-written JSON maps in `data/remote/dto/`. No `@JsonSerializable`,
  no `freezed`, no `build_runner`. The template intentionally doesn't pull
  those deps.
- Domain entities are returned from repos, never DTOs. Map via the
  corresponding extension in `data/remote/mapper/`.
- Every repo method returns `Future<(T?, DomainError?)>`. Never throw from
  repos; map `DioException` via `_mapError` instead.
- For every new endpoint, register every wire schema in
  `mobile/tool/dto_manifest.yaml` — envelope (`XxxList`), request body
  (`CreateXxxRequest`), item schema. Register only after the OpenAPI schema
  exists, otherwise `make check:openapi` fails immediately.

## Testing

- Inject a fake `ApiClient` via `ProviderScope(overrides: [...])` in widget
  tests. See `widget_test.dart` for scope pattern.
- Don't mock Dio directly — it's too low-level. Mock the repository interface.
