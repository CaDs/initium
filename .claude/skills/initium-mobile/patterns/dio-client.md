# Dio client pattern

The Dio client is built in `providers/api_provider.dart`. A refresh-token
interceptor uses a `Completer<void>` to serialize concurrent 401s — this
prevents spurious logouts when multiple requests race and all get 401.

## Using the API client in a repo

```dart
class OrderRepositoryImpl implements OrderRepository {
  OrderRepositoryImpl(this._api);
  final ApiClient _api;

  @override
  Future<Order> create(CreateOrderInput input) async {
    final response = await _api.dio.post('/api/orders', data: input.toJson());
    final dto = OrderDto.fromJson(response.data);
    return dto.toDomain();
  }
}
```

## Rules

- Never instantiate a new `Dio` instance. Always use the one from
  `apiClientProvider`.
- DTOs are hand-written in `data/remote/dto/`. No `@JsonSerializable`, no
  `freezed` (this template intentionally doesn't pull those deps). If the
  drift check at `make check:openapi` catches a missing field, add it to the
  DTO's `fromJson()`.
- Domain entities are returned from repos, never DTOs. Map via the
  corresponding extension in `data/remote/mapper/`.
- Errors: let Dio `DioException` bubble up into the repo. Wrap with a
  domain error type when the repo translates it for the service layer.
- For new endpoints, register the DTO in `mobile/tool/dto_manifest.yaml` so
  drift detection covers it.

## Testing

- Inject a fake `ApiClient` via `ProviderScope(overrides: [...])` in widget
  tests. See `widget_test.dart` for scope pattern.
- Don't mock Dio directly — it's too low-level. Mock the repository
  interface.
