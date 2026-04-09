# Refactor: Test Wiring via Container

## Problem

Controller tests manually construct every use case with mock dependencies:

```go
bulkImportUC := usecases.NewBulkImportObjectsUseCase(mockContainerRepo, mockCollectionRepo, mockAuthService, nil)
bulkImportCollectionUC := usecases.NewBulkImportCollectionUseCase(mockCollectionRepo, mockContainerRepo, mockAuthService, nil, nil)
controller := &ObjectController{
    bulkImportUC:           bulkImportUC,
    bulkImportCollectionUC: bulkImportCollectionUC,
    ...
}
```

Adding a new dependency (e.g. `ImageSearchService`) to a use case constructor forces changes to every test that creates that use case. The image search feature touched **7 test call sites** across 4 files just to add `nil` to the argument list.

## Proposed Fix

Tests should wire through `container.Container` the same way production does:

```go
func setupTestController(t *testing.T) (*ObjectController, *testMocks) {
    ctrl := gomock.NewController(t)
    m := &testMocks{
        containerRepo:  mocks.NewMockContainerRepository(ctrl),
        collectionRepo: mocks.NewMockCollectionRepository(ctrl),
        authService:    mocks.NewMockAuthService(ctrl),
    }
    c := &container.Container{
        ContainerRepo:  m.containerRepo,
        CollectionRepo: m.collectionRepo,
        AuthService:    m.authService,
    }
    // Need a config with defaults for import reserved columns etc.
    c.SetConfig(config.Config{...defaults...})

    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    controller := controllers.NewObjectController(c, logger)
    return controller, m
}
```

## Changes Required

1. **`container.Container`** — Add `SetConfig(*config.Config)` method (or accept a `Config` in a test helper). Currently `config` is private; tests would need access to set import reserved columns, image settings, etc.

2. **`object_controller_test.go`** — Replace 2 `setupObjectController` functions with one that uses `Container`.

3. **`csv_import_test.go`** — Replace 6 manual use case constructions (5x `NewBulkImportCollectionUseCase`, 1x `NewBulkImportObjectsUseCase`) with the container pattern.

4. **Consider for other controllers** — `collection_controller`, `container_controller`, `group_controller` tests likely have the same pattern. Fix them all at once to establish the convention.

## Benefits

- Adding a new service/dependency to a use case only changes `Container` and the constructor — **zero test churn**.
- Tests exercise the same wiring path as production, catching DI bugs.
- Test setup shrinks from ~10 lines of boilerplate per test to 1 helper call.
