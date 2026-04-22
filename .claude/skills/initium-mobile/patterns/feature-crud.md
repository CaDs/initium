# Feature CRUD walkthrough

Step-by-step for adding a new resource to the Flutter app. Example: `Order`
with list + create reachable from Home.

Assume the backend already exposes `POST /api/orders` and `GET /api/orders`
(the spec defines `Order`, `OrderList`, `CreateOrderRequest` schemas — this
is the precondition). See `initium-backend/patterns/feature-crud.md`.

## 1. Domain entity

`mobile/lib/domain/entity/order.dart`:

```dart
class Order {
  final String id;
  final String userId;
  final int totalCents;
  final DateTime createdAt;

  const Order({
    required this.id,
    required this.userId,
    required this.totalCents,
    required this.createdAt,
  });
}
```

Pure Dart. No imports.

## 2. Repository interface

`mobile/lib/domain/repository/order_repository.dart`:

```dart
abstract class OrderRepository {
  Future<(List<Order>?, DomainError?)> list();
  Future<(Order?, DomainError?)> create(int totalCents);
}
```

Positional records. Never throws. Same shape as `UserRepository`.

## 3. DTO + mapper

`mobile/lib/data/remote/dto/order_dto.dart`:

```dart
class OrderDto {
  final String id;
  final String userId;
  final int totalCents;
  final String createdAt; // raw ISO-8601

  OrderDto({
    required this.id,
    required this.userId,
    required this.totalCents,
    required this.createdAt,
  });

  factory OrderDto.fromJson(Map<String, dynamic> json) => OrderDto(
        id: json['id'] as String,
        userId: json['user_id'] as String,
        totalCents: json['total_cents'] as int,
        createdAt: json['created_at'] as String,
      );
}
```

`mobile/lib/data/remote/mapper/order_mapper.dart`:

```dart
import '../../../domain/entity/order.dart';
import '../dto/order_dto.dart';

extension OrderDtoMapper on OrderDto {
  Order toDomain() => Order(
        id: id,
        userId: userId,
        totalCents: totalCents,
        createdAt: DateTime.parse(createdAt), // parse here, not in the entity
      );
}
```

Timestamps parse in the mapper. The domain entity holds `DateTime`, not `String`.

## 4. Manifest registration

`mobile/tool/dto_manifest.yaml` — register **every** wire schema the feature
touches (item + envelope + request body):

```yaml
mappings:
  # existing entries...
  - schema: Order
    dart_file: lib/data/remote/dto/order_dto.dart
    dart_class: OrderDto
  - schema: OrderList
    dart_file: lib/data/remote/dto/order_list_dto.dart
    dart_class: OrderListDto
  - schema: CreateOrderRequest
    dart_file: lib/data/remote/dto/create_order_request_dto.dart
    dart_class: CreateOrderRequestDto
```

Run `make check:openapi` — it should pass only after all three DTOs exist.

## 5. Repository implementation

`mobile/lib/data/repository/order_repository_impl.dart`:

```dart
class OrderRepositoryImpl implements OrderRepository {
  OrderRepositoryImpl(this._client);
  final ApiClient _client;

  @override
  Future<(List<Order>?, DomainError?)> list() async {
    try {
      final response = await _client.dio.get('/api/orders');
      final envelope = response.data as Map<String, dynamic>;
      final items = (envelope['orders'] as List)
          .map((e) => OrderDto.fromJson(e as Map<String, dynamic>).toDomain())
          .toList();
      return (items, null);
    } on DioException catch (e) {
      return (null, _mapError(e));
    }
  }

  @override
  Future<(Order?, DomainError?)> create(int totalCents) async {
    try {
      final response = await _client.dio.post(
        '/api/orders',
        data: {'total_cents': totalCents},
      );
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

Note the envelope unwrap on `list()`. Raw `response.data as List<dynamic>`
would crash.

## 6. Provider

`mobile/lib/providers/orders_provider.dart` — flat under `providers/`, NOT
inside `presentation/orders/`:

```dart
final orderRepositoryProvider = Provider<OrderRepository>((ref) {
  return OrderRepositoryImpl(ref.watch(apiClientProvider));
});

sealed class OrdersState {}
class OrdersLoading extends OrdersState {}
class OrdersLoaded extends OrdersState { OrdersLoaded(this.orders); final List<Order> orders; }
class OrdersError extends OrdersState { OrdersError(this.error); final DomainError error; }

class OrdersNotifier extends StateNotifier<OrdersState> {
  OrdersNotifier(this._repo) : super(OrdersLoading()) { _load(); }
  final OrderRepository _repo;

  Future<void> _load() async {
    final (items, err) = await _repo.list();
    state = err != null ? OrdersError(err) : OrdersLoaded(items!);
  }

  Future<bool> create(int totalCents) async {
    final (order, err) = await _repo.create(totalCents);
    if (err != null || order == null) return false;
    await _load();
    return true;
  }
}

final ordersProvider = StateNotifierProvider<OrdersNotifier, OrdersState>((ref) {
  return OrdersNotifier(ref.watch(orderRepositoryProvider));
});
```

## 7. Screen

`mobile/lib/presentation/orders/orders_screen.dart`:

```dart
class OrdersScreen extends ConsumerWidget {
  const OrdersScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(ordersProvider);
    final l10n = AppLocalizations.of(context)!;
    // ...raw Material widgets: Scaffold + AppBar + body that switches on state
  }
}
```

Use `context.push` from the home-screen button so the back arrow returns to
home. Free-form fields (order amount) get NO `autofillHints`.

## 8. Router

`mobile/lib/presentation/router/app_router.dart` — add the route:

```dart
GoRoute(
  path: '/orders',
  builder: (context, state) => const OrdersScreen(),
),
```

Add to the existing auth-redirect block so unauthenticated users bounce to
`/login`.

## 9. Home button

`mobile/lib/presentation/home/home_screen.dart` — add a button that
`context.push('/orders')`. Localize via `l10n.ordersTitle`.

## 10. i18n

Add keys to ALL THREE ARB files (`app_en.arb`, `app_es.arb`, `app_ja.arb`)
BEFORE referencing them. Run `make gen:mobile` after.

## 11. Verify

```bash
make lint:mobile
make test:mobile
make check:openapi
```

Then commit. `git add -A` to catch every untracked new file (entity,
repository interface, DTOs, mapper, repo impl, provider, screen). A vanilla
`git diff` will miss them.

## 12. Parity

This feature ships on web too (see `initium-web`). Call out the mirror work
in the PR description. See `_shared/parity.md`.
